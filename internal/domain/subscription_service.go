package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type SubscriptionService struct {
	repo       ports.SubscriptionRepo
	tariffRepo ports.TariffRepo
	httpClient *http.Client
}

func NewSubscriptionService(repo ports.SubscriptionRepo, tariffRepo ports.TariffRepo) ports.SubscriptionService {
	return &SubscriptionService{
		repo:       repo,
		tariffRepo: tariffRepo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SubscriptionService) Create(ctx context.Context, botID string, telegramID int64, planCode string) (string, error) {
	tariffs, err := s.tariffRepo.ListAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list tariffs: %w", err)
	}

	var plan *ports.TariffPlan
	for _, t := range tariffs {
		if t.Code == planCode {
			plan = t
			break
		}
	}
	if plan == nil {
		return "", fmt.Errorf("unknown plan code: %s", planCode)
	}

	// бесплатный план
	if plan.Price == 0 {
		expiresAt := time.Now().Add(time.Duration(plan.PeriodDays) * 24 * time.Hour)
		sub := &ports.Subscription{
			BotID:             botID,
			TelegramID:        telegramID,
			PlanID:            plan.ID,
			Status:            "active",
			StartedAt:         time.Now(),
			ExpiresAt:         &expiresAt,
			YookassaPaymentID: nil,
		}
		if err := s.repo.Create(ctx, sub); err != nil {
			return "", fmt.Errorf("failed to activate free plan: %w", err)
		}
		return "", nil
	}

	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secretKey := os.Getenv("YOOKASSA_SECRET_KEY")
	botTokens := os.Getenv("BOT_TOKENS")

	if shopID == "" || secretKey == "" {
		return "", fmt.Errorf("yookassa env variables missing: shopID/secretKey required")
	}

	// нормализуем apiURL: если задан базовый url, дополним до /v3/payments
	if apiURL == "" {
		return "", fmt.Errorf("yookassa env variable YOOKASSA_API_URL missing")
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	// username бота по botID
	var botUsername string
	for _, p := range strings.Split(botTokens, ",") {
		parts := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(parts) == 2 && parts[0] == botID {
			botUsername = strings.Split(parts[1], ":")[0]
			break
		}
	}
	if botUsername == "" {
		return "", fmt.Errorf("bot username not found for botID=%s", botID)
	}

	returnURL := fmt.Sprintf("https://t.me/%s?start=paid_%s", botUsername, botID)

	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", plan.Price),
			"currency": "RUB",
		},
		"capture":     true,
		"description": fmt.Sprintf("Subscription %s for user %d", plan.Code, telegramID),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"receipt": map[string]any{
			"customer": map[string]any{
				"email": fmt.Sprintf("user_%d@example.com", telegramID),
			},
			"items": []map[string]any{
				{
					"description":     fmt.Sprintf("Подписка %s", plan.Code),
					"quantity":        "1.00",
					"amount":          map[string]any{"value": fmt.Sprintf("%.2f", plan.Price), "currency": "RUB"},
					"vat_code":        1,
					"payment_subject": "service",
					"payment_mode":    "full_prepayment",
				},
			},
		},
		"metadata": map[string]any{
			"bot_id":      botID,
			"telegram_id": fmt.Sprintf("%d", telegramID),
		},
	}

	// сериализуем тело и логируем
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request body: %w", err)
	}
	log.Printf("[yookassa][request] url=%s shop=%s plan=%s telegram_id=%d body=%s", apiURL, shopID, plan.Code, telegramID, string(reqBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	// Basic auth (shopID:secret)
	req.SetBasicAuth(shopID, secretKey)

	// Idempotence key — важно
	idemp := fmt.Sprintf("%d", time.Now().UnixNano())
	req.Header.Set("Idempotence-Key", idemp)
	req.Header.Set("Content-Type", "application/json")

	// логируем заголовки (без секретов)
	log.Printf("[yookassa][request.headers] Idempotence-Key=%s Content-Type=%s BasicAuthUser=%s", idemp, req.Header.Get("Content-Type"), shopID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[yookassa][response] status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// возвращаем тело в ошибке — сразу видно причину валидэйшна
		return "", fmt.Errorf("yookassa returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var yresp struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(respBody, &yresp); err != nil {
		return "", fmt.Errorf("decode yookassa response: %w; raw=%s", err, string(respBody))
	}

	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		return "", fmt.Errorf("invalid yookassa response: missing id or url; raw=%s", string(respBody))
	}

	sub := &ports.Subscription{
		BotID:             botID,
		TelegramID:        telegramID,
		PlanID:            plan.ID,
		Status:            "pending",
		StartedAt:         time.Now(),
		YookassaPaymentID: &yresp.ID,
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return "", fmt.Errorf("failed to create subscription: %w", err)
	}

	return yresp.Confirmation.URL, nil
}

func (s *SubscriptionService) Activate(ctx context.Context, paymentID string) error {
	sub, err := s.repo.GetByPaymentID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to find subscription by paymentID: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("subscription not found for paymentID: %s", paymentID)
	}
	return s.repo.UpdateStatus(ctx, sub.ID, "active")
}

func (s *SubscriptionService) GetStatus(ctx context.Context, botID string, telegramID int64) (string, error) {
	sub, err := s.repo.Get(ctx, botID, telegramID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "none", nil
	}
	return sub.Status, nil
}

func (s *SubscriptionService) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	return s.repo.ListAll(ctx)
}
