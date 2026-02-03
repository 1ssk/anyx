package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config хранит параметры приложения.
type Config struct {
	BindAddr          string
	DatabasePath      string
	JWTSecret         string
	JWTTTL            time.Duration
	AppDomain         string
	BaseURL           string
	TrialDays         int
	MonthlyPrice      int
	YooShopID         string
	YooSecretKey      string
	CheckoutReturn    string
	AdminEmail        string
	AdminPassword     string
	WebhookURL        string
	WebhookSecret     string
	WebhookIPAllowlist string
}

// Load загружает конфигурацию из .env и окружения.
func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		BindAddr:           getEnv("BIND_ADDR", ":8080"),
		DatabasePath:       getEnv("DATABASE_PATH", "invite.db"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret"),
		JWTTTL:             getDuration("JWT_TTL", 24*time.Hour),
		AppDomain:          getEnv("APP_DOMAIN", "invite.annonyx.ru"),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		TrialDays:          getInt("TRIAL_DAYS", 7),
		MonthlyPrice:       getInt("MONTHLY_PRICE_RUB", 150),
		YooShopID:          getEnv("YOOKASSA_SHOP_ID", ""),
		YooSecretKey:       getEnv("YOOKASSA_SECRET", ""),
		CheckoutReturn:     getEnv("CHECKOUT_RETURN_URL", "http://localhost:8080/dashboard"),
		AdminEmail:         getEnv("ADMIN_EMAIL", ""),
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
		WebhookSecret:      getEnv("WEBHOOK_SECRET", ""),
		WebhookIPAllowlist: getEnv("WEBHOOK_IP_ALLOWLIST", ""),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET обязателен")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
