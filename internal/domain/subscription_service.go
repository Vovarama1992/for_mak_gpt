package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	var planID int
	var price float64
	for _, t := range tariffs {
		if t.Code == planCode {
			planID = t.ID
			price = t.Price
			break
		}
	}
	if planID == 0 {
		return "", fmt.Errorf("unknown plan code: %s", planCode)
	}

	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secretKey := os.Getenv("YOOKASSA_SECRET_KEY")
	botTokens := os.Getenv("BOT_TOKENS")

	if apiURL == "" || shopID == "" || secretKey == "" {
		return "", fmt.Errorf("yookassa env variables missing")
	}

	// достаём username бота по botID
	var botUsername string
	pairs := strings.Split(botTokens, ",")
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		if name == botID {
			// Telegram username — это всё до двоеточия
			token := parts[1]
			usernamePart := strings.Split(token, ":")[0]
			botUsername = usernamePart
			break
		}
	}
	if botUsername == "" {
		return "", fmt.Errorf("bot username not found for botID=%s", botID)
	}

	// return_url под конкретного бота
	returnURL := fmt.Sprintf("https://t.me/%s?start=paid_%s", botUsername, botID)

	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", price),
			"currency": "RUB",
		},
		"capture":     true,
		"description": fmt.Sprintf("Subscription %s for user %d", planCode, telegramID),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"metadata": map[string]any{
			"bot_id":      botID,
			"telegram_id": fmt.Sprintf("%d", telegramID),
		},
	}

	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req.SetBasicAuth(shopID, secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("yookassa returned status %d", resp.StatusCode)
	}

	var yresp struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&yresp); err != nil {
		return "", fmt.Errorf("decode yookassa response: %w", err)
	}

	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		return "", fmt.Errorf("invalid yookassa response: missing id or url")
	}

	sub := &ports.Subscription{
		BotID:             botID,
		TelegramID:        telegramID,
		PlanID:            planID,
		Status:            "pending",
		StartedAt:         time.Now(),
		YookassaPaymentID: yresp.ID,
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
