package minutes_packages

import "context"

type MinutePackage struct {
	ID      int64   `json:"id"`
	BotID   string  `json:"bot_id"`
	Name    string  `json:"name"`
	Minutes int     `json:"minutes"`
	Price   float64 `json:"price"`
	Active  bool    `json:"active"`
}

type MinutePackageRepo interface {
	Create(ctx context.Context, pkg *MinutePackage) error
	Update(ctx context.Context, pkg *MinutePackage) error
	Delete(ctx context.Context, botID string, id int64) error
	GetByID(ctx context.Context, botID string, id int64) (*MinutePackage, error)
	ListAll(ctx context.Context, botID string) ([]*MinutePackage, error)
}

type MinutePackageService interface {
	// CRUD
	Create(ctx context.Context, pkg *MinutePackage) error
	Update(ctx context.Context, pkg *MinutePackage) error
	Delete(ctx context.Context, botID string, id int64) error
	GetByID(ctx context.Context, botID string, id int64) (*MinutePackage, error)
	ListAll(ctx context.Context, botID string) ([]*MinutePackage, error)

	// покупка пакета минут (создание оплаты)
	CreatePayment(ctx context.Context, botID string, telegramID int64, packageID int64) (string, error)
}
