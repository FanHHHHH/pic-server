package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getServerInfo() (uploadDir, host, port string) {
	uploadDir = viper.GetString("server.uploadDir")
	host = viper.GetString("server.host")
	port = viper.GetString("server.port")
	return uploadDir, host, port
}

func UploadService(c *gin.Context) {

	uploadDir, host, port := getServerInfo()

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	dst := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "文件上传成功",
		"url":     fmt.Sprintf("http://%s:%s/uploads/%s", host, port, file.Filename),
	})
}

func GetFileDetail(c *gin.Context) {
	uploadDir, host, port := getServerInfo()

	filename := c.Param("filename")
	filePath := filepath.Join(uploadDir, filename)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":     fileInfo.Name(),
		"size":     fileInfo.Size(),
		"mod_time": fileInfo.ModTime().Format(time.RFC3339),
		"url":      fmt.Sprintf("http://%s:%s/uploads/%s", host, port, fileInfo.Name()),
	})

}

func ListPics(c *gin.Context) {
	uploadDir, host, port := getServerInfo()

	files, err := ioutil.ReadDir(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法读取上传目录",
		})
		return
	}

	images := []gin.H{}
	for _, file := range files {
		if !file.IsDir() {
			images = append(images, gin.H{
				"name": file.Name(),
				"url":  fmt.Sprintf("http://%s:%s/uploads/%s", host, port, file.Name()),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})

}
