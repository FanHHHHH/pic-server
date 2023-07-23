package router

import (
	"pic-server/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Router() *gin.Engine {
	r := gin.Default()

	r.Static("/uploads", viper.GetString("server.uploadDir"))

	r.POST("/upload", service.UploadService)
	r.GET("/uploadList", service.ListPics)
	r.GET("/uploadDetail/:filename", service.GetFileDetail)

	return r
}
