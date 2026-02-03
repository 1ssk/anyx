package services

import (
	"strings"

	"anyx/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// EnsureAdmin создает администратора, если его нет.
func EnsureAdmin(db *gorm.DB, email, password string) error {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil
	}

	var existing models.User
	if err := db.Where("email = ?", strings.ToLower(email)).First(&existing).Error; err == nil {
		if existing.Role != "admin" {
			existing.Role = "admin"
			return db.Save(&existing).Error
		}
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Email:        strings.ToLower(email),
		PasswordHash: string(hash),
		Role:         "admin",
	}

	return db.Create(&admin).Error
}
