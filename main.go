package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/wechat-task/api/docs"
	"github.com/wechat-task/api/internal/config"
	"github.com/wechat-task/api/internal/database"
	"github.com/wechat-task/api/internal/handler"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/middleware"
	"github.com/wechat-task/api/internal/repository"
	"github.com/wechat-task/api/internal/service"
)

// @title           WeChat Task API
// @version         1.0
// @description     User management API with Passkeys (WebAuthn) authentication

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.Load()
	logger.Init(cfg)

	logger.Info("Starting WeChat Task API...")
	logger.Infof("Server mode: %s", cfg.Server.Mode)
	logger.Infof("Server port: %d", cfg.Server.Port)

	gin.SetMode(cfg.Server.Mode)

	db, err := database.Init(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to init database:", err)
	}

	logger.Info("Database connected successfully")

	if err := database.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate database:", err)
	}

	logger.Info("Database migration completed")

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
		logger.Fatal("Failed to init auth service:", err)
	}

	userService := service.NewUserService(userRepo)

	jwtService := service.NewJWTService(cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authService, userService, jwtService)
	userHandler := handler.NewUserHandler(userService)

	r := gin.Default()

	r.Use(middleware.Logger())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/start", authHandler.BeginAuth)
		auth.POST("/finish", authHandler.FinishAuth)
	}

	user := r.Group("/api/v1/user")
	user.Use(middleware.Auth(jwtService))
	{
		user.GET("/me", userHandler.GetCurrentUser)
		user.PUT("/username", userHandler.SetUsername)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	logger.Infof("Swagger UI available at http://localhost%s/swagger/index.html", addr)

	r.Run(addr)
}
