package ports

import "context"

type PaymentProvider interface {
	// Возвращает redirect URL для оплаты
	CreateMinutePackagePayment(ctx context.Context, botID string, telegramID int64, packageID int64, price float64, title string, minutes int) (payURL string, providerPaymentID string, err error)

	// Возвращает redirect URL для оплаты подписки
	CreateSubscriptionPayment(
		ctx context.Context,
		botID string,
		telegramID int64,
		planCode string,
		price float64,
		invoiceID string, // <-- добавили
	) (string, string, error)
}
