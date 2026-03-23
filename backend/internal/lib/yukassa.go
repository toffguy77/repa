package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const yukassaBaseURL = "https://api.yookassa.ru/v3"

type YukassaClient struct {
	shopID    string
	secretKey string
	returnURL string
	client    *http.Client
}

func NewYukassaClient(shopID, secretKey, returnURL string) *YukassaClient {
	return &YukassaClient{
		shopID:    shopID,
		secretKey: secretKey,
		returnURL: returnURL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type YukassaAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type YukassaConfirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

type YukassaPayment struct {
	ID           string               `json:"id"`
	Status       string               `json:"status"`
	Amount       YukassaAmount        `json:"amount"`
	Confirmation *YukassaConfirmation `json:"confirmation,omitempty"`
	Description  string               `json:"description,omitempty"`
	CreatedAt    string               `json:"created_at,omitempty"`
}

type createPaymentRequest struct {
	Amount       YukassaAmount       `json:"amount"`
	Confirmation YukassaConfirmation `json:"confirmation"`
	Description  string              `json:"description"`
	Capture      bool                `json:"capture"`
}

func (c *YukassaClient) CreatePayment(ctx context.Context, amountKopecks int, description string) (*YukassaPayment, error) {
	rubles := fmt.Sprintf("%d.%02d", amountKopecks/100, amountKopecks%100)

	body := createPaymentRequest{
		Amount: YukassaAmount{
			Value:    rubles,
			Currency: "RUB",
		},
		Confirmation: YukassaConfirmation{
			Type:      "redirect",
			ReturnURL: c.returnURL,
		},
		Description: description,
		Capture:     true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yukassaBaseURL+"/payments", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", uuid.New().String())
	req.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yukassa error %d: %s", resp.StatusCode, string(respBody))
	}

	var payment YukassaPayment
	if err := json.Unmarshal(respBody, &payment); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &payment, nil
}

func (c *YukassaClient) GetPayment(ctx context.Context, paymentID string) (*YukassaPayment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, yukassaBaseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(c.shopID, c.secretKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("yukassa error %d: %s", resp.StatusCode, string(respBody))
	}

	var payment YukassaPayment
	if err := json.Unmarshal(respBody, &payment); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &payment, nil
}
