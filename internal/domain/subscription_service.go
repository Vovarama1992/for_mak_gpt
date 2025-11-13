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

// ------------------------------------------------
// CREATE
// ------------------------------------------------
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

	// юкасса env
	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secret := os.Getenv("YOOKASSA_SECRET_KEY")

	if apiURL == "" || shopID == "" || secret == "" {
		return "", fmt.Errorf("missing yookassa env variables")
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	// return URL
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

	reqBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("yookassa returned %d: %s", resp.StatusCode, string(raw))
	}

	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(raw, &yresp); err != nil {
		return "", fmt.Errorf("decode yookassa: %w", err)
	}
	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		return "", fmt.Errorf("invalid yookassa response: %s", string(raw))
	}

	// создаём pending подписку
	now := time.Now()

	sub := &ports.Subscription{
		BotID:             botID,
		TelegramID:        telegramID,
		PlanID:            int64(plan.ID),
		Status:            "pending",
		StartedAt:         &now,
		ExpiresAt:         nil,
		YookassaPaymentID: &yresp.ID,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return "", fmt.Errorf("failed to create subscription: %w", err)
	}

	return yresp.Confirmation.URL, nil
}

// ------------------------------------------------
// ACTIVATE
// ------------------------------------------------
func (s *SubscriptionService) Activate(ctx context.Context, paymentID string) error {

	sub, err := s.repo.GetByPaymentID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("load subscription: %w", err)
	}
	if sub == nil {
		return fmt.Errorf("subscription not found for paymentID=%s", paymentID)
	}

	plan, err := s.tariffRepo.GetByID(ctx, int(sub.PlanID))
	if err != nil {
		return fmt.Errorf("load plan: %w", err)
	}
	if plan == nil {
		return fmt.Errorf("plan not found id=%d", sub.PlanID)
	}

	start := time.Now()
	exp := start.Add(time.Duration(plan.DurationMinutes) * time.Minute)

	return s.repo.Activate(ctx, sub.ID, start, exp)
}

// ------------------------------------------------
// STATUS
// ------------------------------------------------
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
