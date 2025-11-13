package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	// ищем тариф
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

	// =====================================================================
	// ВСЕ тарифы платные — создаём платёж Юкассы
	// =====================================================================

	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secretKey := os.Getenv("YOOKASSA_SECRET_KEY")

	if apiURL == "" || shopID == "" || secretKey == "" {
		return "", fmt.Errorf("missing yookassa env variables")
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	// 🎯 ПРАВИЛЬНЫЙ RETURN_URL
	// сюда вернётся браузер после оплаты
	returnURL := "https://aifulls.com/success.html"

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

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.SetBasicAuth(shopID, secretKey)
	idemp := fmt.Sprintf("%d", time.Now().UnixNano())
	req.Header.Set("Idempotence-Key", idemp)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
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
		return "", fmt.Errorf("invalid yookassa response: %s", string(respBody))
	}

	// подписка пока pending
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

	plan, err := s.tariffRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("failed to load tariff plan: %w", err)
	}
	if plan == nil {
		return fmt.Errorf("tariff plan not found for plan_id=%d", sub.PlanID)
	}

	startedAt := time.Now()
	expiresAt := startedAt.Add(time.Duration(plan.DurationMinutes) * time.Minute)

	return s.repo.Activate(ctx, sub.ID, startedAt, expiresAt)
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
