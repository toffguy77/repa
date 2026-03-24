package crystals

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

var (
	ErrPackageNotFound    = errors.New("package not found")
	ErrInsufficientFunds  = errors.New("insufficient crystal balance")
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrDuplicatePayment   = errors.New("payment already processed")
)

type CrystalPackage struct {
	ID           string `json:"id"`
	Crystals     int    `json:"crystals"`
	Bonus        int    `json:"bonus"`
	PriceKopecks int    `json:"price_kopecks"`
}

var Packages = []CrystalPackage{
	{ID: "starter", Crystals: 10, Bonus: 0, PriceKopecks: 5900},
	{ID: "popular", Crystals: 30, Bonus: 5, PriceKopecks: 14900},
	{ID: "advanced", Crystals: 70, Bonus: 15, PriceKopecks: 29900},
	{ID: "max", Crystals: 160, Bonus: 40, PriceKopecks: 59900},
}

type paymentInfo struct {
	UserID    string `json:"user_id"`
	PackageID string `json:"package_id"`
}

const paymentKeyPrefix = "payment:"
const paymentTTL = 1 * time.Hour

type Service struct {
	queries  db.Querier
	sqlDB    *sql.DB
	rdb      *redis.Client
	yukassa  *lib.YukassaClient
}

func NewService(queries db.Querier, sqlDB *sql.DB, rdb *redis.Client, yukassa *lib.YukassaClient) *Service {
	return &Service{
		queries: queries,
		sqlDB:   sqlDB,
		rdb:     rdb,
		yukassa: yukassa,
	}
}

func (s *Service) GetBalance(ctx context.Context, userID string) (int32, error) {
	return s.queries.GetUserBalance(ctx, userID)
}

func (s *Service) GetPackages() []CrystalPackage {
	return Packages
}

func findPackage(id string) *CrystalPackage {
	for _, p := range Packages {
		if p.ID == id {
			return &p
		}
	}
	return nil
}

type InitPurchaseResult struct {
	PaymentURL string `json:"payment_url"`
	PaymentID  string `json:"payment_id"`
}

var ErrPaymentsUnavailable = errors.New("payments are not configured")

func (s *Service) InitPurchase(ctx context.Context, userID, packageID string) (*InitPurchaseResult, error) {
	if s.yukassa == nil {
		return nil, ErrPaymentsUnavailable
	}

	pkg := findPackage(packageID)
	if pkg == nil {
		return nil, ErrPackageNotFound
	}

	description := fmt.Sprintf("Repa: %d crystals", pkg.Crystals+pkg.Bonus)

	payment, err := s.yukassa.CreatePayment(ctx, pkg.PriceKopecks, description)
	if err != nil {
		return nil, fmt.Errorf("create yukassa payment: %w", err)
	}

	info := paymentInfo{
		UserID:    userID,
		PackageID: packageID,
	}
	data, _ := json.Marshal(info)
	if err := s.rdb.Set(ctx, paymentKeyPrefix+payment.ID, data, paymentTTL).Err(); err != nil {
		return nil, fmt.Errorf("store payment info in redis: %w", err)
	}

	confirmationURL := ""
	if payment.Confirmation != nil {
		confirmationURL = payment.Confirmation.ConfirmationURL
	}

	return &InitPurchaseResult{
		PaymentURL: confirmationURL,
		PaymentID:  payment.ID,
	}, nil
}

type WebhookEvent struct {
	Type   string          `json:"type"`
	Event  string          `json:"event"`
	Object json.RawMessage `json:"object"`
}

func (s *Service) ProcessWebhook(ctx context.Context, event WebhookEvent) error {
	if event.Event != "payment.succeeded" {
		return nil
	}

	var payment lib.YukassaPayment
	if err := json.Unmarshal(event.Object, &payment); err != nil {
		return fmt.Errorf("unmarshal payment object: %w", err)
	}

	data, err := s.rdb.Get(ctx, paymentKeyPrefix+payment.ID).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Warn().Str("payment_id", payment.ID).Msg("payment info not found in redis, may have expired")
			return ErrPaymentNotFound
		}
		return fmt.Errorf("get payment info from redis: %w", err)
	}

	var info paymentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("unmarshal payment info: %w", err)
	}

	pkg := findPackage(info.PackageID)
	if pkg == nil {
		return ErrPackageNotFound
	}

	totalCrystals := int32(pkg.Crystals + pkg.Bonus)

	if err := s.creditCrystals(ctx, info.UserID, totalCrystals, payment.ID); err != nil {
		if errors.Is(err, ErrDuplicatePayment) {
			return nil // idempotent
		}
		return err
	}

	// Clean up redis key
	s.rdb.Del(ctx, paymentKeyPrefix+payment.ID)

	log.Info().
		Str("user_id", info.UserID).
		Str("payment_id", payment.ID).
		Int32("crystals", totalCrystals).
		Msg("crystals credited via webhook")

	return nil
}

func (s *Service) creditCrystals(ctx context.Context, userID string, amount int32, paymentID string) error {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := db.New(tx)

	balance, err := q.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}

	newBalance := balance + amount
	_, err = q.CreateCrystalLog(ctx, db.CreateCrystalLogParams{
		ID:          uuid.New().String(),
		UserID:      userID,
		Delta:       amount,
		Balance:     newBalance,
		Type:        db.CrystalLogTypePURCHASE,
		Description: sql.NullString{String: fmt.Sprintf("Purchase: %d crystals", amount), Valid: true},
		ExternalID:  sql.NullString{String: paymentID, Valid: true},
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicatePayment
		}
		return fmt.Errorf("create crystal log: %w", err)
	}

	return tx.Commit()
}

type VerifyResult struct {
	Status     string `json:"status"`
	NewBalance *int32 `json:"new_balance,omitempty"`
}

func (s *Service) VerifyPurchase(ctx context.Context, userID, paymentID string) (*VerifyResult, error) {
	if s.yukassa == nil {
		return nil, ErrPaymentsUnavailable
	}

	// Check redis to confirm this payment belongs to this user
	data, err := s.rdb.Get(ctx, paymentKeyPrefix+paymentID).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("get payment info from redis: %w", err)
	}

	if err == nil {
		// Redis key exists — verify ownership
		var info paymentInfo
		if json.Unmarshal(data, &info) == nil && info.UserID != userID {
			return nil, ErrPaymentNotFound
		}
	} else {
		// Redis key expired — check crystal_logs for this payment to verify ownership
		logs, logErr := s.queries.GetUserCrystalLogs(ctx, db.GetUserCrystalLogsParams{
			UserID: userID,
			Limit:  100,
			Offset: 0,
		})
		if logErr != nil {
			return nil, ErrPaymentNotFound
		}
		found := false
		for _, l := range logs {
			if l.ExternalID.Valid && l.ExternalID.String == paymentID {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrPaymentNotFound
		}
	}

	payment, err := s.yukassa.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get yukassa payment: %w", err)
	}

	result := &VerifyResult{
		Status: payment.Status,
	}

	if payment.Status == "succeeded" {
		balance, err := s.queries.GetUserBalance(ctx, userID)
		if err == nil {
			result.NewBalance = &balance
		}
	}

	return result, nil
}

func isDuplicateKeyError(err error) bool {
	// PostgreSQL unique_violation error code 23505
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key")
}
