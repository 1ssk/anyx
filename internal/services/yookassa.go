package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"anyx/internal/config"
)

// YooKassaService работает с API ЮKassa.
type YooKassaService struct {
	cfg    config.Config
	client *http.Client
}

// NewYooKassaService создает сервис ЮKassa.
func NewYooKassaService(cfg config.Config) *YooKassaService {
	return &YooKassaService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PaymentConfirmation содержит ссылку на оплату.
type PaymentConfirmation struct {
	URL string
}

// CreatePayment создает платеж и возвращает ссылку на оплату.
func (s *YooKassaService) CreatePayment(userID uint, amount int, returnURL string) (PaymentConfirmation, error) {
	payload := map[string]interface{}{
		"amount": map[string]string{
			"value":    fmt.Sprintf("%d.00", amount),
			"currency": "RUB",
		},
		"confirmation": map[string]string{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"capture":     true,
		"description": fmt.Sprintf("Оплата подписки пользователя %d", userID),
		"metadata": map[string]interface{}{
			"user_id": userID,
			"months":  1,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return PaymentConfirmation{}, err
	}

	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewBuffer(body))
	if err != nil {
		return PaymentConfirmation{}, err
	}

	idempotencyKey := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	credential := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.cfg.YooShopID, s.cfg.YooSecretKey)))

	req.Header.Set("Authorization", "Basic "+credential)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", idempotencyKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return PaymentConfirmation{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return PaymentConfirmation{}, fmt.Errorf("ошибка ЮKassa: %s", resp.Status)
	}

	var response struct {
		Confirmation struct {
			ConfirmationURL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return PaymentConfirmation{}, err
	}

	return PaymentConfirmation{URL: response.Confirmation.ConfirmationURL}, nil
}

// ParseWebhook читает уведомление ЮKassa.
func (s *YooKassaService) ParseWebhook(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// WebhookEvent описывает payload уведомления.
type WebhookEvent struct {
	Event  string        `json:"event"`
	Object WebhookObject `json:"object"`
}

// WebhookObject описывает объект платежа.
type WebhookObject struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Paid   bool   `json:"paid"`
	Meta   struct {
		UserID uint `json:"user_id"`
		Months int  `json:"months"`
	} `json:"metadata"`
}
