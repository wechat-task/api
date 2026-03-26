package main

import (
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

	db, err := database.Init(cfg)
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
			RPDisplayName: cfg.WebAuthnRPDisplayName,
			RPID:          cfg.WebAuthnRPID,
			RPOrigins:     cfg.WebAuthnRPOrigins,
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

	r.Run(":8080")
}
