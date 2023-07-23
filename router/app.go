package router

import (
	"pic-server/middleware"
	"pic-server/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Router() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.RateLimit)
	r.Static("/uploads", viper.GetString("server.uploadDir"))

	api := r.Group("/api")
	{
		api.POST("/upload", service.UploadService)
		api.GET("/uploadList", service.ListPics)
		api.GET("/uploadDetail/:filename", service.GetFileDetail)
	}

	return r
}
