package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AnthropicClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type ModerationResult struct {
	Approved bool    `json:"approved"`
	Reason   *string `json:"reason"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

const moderationSystem = `Ты модератор вопросов для мобильного приложения «Репа».
Приложение — социальная игра для школьников и студентов 14-22 лет.
Правила допустимых вопросов:
- Вопрос должен быть юмористическим или наблюдательным
- Нельзя: оскорбления, мат, расизм, сексизм, буллинг конкретных людей
- Нельзя: сексуальный контент (даже намёки для аудитории до 18)
- Нельзя: призывы к насилию или самоповреждению
- Можно: безобидный юмор, наблюдения о поведении, лёгкая провокация без агрессии
Отвечай ТОЛЬКО валидным JSON: { "approved": boolean, "reason": string | null }
Reason — короткое объяснение только при отклонении (на русском).`

func (c *AnthropicClient) ModerateQuestion(ctx context.Context, text string) (*ModerationResult, error) {
	reqBody := anthropicRequest{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: 100,
		System:    moderationSystem,
		Messages: []anthropicMessage{
			{Role: "user", Content: fmt.Sprintf("Вопрос: «%s»", text)},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Content) == 0 || apiResp.Content[0].Type != "text" {
		return nil, fmt.Errorf("unexpected response format")
	}

	var result ModerationResult
	if err := json.Unmarshal([]byte(apiResp.Content[0].Text), &result); err != nil {
		return nil, fmt.Errorf("parse moderation result: %w", err)
	}

	return &result, nil
}
