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

	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type SubscriptionService struct {
	repo       ports.SubscriptionRepo
	tariffRepo ports.TariffRepo
	httpClient *http.Client
	notifier   error_notificator.Notificator
}

func NewSubscriptionService(
	repo ports.SubscriptionRepo,
	tariffRepo ports.TariffRepo,
	notifier error_notificator.Notificator,
) ports.SubscriptionService {

	return &SubscriptionService{
		repo:       repo,
		tariffRepo: tariffRepo,
		notifier:   notifier,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ==================================================
// CREATE
// ==================================================
func (s *SubscriptionService) Create(ctx context.Context, botID string, telegramID int64, planCode string) (string, error) {

	// 1. Находим тариф
	tariffs, err := s.tariffRepo.ListAll(ctx)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка чтения тарифов при создании подписки")
		return "", fmt.Errorf("list tariffs: %w", err)
	}

	var plan *ports.TariffPlan
	for _, t := range tariffs {
		if t.Code == planCode {
			plan = t
			break
		}
	}
	if plan == nil {
		err := fmt.Errorf("unknown plan code: %s", planCode)
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Пользователь выбрал несуществующий тариф (%s)", planCode))
		return "", err
	}

	// 2. ENVs
	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secret := os.Getenv("YOOKASSA_SECRET_KEY")

	if apiURL == "" || shopID == "" || secret == "" {
		err := fmt.Errorf("missing Yookassa ENV variables")
		s.notifier.Notify(ctx, botID, err,
			"Отсутствуют переменные окружения для YooKassa (YOOKASSA_API_URL / SHOP_ID / SECRET)")
		return "", err
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	// 3. Формируем запрос
	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", plan.Price),
			"currency": "RUB",
		},
		"capture":     true,
		"description": fmt.Sprintf("Subscription %s for user %d", plan.Code, telegramID),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": "https://aifulls.com/success.html",
		},
		"metadata": map[string]any{
			"bot_id":      botID,
			"telegram_id": fmt.Sprintf("%d", telegramID),
		},
	}

	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			"Ошибка формирования запроса в YooKassa при создании подписки")
		return "", fmt.Errorf("build request: %w", err)
	}

	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	// 4. Вызов YooKassa
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			"Ошибка сети при обращении к YooKassa (подписка не создана)")
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("yookassa returned %d: %s", resp.StatusCode, string(raw))
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("YooKassa отклонила запрос (код %d)", resp.StatusCode))
		return "", err
	}

	// 5. Декодируем
	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(raw, &yresp); err != nil {
		s.notifier.Notify(ctx, botID, err,
			"YooKassa вернула невалидный JSON")
		return "", fmt.Errorf("decode yookassa: %w", err)
	}

	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		err := fmt.Errorf("invalid yookassa response: %s", string(raw))
		s.notifier.Notify(ctx, botID, err,
			"От YooKassa пришёл пустой ответ без URL оплаты")
		return "", err
	}

	// 6. Сохраняем подписку
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
		s.notifier.Notify(ctx, botID, err,
			"Не удалось создать запись подписки в БД")
		return "", fmt.Errorf("create subscription: %w", err)
	}

	return yresp.Confirmation.URL, nil
}

func (s *SubscriptionService) Get(ctx context.Context, botID string, telegramID int64) (*ports.Subscription, error) {
	sub, err := s.repo.Get(ctx, botID, telegramID)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка загрузки подписки (Get)")
		return nil, err
	}
	return sub, nil
}

// ==================================================
// ACTIVATE
// ==================================================
func (s *SubscriptionService) Activate(ctx context.Context, paymentID string) error {

	sub, err := s.repo.GetByPaymentID(ctx, paymentID)
	if err != nil {
		s.notifier.Notify(ctx, "unknown", err,
			"Ошибка загрузки подписки по paymentID в Activate()")
		return fmt.Errorf("load subscription: %w", err)
	}
	if sub == nil {
		err := fmt.Errorf("subscription not found for paymentID=%s", paymentID)
		s.notifier.Notify(ctx, "unknown", err,
			fmt.Sprintf("Webhook YooKassa, но подписка не найдена (%s)", paymentID))
		return err
	}

	plan, err := s.tariffRepo.GetByID(ctx, int(sub.PlanID))
	if err != nil {
		s.notifier.Notify(ctx, sub.BotID, err,
			"Ошибка загрузки тарифного плана при активации")
		return fmt.Errorf("load plan: %w", err)
	}
	if plan == nil {
		err := fmt.Errorf("plan not found id=%d", sub.PlanID)
		s.notifier.Notify(ctx, sub.BotID, err,
			"Webhook YooKassa: тариф не найден!")
		return err
	}

	start := time.Now()
	exp := start.Add(time.Duration(plan.DurationMinutes) * time.Minute)

	if err := s.repo.Activate(ctx, sub.ID, start, exp, plan.VoiceMinutes); err != nil {
		s.notifier.Notify(ctx, sub.BotID, err,
			"Не удалось активировать подписку в БД")
		return fmt.Errorf("activate: %w", err)
	}

	return nil
}

// ==================================================
// STATUS
// ==================================================
func (s *SubscriptionService) GetStatus(ctx context.Context, botID string, telegramID int64) (string, error) {
	sub, err := s.repo.Get(ctx, botID, telegramID)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка получения статуса подписки из БД")
		return "", err
	}
	if sub == nil {
		return "none", nil
	}
	if sub.ExpiresAt != nil && time.Now().After(*sub.ExpiresAt) {
		_ = s.repo.UpdateStatus(ctx, sub.ID, "expired")
		return "expired", nil
	}
	return sub.Status, nil
}

// ==================================================
func (s *SubscriptionService) ListAll(ctx context.Context) ([]*ports.Subscription, error) {
	list, err := s.repo.ListAll(ctx)
	if err != nil {
		s.notifier.Notify(ctx, "global", err, "Ошибка чтения списка всех подписок")
	}
	return list, err
}

func (s *SubscriptionService) UseVoiceMinutes(ctx context.Context, botID string, telegramID int64, used float64) (bool, error) {
	ok, err := s.repo.UseVoiceMinutes(ctx, botID, telegramID, used)
	if err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка списания голосовых минут (%.2f)", used))
	}
	return ok, err
}
