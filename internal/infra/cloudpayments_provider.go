package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type CloudPaymentsProvider struct {
	httpClient *http.Client
}

func NewCloudPaymentsProvider() ports.PaymentProvider {
	return &CloudPaymentsProvider{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type cpResponse struct {
	Success bool   `json:"Success"`
	Message string `json:"Message"`
	Model   struct {
		URL string `json:"Url"`
		Id  string `json:"Id"`
	} `json:"Model"`
}

// ----------------------------------------------------
// Minute packages
// ----------------------------------------------------

func (p *CloudPaymentsProvider) CreateMinutePackagePayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	packageID int64,
	price float64,
	title string,
	minutes int,
) (string, string, error) {

	body := map[string]any{
		"Amount":      price,
		"Currency":    "RUB",
		"Description": fmt.Sprintf("Minute package '%s' (%d min)", title, minutes),
		"InvoiceId":   fmt.Sprintf("pkg_%d_%d", telegramID, time.Now().Unix()),
		"AccountId":   fmt.Sprintf("%d", telegramID),
		"JsonData": map[string]any{
			"payment_type": "minute_package",
			"bot_id":       botID,
			"telegram_id":  telegramID,
			"package_id":   packageID,
		},
	}

	return p.createOrder(ctx, body)
}

// ----------------------------------------------------
// Subscriptions
// ----------------------------------------------------

func (p *CloudPaymentsProvider) CreateSubscriptionPayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	planCode string,
	price float64,
) (string, string, error) {

	body := map[string]any{
		"Amount":      price,
		"Currency":    "RUB",
		"Description": fmt.Sprintf("Subscription '%s'", planCode),
		"InvoiceId":   fmt.Sprintf("sub_%d_%d", telegramID, time.Now().Unix()),
		"AccountId":   fmt.Sprintf("%d", telegramID),
		"JsonData": map[string]any{
			"payment_type": "subscription",
			"bot_id":       botID,
			"telegram_id":  telegramID,
			"plan_code":    planCode,
		},
	}

	return p.createOrder(ctx, body)
}

// ----------------------------------------------------

func (p *CloudPaymentsProvider) createOrder(
	ctx context.Context,
	payload map[string]any,
) (string, string, error) {

	apiURL := "https://api.cloudpayments.ru/orders/create"
	publicID := os.Getenv("CLOUDPAYMENTS_PUBLIC_ID")
	secret := os.Getenv("CLOUDPAYMENTS_API_SECRET")

	reqBody, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	req.SetBasicAuth(publicID, secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var cp cpResponse
	if err := json.Unmarshal(raw, &cp); err != nil {
		return "", "", err
	}

	if !cp.Success {
		return "", "", fmt.Errorf("cloudpayments error: %s", cp.Message)
	}

	return cp.Model.URL, cp.Model.Id, nil
}
