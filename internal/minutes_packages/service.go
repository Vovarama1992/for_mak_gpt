package minutes_packages

import (
	"context"
	"fmt"
	"log"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type service struct {
	repo            MinutePackageRepo
	paymentProvider ports.PaymentProvider
}

func NewService(
	repo MinutePackageRepo,
	paymentProvider ports.PaymentProvider,
) MinutePackageService {
	return &service{
		repo:            repo,
		paymentProvider: paymentProvider,
	}
}

// ---------------------------
// CRUD
// ---------------------------
func (s *service) Create(ctx context.Context, pkg *MinutePackage) error {
	return s.repo.Create(ctx, pkg)
}

func (s *service) Update(ctx context.Context, pkg *MinutePackage) error {
	return s.repo.Update(ctx, pkg)
}

func (s *service) Delete(ctx context.Context, botID string, id int64) error {
	return s.repo.Delete(ctx, botID, id)
}

func (s *service) GetByID(ctx context.Context, botID string, id int64) (*MinutePackage, error) {
	return s.repo.GetByID(ctx, botID, id)
}

func (s *service) ListAll(ctx context.Context) ([]*MinutePackage, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) CreatePayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	packageID int64,
) (string, error) {

	log.Printf("[PAY] create payment start bot=%s tg=%d pkg=%d",
		botID, telegramID, packageID)

	pkg, err := s.repo.GetByID(ctx, botID, packageID)
	if err != nil {
		log.Printf("[PAY] load package error: %v", err)
		return "", fmt.Errorf("load minute package: %w", err)
	}
	if pkg == nil {
		log.Printf("[PAY] package not found pkg=%d", packageID)
		return "", fmt.Errorf("minute package not found: %d", packageID)
	}
	if !pkg.Active {
		log.Printf("[PAY] package inactive pkg=%d", packageID)
		return "", fmt.Errorf("minute package inactive: %d", packageID)
	}

	log.Printf("[PAY] package loaded id=%d price=%.2f minutes=%d",
		pkg.ID, pkg.Price, pkg.Minutes)

	payURL, payID, err := s.paymentProvider.CreateMinutePackagePayment(
		ctx,
		botID,
		telegramID,
		packageID,
		pkg.Price,
		pkg.Name,
		pkg.Minutes,
	)
	if err != nil {
		log.Printf("[PAY] provider error: %v", err)
		return "", err
	}

	log.Printf("[PAY] created url=%s id=%s", payURL, payID)

	if payURL == "" {
		log.Printf("[PAY] WARNING empty payURL id=%s", payID)
	}

	return payURL, nil
}
