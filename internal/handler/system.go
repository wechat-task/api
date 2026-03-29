package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary      Get project info
// @Description  Returns project details and available API documentation link
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       / [get]
func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":         "WeChat Task API",
		"version":      "1.0.0",
		"description":  "Task management API with Passkeys (WebAuthn) authentication",
		"swagger_docs": "/swagger/index.html",
	})
}

// @Summary      Health check
// @Description  Returns the health status of the service and database connectivity
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func Health(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "ok"
		dbStatus := "up"
		httpStatus := http.StatusOK

		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			status = "degraded"
			dbStatus = "down"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    status,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services": gin.H{
				"database": dbStatus,
			},
		})
	}
}
