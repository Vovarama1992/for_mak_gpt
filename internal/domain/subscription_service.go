package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	"github.com/Vovarama1992/make_ziper/internal/notificator"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/trial"
)

type SubscriptionService struct {
	repo            ports.SubscriptionRepo
	tariffRepo      ports.TariffRepo
	trialRepo       trial.RepoInf
	minuteSvc       minutes_packages.MinutePackageService
	notifier        notificator.Notificator
	paymentProvider ports.PaymentProvider
}

func NewSubscriptionService(
	repo ports.SubscriptionRepo,
	tariffRepo ports.TariffRepo,
	trialRepo trial.RepoInf,
	minuteSvc minutes_packages.MinutePackageService,
	notifier notificator.Notificator,
	paymentProvider ports.PaymentProvider,
) ports.SubscriptionService {
	return &SubscriptionService{
		repo:            repo,
		tariffRepo:      tariffRepo,
		trialRepo:       trialRepo,
		minuteSvc:       minuteSvc,
		notifier:        notifier,
		paymentProvider: paymentProvider,
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

	tariffs, err := s.tariffRepo.ListAll(ctx)
	if err != nil {
		return "", err
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

	// ВАЖНО: генерим InvoiceId сами
	invoiceID := fmt.Sprintf("sub_%d_%d", telegramID, time.Now().Unix())

	payURL, _, err := s.paymentProvider.CreateSubscriptionPayment(
		ctx,
		botID,
		telegramID,
		plan.Code,
		plan.Price,
		invoiceID, // передаём внутрь
	)
	if err != nil {
		return "", err
	}

	now := time.Now()
	planID := int64(plan.ID)

	sub := &ports.Subscription{
		BotID:             botID,
		TelegramID:        telegramID,
		PlanID:            &planID,
		Status:            "pending",
		StartedAt:         &now,
		YookassaPaymentID: &invoiceID, // <-- ВМЕСТО Model.Id
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return "", err
	}

	return payURL, nil
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

func (s *SubscriptionService) ExpireAndNotifyTrials(ctx context.Context) error {
	subs, err := s.repo.ExpireDue(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		// если это trial и ещё не уведомляли
		if sub.PlanID != nil {
			trial, _ := s.tariffRepo.GetTrial(ctx, sub.BotID)
			if trial != nil && int64(trial.ID) == *sub.PlanID {
				// отправка уведомления + MarkTrialNotified
			}
		}
	}
	return nil
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

	// если был trial — чистим факт использования
	used, err := s.trialRepo.Exists(ctx, botID, telegramID)
	if err != nil {
		return err
	}

	if used {
		if err := s.trialRepo.Delete(ctx, botID, telegramID); err != nil {
			s.notifier.Notify(ctx, botID, err,
				fmt.Sprintf("Ошибка очистки trial_usage (tg=%d)", telegramID))
			return err
		}
	}

	// удаляем подписку
	if err := s.repo.Delete(ctx, botID, telegramID); err != nil {
		s.notifier.Notify(ctx, botID, err,
			fmt.Sprintf("Ошибка удаления подписки (tg=%d)", telegramID))
		return err
	}

	return nil
}

func (s *SubscriptionService) CleanupExpiredTrials(
	ctx context.Context,
	botID string,
) error {

	trial, err := s.tariffRepo.GetTrial(ctx, botID)
	if err != nil || trial == nil {
		return err
	}

	now := time.Now()

	subs, err := s.repo.ListAll(ctx)
	if err != nil {
		return err
	}

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

		// ❗ ТУТ НИЧЕГО НЕ ОТПРАВЛЯЕМ
		// максимум — смена статуса, если нужно
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

func (s *SubscriptionService) UpdateLimits(
	ctx context.Context,
	subscriptionID int64,
	status string,
	expiresAt *time.Time,
	voiceMinutes float64,
) error {

	// если статус не передали — не ломаемся
	if status == "" {
		status = "active"
	}

	// expiresAt обязателен на уровне repo → тут защищаемся
	if expiresAt == nil {
		return fmt.Errorf("expiresAt is required")
	}

	if err := s.repo.UpdateLimits(
		ctx,
		subscriptionID,
		*expiresAt,
		voiceMinutes,
		status,
	); err != nil {
		s.notifier.Notify(
			ctx,
			"admin",
			err,
			fmt.Sprintf("Ошибка ручного редактирования подписки id=%d", subscriptionID),
		)
		return err
	}

	return nil
}
