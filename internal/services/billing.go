package services

import (
	"fmt"
	"time"

	"anyx/internal/config"
)

// BillingService описывает работу с оплатой.
type BillingService struct {
	cfg config.Config
}

// NewBillingService создает сервис биллинга.
func NewBillingService(cfg config.Config) *BillingService {
	return &BillingService{cfg: cfg}
}

// CheckoutURL возвращает ссылку для оплаты (заглушка для ЮKassa).
func (s *BillingService) CheckoutURL(userID uint) string {
	return fmt.Sprintf("%s/billing/checkout?user=%d&timestamp=%d", s.cfg.BaseURL, userID, time.Now().Unix())
}
