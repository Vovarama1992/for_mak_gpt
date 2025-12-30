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

	"github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	"github.com/Vovarama1992/make_ziper/internal/notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/trial"
)

type SubscriptionService struct {
	repo       ports.SubscriptionRepo
	tariffRepo ports.TariffRepo
	trialRepo  trial.RepoInf

	httpClient *http.Client
	notifier   notificator.Notificator
	minuteSvc  minutes_packages.MinutePackageService
}

func NewSubscriptionService(
	repo ports.SubscriptionRepo,
	tariffRepo ports.TariffRepo,
	trialRepo trial.RepoInf,
	minuteSvc minutes_packages.MinutePackageService,
	notifier notificator.Notificator,
) ports.SubscriptionService {
	return &SubscriptionService{
		repo:       repo,
		tariffRepo: tariffRepo,
		trialRepo:  trialRepo,
		minuteSvc:  minuteSvc,
		notifier:   notifier,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ==================================================
// CREATE
// ==================================================
func (s *SubscriptionService) Create(
	ctx context.Context,
	botID string,
	telegramID int64,
	planCode string,
) (string, error) {

	// 1. Ищем тариф
	tariffs, err := s.tariffRepo.ListAll(ctx)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка чтения тарифов (подписка)")
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
	customerPhone := os.Getenv("YOOKASSA_CUSTOMER_PHONE")

	if apiURL == "" || shopID == "" || secret == "" {
		err := fmt.Errorf("missing yookassa ENV variables")
		s.notifier.Notify(ctx, botID, err,
			"Переменные окружения YooKassa отсутствуют")
		return "", err
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	if customerPhone == "" {
		customerPhone = "79000000000"
	}

	// 3. Формируем тело запроса — ИСПРАВЛЕНО payment_mode
	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", plan.Price),
			"currency": "RUB",
		},
		"capture":     true,
		"description": fmt.Sprintf("Subscription '%s' (user %d)", plan.Code, telegramID),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": "https://aifulls.com/success.html",
		},
		"receipt": map[string]any{
			"customer": map[string]any{
				"phone": customerPhone,
			},
			"items": []map[string]any{
				{
					"description":     fmt.Sprintf("Subscription %s", plan.Code),
					"quantity":        "1.00",
					"amount":          map[string]any{"value": fmt.Sprintf("%.2f", plan.Price), "currency": "RUB"},
					"payment_subject": "service",
					"payment_mode":    "full_prepayment", // ← FIX
					"vat_code":        1,
				},
			},
		},
		"metadata": map[string]any{
			"bot_id":       botID,
			"telegram_id":  fmt.Sprintf("%d", telegramID),
			"payment_type": "subscription",
			"plan_code":    plan.Code,
		},
	}

	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка формирования HTTP-запроса в YooKassa")
		return "", fmt.Errorf("build request: %w", err)
	}

	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	// 4. Выполняем запрос
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка сети при обращении к YooKassa")
		return "", fmt.Errorf("yookassa request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("yookassa returned %d: %s", resp.StatusCode, string(raw))
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("YooKassa отклонила подписку (код %d)", resp.StatusCode))
		return "", err
	}

	// 5. Парсим JSON
	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(raw, &yresp); err != nil {
		s.notifier.Notify(ctx, botID, err, "Невалидный JSON от YooKassa")
		return "", fmt.Errorf("decode yookassa: %w", err)
	}

	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		err := fmt.Errorf("invalid yookassa response: %s", string(raw))
		s.notifier.Notify(ctx, botID, err, "Пустой ответ от YooKassa")
		return "", err
	}

	// 6. Сохраняем
	now := time.Now()
	planID := int64(plan.ID)

	sub := &ports.Subscription{
		BotID:             botID,
		TelegramID:        telegramID,
		PlanID:            &planID, // ← FIX
		Status:            "pending",
		StartedAt:         &now,
		ExpiresAt:         nil,
		YookassaPaymentID: &yresp.ID,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка сохранения подписки в БД")
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

	if sub.PlanID == nil {
		err := fmt.Errorf("subscription %d has nil plan_id", sub.ID)
		s.notifier.Notify(ctx, sub.BotID, err,
			"Webhook YooKassa: plan_id is NULL")
		return err
	}

	plan, err := s.tariffRepo.GetByID(ctx, sub.BotID, int(*sub.PlanID))
	if err != nil {
		s.notifier.Notify(ctx, sub.BotID, err,
			"Ошибка загрузки тарифного плана при активации")
		return fmt.Errorf("load plan: %w", err)
	}

	if plan == nil {
		err := fmt.Errorf("plan not found id=%d", *sub.PlanID)
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

func (s *SubscriptionService) ActivateTrial(
	ctx context.Context,
	botID string,
	telegramID int64,
	planCode string,
) error {

	// 1. Проверяем: был ли уже trial
	exists, err := s.trialRepo.Exists(ctx, botID, telegramID)
	if err != nil {
		return err
	}

	// если trial уже был — МОЛЧА выходим
	if exists {
		return nil
	}

	// 2. Ищем trial-тариф
	plan, err := s.tariffRepo.GetTrial(ctx, botID)
	if err != nil {
		return err
	}
	if plan == nil {
		return fmt.Errorf("trial tariff not configured")
	}
	if plan.Code != planCode || !plan.IsTrial {
		return fmt.Errorf("tariff is not trial: %s", planCode)
	}

	// 3. Даты
	start := time.Now()
	exp := start.Add(time.Duration(plan.DurationMinutes) * time.Minute)
	planID := int64(plan.ID)

	// 4. Создаём подписку
	sub := &ports.Subscription{
		BotID:      botID,
		TelegramID: telegramID,
		PlanID:     &planID,
		Status:     "active",
		StartedAt:  &start,
		ExpiresAt:  &exp,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return err
	}

	// 5. Фиксируем факт trial
	if err := s.trialRepo.Create(ctx, botID, telegramID); err != nil {
		// подписка создана — не откатываем
		return nil
	}

	// 6. Минуты
	if plan.VoiceMinutes > 0 {
		_ = s.repo.AddVoiceMinutes(ctx, botID, telegramID, plan.VoiceMinutes)
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

func (s *SubscriptionService) AddMinutesFromPackage(
	ctx context.Context,
	botID string,
	telegramID int64,
	packageID int64,
) error {

	pkg, err := s.minuteSvc.GetByID(ctx, botID, packageID)
	if err != nil {
		return err
	}
	if pkg == nil || !pkg.Active {
		return fmt.Errorf("invalid minute package: %d", packageID)
	}

	// теперь используем repo только для изменения подписки:
	return s.repo.AddVoiceMinutes(ctx, botID, telegramID, float64(pkg.Minutes))
}

func (s *SubscriptionService) CleanupPending(ctx context.Context, olderThan time.Duration) error {
	return s.repo.CleanupPending(ctx, olderThan)
}

func (s *SubscriptionService) Delete(
	ctx context.Context,
	botID string,
	telegramID int64,
) error {

	err := s.repo.Delete(ctx, botID, telegramID)
	if err != nil {
		s.notifier.Notify(
			ctx,
			botID,
			err,
			fmt.Sprintf("Ошибка удаления подписки (tg=%d)", telegramID),
		)
		return err
	}

	return nil
}

func (s *SubscriptionService) CleanupExpiredTrials(
	ctx context.Context,
	botID string,
) error {

	// 1) получаем trial-тариф
	trial, err := s.tariffRepo.GetTrial(ctx, botID)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка загрузки trial-тарифа (cleanup)")
		return err
	}
	if trial == nil {
		return nil // trial не настроен — нечего чистить
	}

	now := time.Now()

	// 2) берём все подписки
	subs, err := s.repo.ListAll(ctx)
	if err != nil {
		s.notifier.Notify(ctx, botID, err, "Ошибка чтения подписок (cleanup)")
		return err
	}

	// 3) фильтруем только истёкшие trial
	for _, sub := range subs {
		if sub.BotID != botID {
			continue
		}
		if sub.PlanID == nil || *sub.PlanID != int64(trial.ID) {
			continue
		}
		if sub.ExpiresAt == nil || sub.ExpiresAt.After(now) {
			continue
		}

		// 4) уведомляем пользователя
		_ = s.notifier.UserNotify(
			ctx,
			botID,
			sub.TelegramID, // chatID == telegramID
			"⏳ Пробный период закончился. Чтобы продолжить — оформи подписку в меню.",
		)

	}

	return nil
}

func (s *SubscriptionService) NotifyExpiredTrials(ctx context.Context) error {
	subs, err := s.repo.GetExpiredTrialsForNotify(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		// 1. помечаем подписку как истёкшую
		if err := s.repo.UpdateStatus(ctx, sub.ID, "expired"); err != nil {
			continue
		}

		// 2. помечаем, что уведомление отправлено
		if err := s.repo.MarkTrialNotified(ctx, sub.ID); err != nil {
			continue
		}

		// 3. здесь НИЧЕГО БОЛЬШЕ
	}

	return nil
}
