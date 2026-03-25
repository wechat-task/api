package main

import (
	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	r.Use(middleware.Logger())

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	api := r.Group("/api/v1")
	{
		api.GET("/tasks", taskHandler.ListTasks)
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.PUT("/tasks/:id/complete", taskHandler.CompleteTask)
	}

	r.Run(":8080")
}
