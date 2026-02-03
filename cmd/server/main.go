package main

import (
	"log"

	"anyx/internal/config"
	"anyx/internal/handlers"
	"anyx/internal/middleware"
	"anyx/internal/models"
	"anyx/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Точка входа приложения.
func main() {
	cfg := config.Load()

	db, err := gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatal("не удалось открыть базу данных")
	}

	if err := db.AutoMigrate(&models.User{}, &models.InviteLink{}, &models.Click{}, &models.Subscription{}); err != nil {
		log.Fatal("не удалось выполнить миграции")
	}

	if err := services.EnsureAdmin(db, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatal("не удалось создать администратора")
	}

	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTTTL)
	billingService := services.NewBillingService(cfg)
	yookassaService := services.NewYooKassaService(cfg)

	authHandler := handlers.NewAuthHandler(db, jwtService, cfg)
	linkHandler := handlers.NewLinkHandler(db, cfg)
	userHandler := handlers.NewUserHandler(db, cfg)
	billingHandler := handlers.NewBillingHandler(db, billingService, yookassaService, cfg)
	adminHandler := handlers.NewAdminHandler(db)

	r := gin.Default()
	r.Static("/static", "web/static")

	r.GET("/", func(c *gin.Context) { c.File("web/index.html") })
	r.GET("/login", func(c *gin.Context) { c.File("web/login.html") })
	r.GET("/register", func(c *gin.Context) { c.File("web/register.html") })
	r.GET("/dashboard", func(c *gin.Context) { c.File("web/dashboard.html") })
	r.GET("/stats", func(c *gin.Context) { c.File("web/stats.html") })
	r.GET("/admin", func(c *gin.Context) { c.File("web/admin.html") })

	r.GET("/api/config", handlers.ConfigHandler(cfg))
	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)

	r.POST("/api/webhook/yookassa", billingHandler.Webhook)
	r.POST("/webhook", billingHandler.Webhook)
	r.POST("/webhook/yookassa", billingHandler.Webhook)

	r.GET("/i/:code", linkHandler.Redirect)

	auth := r.Group("/api")
	auth.Use(middleware.JWTAuth(jwtService, db))
	auth.GET("/me", userHandler.Me)
	auth.GET("/links", linkHandler.List)
	auth.POST("/links", linkHandler.Create)
	auth.GET("/links/:id/stats", linkHandler.Stats)
	auth.GET("/billing/status", billingHandler.Status)
	auth.POST("/billing/checkout", billingHandler.Checkout)

	admin := auth.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/users", adminHandler.Users)
	admin.GET("/users/:id", adminHandler.UserDetail)
	admin.DELETE("/users/:id", adminHandler.DeleteUser)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	if err := r.Run(cfg.BindAddr); err != nil {
		log.Fatal("не удалось запустить сервер")
	}
}
