package minutes_packages

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
)

type service struct {
	repo       MinutePackageRepo
	httpClient *http.Client
}

func NewService(repo MinutePackageRepo) MinutePackageService {
	return &service{
		repo:       repo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
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

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id int64) (*MinutePackage, error) {
	return s.repo.GetByID(ctx, id)
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

	// 1. пакет
	pkg, err := s.repo.GetByID(ctx, packageID)
	if err != nil {
		return "", fmt.Errorf("load minute package: %w", err)
	}
	if pkg == nil || !pkg.Active {
		return "", fmt.Errorf("minute package not found or inactive: %d", packageID)
	}

	// 2. ENVs
	apiURL := os.Getenv("YOOKASSA_API_URL")
	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secret := os.Getenv("YOOKASSA_SECRET_KEY")

	if apiURL == "" || shopID == "" || secret == "" {
		return "", fmt.Errorf("missing yookassa ENV variables")
	}
	if !strings.Contains(apiURL, "/v3/payments") {
		apiURL = strings.TrimRight(apiURL, "/") + "/v3/payments"
	}

	// === 3. body (с receipt — фикс) ===
	body := map[string]any{
		"amount": map[string]any{
			"value":    fmt.Sprintf("%.2f", pkg.Price),
			"currency": "RUB",
		},
		"capture": true,
		"description": fmt.Sprintf(
			"Minute package '%s' (%d min)",
			pkg.Name, pkg.Minutes,
		),

		"confirmation": map[string]any{
			"type":       "redirect",
			"return_url": "https://aifulls.com/success.html",
		},

		// === ДОБАВЛЕННЫЙ ЧЕК (ОБЯЗАТЕЛЕН) ===
		"receipt": map[string]any{
			"customer": map[string]any{
				"phone": "79384095762",
			},
			"items": []map[string]any{
				{
					"description": fmt.Sprintf("Пакет минут '%s' (%d мин)", pkg.Name, pkg.Minutes),
					"quantity":    "1",
					"amount": map[string]any{
						"value":    fmt.Sprintf("%.2f", pkg.Price),
						"currency": "RUB",
					},
					"vat_code": 1,
				},
			},
		},

		"metadata": map[string]any{
			"bot_id":       botID,
			"telegram_id":  fmt.Sprintf("%d", telegramID),
			"package_id":   fmt.Sprintf("%d", packageID),
			"payment_type": "minute_package",
		},
	}

	// === 4. execute ===
	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.SetBasicAuth(shopID, secret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to yookassa: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("yookassa returned %d: %s", resp.StatusCode, string(raw))
	}

	// === 5. decode ===
	var yresp struct {
		ID           string `json:"id"`
		Confirmation struct {
			URL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(raw, &yresp); err != nil {
		return "", fmt.Errorf("decode yookassa: %w", err)
	}

	if yresp.ID == "" || yresp.Confirmation.URL == "" {
		return "", fmt.Errorf("invalid yookassa response: %s", string(raw))
	}

	return yresp.Confirmation.URL, nil
}
