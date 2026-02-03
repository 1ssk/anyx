package services

import (
	"time"

	"anyx/internal/models"
)

// HasAccess проверяет доступ по подписке или пробному периоду.
func HasAccess(subscription models.Subscription) bool {
	now := time.Now()
	if subscription.Status == "active" && subscription.CurrentPeriodEnd.After(now) {
		return true
	}
	if subscription.Status == "trial" && subscription.TrialEndsAt.After(now) {
		return true
	}
	if subscription.Status == "pending" && subscription.TrialEndsAt.After(now) {
		return true
	}
	return false
}
