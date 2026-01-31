package minutes_packages

import (
	"context"
	"fmt"

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

// ---------------------------
// CREATE PAYMENT
// ---------------------------
func (s *service) CreatePayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	packageID int64,
) (string, error) {

	pkg, err := s.repo.GetByID(ctx, botID, packageID)
	if err != nil {
		return "", fmt.Errorf("load minute package: %w", err)
	}
	if pkg == nil || !pkg.Active {
		return "", fmt.Errorf("minute package not found or inactive: %d", packageID)
	}

	payURL, _, err := s.paymentProvider.CreateMinutePackagePayment(
		ctx,
		botID,
		telegramID,
		packageID,
		pkg.Price,
		pkg.Name,
		pkg.Minutes,
	)
	if err != nil {
		return "", err
	}

	return payURL, nil
}
