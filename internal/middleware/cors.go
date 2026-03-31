package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/config"
)

func CORS() gin.HandlerFunc {
	cfg := config.Get()

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowAllOrigins = true
	}

	return cors.New(corsConfig)
}
