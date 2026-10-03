package cards

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/repa-app/repa/internal/cardorder"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/repa-app/repa/internal/lib"
	"github.com/rs/zerolog/log"
)

// Service generates reputation card images and stores them in S3.
type Service struct {
	queries db.Querier
	s3      *lib.S3Client
}

func NewService(queries db.Querier, s3 *lib.S3Client) *Service {
	return &Service{queries: queries, s3: s3}
}

// GenerateCardsForSeason generates card images for all members of a season.
func (s *Service) GenerateCardsForSeason(ctx context.Context, seasonID string) error {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return fmt.Errorf("get season: %w", err)
	}

	group, err := s.queries.GetGroupByID(ctx, season.GroupID)
	if err != nil {
		return fmt.Errorf("get group: %w", err)
	}

	members, err := s.queries.GetGroupMembers(ctx, season.GroupID)
	if err != nil {
		return fmt.Errorf("get members: %w", err)
	}

	// Create a single browser context for all cards in this season
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	for _, m := range members {
		if err := s.generateCardForUser(browserCtx, season, group, m); err != nil {
			log.Error().Err(err).
				Str("user_id", m.ID).
				Str("season_id", seasonID).
				Msg("failed to generate card for user, skipping")
			continue
		}
	}

	return nil
}

// cardTopAttributes picks what the rendered card shows: the leading three attributes and the title.
//
// Split out from generateCardForUser so it can be tested without a browser — the ordering rule is the
// part worth testing, and it is the part that would otherwise be verifiable only by looking at a PNG.
//
// Ordered the same way the API orders a card, because this *is* the card: the PNG is the artifact a
// person is invited to share, so if any path must lead with something worth sharing it is this one.
// Taking the results in raw percentage order here would have let the app and the shared image
// disagree about which attribute comes first.
func cardTopAttributes(results []db.GetSeasonResultsByUserRow) ([]CardAttribute, string) {
	ordered := cardorder.ByTone(results)

	topAttrs := make([]CardAttribute, 0, 3)
	for i, r := range ordered {
		if i >= 3 {
			break
		}
		topAttrs = append(topAttrs, CardAttribute{
			QuestionText: r.QuestionText,
			Percentage:   r.Percentage,
		})
	}

	// The title describes the attribute the reader sees first, not the highest percentage.
	title := "Загадка века"
	if len(ordered) > 0 {
		title = titleForCategory(string(ordered[0].QuestionCategory))
	}
	return topAttrs, title
}

func (s *Service) generateCardForUser(browserCtx context.Context, season db.Season, group db.Group, member db.GetGroupMembersRow) error {
	results, err := s.queries.GetSeasonResultsByUser(browserCtx, db.GetSeasonResultsByUserParams{
		SeasonID: season.ID,
		TargetID: member.ID,
	})
	if err != nil {
		return fmt.Errorf("get results: %w", err)
	}

	topAttrs, title := cardTopAttributes(results)

	avatarEmoji := "\U0001F346"
	if member.AvatarEmoji.Valid && member.AvatarEmoji.String != "" {
		avatarEmoji = member.AvatarEmoji.String
	}

	// The card doubles as the group's invitation, so it carries the live invite code and a
	// QR of its link. A QR failure degrades to the printed code rather than losing the card.
	// The card is generated per member, so its link can carry *who* shared it — a group's
	// invite code identifies the group, not the inviter, and a referral reward has to reach the
	// person who actually invited rather than the group's admin.
	inviteURL := ""
	if group.InviteCode != "" {
		inviteURL = fmt.Sprintf("https://repa.app/join/%s?s=card&ref=%s", group.InviteCode, member.ID)
	}

	data := CardData{
		Username:        member.Username,
		AvatarEmoji:     avatarEmoji,
		TopAttributes:   topAttrs,
		ReputationTitle: title,
		GroupName:       group.Name,
		SeasonNumber:    int(season.Number),
		InviteCode:      group.InviteCode,
		InviteQRDataURI: InviteQRDataURI(inviteURL),
	}

	html := BuildCardHTML(data)

	// Render to PNG using chromedp
	png, err := renderHTMLToPNG(browserCtx, html)
	if err != nil {
		return fmt.Errorf("render card: %w", err)
	}

	// Upload to S3
	key := fmt.Sprintf("cards/%s/%s.png", season.ID, member.ID)
	imageURL, err := s.s3.Upload(browserCtx, key, png, "image/png")
	if err != nil {
		return fmt.Errorf("upload card: %w", err)
	}

	// Save to card_cache
	_, err = s.queries.UpsertCardCache(browserCtx, db.UpsertCardCacheParams{
		ID:       uuid.New().String(),
		UserID:   member.ID,
		SeasonID: season.ID,
		ImageUrl: imageURL,
	})
	if err != nil {
		return fmt.Errorf("upsert card cache: %w", err)
	}

	log.Info().
		Str("user_id", member.ID).
		Str("season_id", season.ID).
		Msg("card generated")

	return nil
}

func renderHTMLToPNG(browserCtx context.Context, html string) ([]byte, error) {
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	var buf []byte
	err := chromedp.Run(tabCtx,
		chromedp.EmulateViewport(1080, 1920),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.FullScreenshot(&buf, 100),
	)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func titleForCategory(cat string) string {
	switch cat {
	case "HOT":
		return "Горячая штучка"
	case "FUNNY":
		return "Душа компании"
	case "SECRETS":
		return "Хранитель тайн"
	case "SKILLS":
		return "Мастер на все руки"
	case "ROMANCE":
		return "Сердцеед"
	case "STUDY":
		return "Ботан года"
	default:
		return "Загадка века"
	}
}

// GetCardURL returns the card image URL for a user in a season, if available.
func (s *Service) GetCardURL(ctx context.Context, seasonID, userID string) (string, error) {
	cache, err := s.queries.GetCardCache(ctx, db.GetCardCacheParams{
		UserID:   userID,
		SeasonID: seasonID,
	})
	if err != nil {
		return "", err
	}
	return cache.ImageUrl, nil
}
