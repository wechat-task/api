package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/wechat-task/api/internal/config"
	"github.com/wechat-task/api/internal/database"
	"github.com/wechat-task/api/internal/handler"
	"github.com/wechat-task/api/internal/middleware"
	"github.com/wechat-task/api/internal/repository"
	"github.com/wechat-task/api/internal/service"
	"log"
)

func main() {
	cfg := config.Load()

	gin.SetMode(cfg.Server.Mode)

	db, err := database.Init(cfg.Database.URL)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userRepo := repository.NewUserRepository(db)
	credentialRepo := repository.NewCredentialRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	sessionService := service.NewSessionService(sessionRepo)
	sessionService.CleanupExpired()

	authService, err := service.NewAuthService(
		webauthn.Config{
			RPDisplayName: cfg.WebAuthn.RPDisplayName,
			RPID:          cfg.WebAuthn.RPID,
			RPOrigins:     cfg.WebAuthn.RPOrigins,
		},
		userRepo,
		credentialRepo,
		sessionService,
	)
	if err != nil {
		log.Fatal("Failed to init auth service:", err)
	}

	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	r := gin.Default()

	r.Use(middleware.Logger())

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/start", authHandler.BeginAuth)
		auth.POST("/finish", authHandler.FinishAuth)
	}

	user := r.Group("/api/v1/user")
	user.Use(middleware.Auth())
	{
		user.GET("/me", userHandler.GetCurrentUser)
		user.PUT("/username", userHandler.SetUsername)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	r.Run(addr)
}
