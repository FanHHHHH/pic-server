package router

import (
	"pic-server/middleware"
	"pic-server/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Router() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware()).Use(middleware.RateLimit)
	r.Static("/uploads", viper.GetString("server.uploadDir"))
	r.Static("/compress_uploads", viper.GetString("server.compressUploadDir"))

	api := r.Group("/api").Use(middleware.RemoteAuthz())
	{
		api.POST("/upload", service.UploadService)
		api.GET("/uploadRawList", service.ListRawPics)
		api.GET("/uploadList", service.ListPics)
		api.GET("/uploadDetail/:filename", service.GetFileDetail)
	}

	return r
}
