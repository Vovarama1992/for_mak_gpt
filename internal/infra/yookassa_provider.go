package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ports"
)

type YooKassaProvider struct {
	httpClient *http.Client
}

func NewYooKassaProvider() ports.PaymentProvider {
	return &YooKassaProvider{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// ----------------------------------------------------
// Minute packages
// ----------------------------------------------------

func (p *YooKassaProvider) CreateMinutePackagePayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	packageID int64,
	price float64,
	title string,
	minutes int,
) (string, string, error) {

	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secret := os.Getenv("YOOKASSA_SECRET_KEY")

	log.Printf("[YK] start create payment bot=%s tg=%d pkg=%d price=%.2f",
		botID, telegramID, packageID, price)

	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", price),
			"currency": "RUB",
		},
		"capture": true,
		"description": fmt.Sprintf(
			"Minute package '%s' (%d min)", title, minutes,
		),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": "https://aifulls.com/success.html",
		},
		"metadata": map[string]any{
			"bot_id":       botID,
			"telegram_id":  fmt.Sprintf("%d", telegramID),
			"package_id":   fmt.Sprintf("%d", packageID),
			"payment_type": "minute_package",
		},
	}

	reqBody, _ := json.Marshal(body)

	log.Printf("[YK] request url=%s", apiURL)
	log.Printf("[YK] request body=%s", string(reqBody))

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[YK] build request error: %v", err)
		return "", "", err
	}

	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("[YK] http error: %v", err)
		return "", "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	log.Printf("[YK] response status=%d", resp.StatusCode)
	log.Printf("[YK] response body=%s", string(raw))

	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("yookassa error status=%d body=%s",
			resp.StatusCode, string(raw))
	}

	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}

	if err := json.Unmarshal(raw, &yresp); err != nil {
		log.Printf("[YK] json parse error: %v", err)
		return "", "", err
	}

	log.Printf("[YK] parsed id=%s url=%s", yresp.ID, yresp.Confirmation.URL)

	return yresp.Confirmation.URL, yresp.ID, nil
}

// ----------------------------------------------------
// Subscriptions
// ----------------------------------------------------

func (p *YooKassaProvider) CreateSubscriptionPayment(
	ctx context.Context,
	botID string,
	telegramID int64,
	planCode string,
	price float64,
	invoiceID string,
) (string, string, error) {

	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secret := os.Getenv("YOOKASSA_SECRET_KEY")

	log.Printf("[YK] create subscription bot=%s tg=%d plan=%s price=%.2f invoice=%s",
		botID, telegramID, planCode, price, invoiceID)

	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", price),
			"currency": "RUB",
		},
		"capture": true,
		"description": fmt.Sprintf(
			"Subscription '%s'", planCode,
		),
		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": "https://aifulls.com/success.html",
		},
		"metadata": map[string]any{
			"bot_id":       botID,
			"telegram_id":  fmt.Sprintf("%d", telegramID),
			"payment_type": "subscription",
			"plan_code":    planCode,
			"invoice_id":   invoiceID,
		},
	}

	reqBody, _ := json.Marshal(body)

	log.Printf("[YK] request url=%s", apiURL)
	log.Printf("[YK] request body=%s", string(reqBody))

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[YK] build request error: %v", err)
		return "", "", err
	}

	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("[YK] http error: %v", err)
		return "", "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	log.Printf("[YK] response status=%d", resp.StatusCode)
	log.Printf("[YK] response body=%s", string(raw))

	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("yookassa error status=%d body=%s",
			resp.StatusCode, string(raw))
	}

	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}

	if err := json.Unmarshal(raw, &yresp); err != nil {
		log.Printf("[YK] json parse error: %v", err)
		return "", "", err
	}

	log.Printf("[YK] parsed id=%s url=%s", yresp.ID, yresp.Confirmation.URL)

	return yresp.Confirmation.URL, yresp.ID, nil
}
