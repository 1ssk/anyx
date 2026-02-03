package models

import "time"

// User хранит учетные данные пользователя.
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time
}

// InviteLink описывает инвайт-ссылку.
type InviteLink struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Code      string    `gorm:"uniqueIndex"`
	TargetURL string    `gorm:"not null"`
	CreatedAt time.Time
}

// Click фиксирует клик по инвайт-ссылке.
type Click struct {
	ID        uint      `gorm:"primaryKey"`
	LinkID    uint      `gorm:"index"`
	ClickedAt time.Time `gorm:"index"`
}

// Subscription хранит данные подписки пользователя.
type Subscription struct {
	ID               uint      `gorm:"primaryKey"`
	UserID           uint      `gorm:"uniqueIndex"`
	Status           string    `gorm:"index"`
	TrialEndsAt      time.Time
	CurrentPeriodEnd time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
