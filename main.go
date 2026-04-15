package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/wechat-task/api/docs"
	"github.com/wechat-task/api/internal/channels"
	"github.com/wechat-task/api/internal/channels/ilink"
	"github.com/wechat-task/api/internal/channels/lark"
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

	botRepo := repository.NewBotRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	botService := service.NewBotService(botRepo)
	channelProviders := channels.NewRegistry(
		ilink.NewProvider(),
		lark.NewProvider(),
	)
	channelContextRepo := repository.NewChannelContextRepository(db)
	channelService := service.NewChannelService(channelRepo, botRepo, channelContextRepo, channelProviders)
	channelService.RecoverPendingChannels()
	channelService.StartActiveChannelPollers()

	// Skill repositories
	skillRepo := repository.NewSkillRepository(db)
	skillSubscriptionRepo := repository.NewSkillSubscriptionRepository(db)
	skillExecutionLogRepo := repository.NewSkillExecutionLogRepository(db)
	userLLMConfigRepo := repository.NewUserLLMConfigRepository(db)

	// Skill service
	skillService := service.NewSkillService(
		skillRepo,
		skillSubscriptionRepo,
		skillExecutionLogRepo,
		userLLMConfigRepo,
	)

	jwtService := service.NewJWTService(cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authService, jwtService)
	userHandler := handler.NewUserHandler(userService)
	botHandler := handler.NewBotHandler(botService)
	channelHandler := handler.NewChannelHandler(channelService)
	skillHandler := handler.NewSkillHandler(skillService)

	r := gin.Default()

	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	r.GET("/", handler.Index)
	r.GET("/health", handler.Health(db))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	passkey := r.Group("/api/v1/auth/passkey")
	{
		passkey.POST("/register/options", authHandler.RegisterOptions)
		passkey.POST("/register/verify", authHandler.RegisterVerify)
		passkey.POST("/login/options", authHandler.LoginOptions)
		passkey.POST("/login/verify", authHandler.LoginVerify)
	}

	user := r.Group("/api/v1/user")
	user.Use(middleware.Auth(jwtService))
	{
		user.GET("/me", userHandler.GetCurrentUser)
		user.PUT("/profile", userHandler.UpdateProfile)
	}

	bots := r.Group("/api/v1/bots")
	bots.Use(middleware.Auth(jwtService))
	{
		bots.POST("", botHandler.CreateBot)
		bots.GET("", botHandler.ListBots)
		bots.GET("/:id", botHandler.GetBot)
		bots.PUT("/:id", botHandler.UpdateBot)
		bots.DELETE("/:id", botHandler.DeleteBot)

		channels := bots.Group("/:id/channels")
		{
			channels.POST("/lark", channelHandler.CreateLarkChannel)
			channels.POST("/wechat-clawbot", channelHandler.CreateWechatClawbotChannel)
			channels.GET("", channelHandler.ListChannels)
			channels.DELETE("/:channelId", channelHandler.DeleteChannel)
			channels.POST("/:channelId/send", channelHandler.SendMessage)
		}
	}

	// Skill routes
	skills := r.Group("/api/v1/skills")
	skills.Use(middleware.Auth(jwtService))
	{
		skills.POST("", skillHandler.CreateSkill)
		skills.GET("/me", skillHandler.GetMySkills)
		skills.GET("/search", skillHandler.SearchSkills)
		skills.PUT("/:id", skillHandler.UpdateSkill)
		skills.DELETE("/:id", skillHandler.DeleteSkill)
		skills.POST("/:id/publish", skillHandler.PublishSkill)
		skills.POST("/:id/archive", skillHandler.ArchiveSkill)
		skills.POST("/:id/subscribe", skillHandler.SubscribeToSkill)
		skills.DELETE("/:id/subscribe", skillHandler.UnsubscribeFromSkill)
		skills.GET("/subscriptions", skillHandler.GetUserSubscriptions)
		skills.GET("/subscriptions/:id", skillHandler.GetSubscription)
		skills.PUT("/subscriptions/:id", skillHandler.UpdateSubscription)
		skills.DELETE("/subscriptions/:id", skillHandler.DeleteSubscription)
	}

	// Public skill routes (no auth required for viewing)
	publicSkills := r.Group("/api/v1/skills")
	{
		publicSkills.GET("/:id", skillHandler.GetSkill)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Server listening on %s", addr)
	logger.Infof("Swagger UI available at http://localhost%s/swagger/index.html", addr)

	r.Run(addr)
}
