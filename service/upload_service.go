package service

import (
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"pic-server/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/spf13/viper"
)

func getServerInfo() (uploadDir, host, port, compressUploadDir string) {
	uploadDir = viper.GetString("server.uploadDir")
	compressUploadDir = viper.GetString("server.compressUploadDir")
	host = viper.GetString("server.host")
	port = viper.GetString("server.port")
	return uploadDir, host, port, compressUploadDir
}

func compressImage(inputFile, outputFile string, newWidth uint) error {
	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 解码 JPEG 图片
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	// 调整图片大小
	resizedImg := resize.Resize(newWidth, 0, img, resize.Lanczos3)

	// 打开输出文件
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer out.Close()

	// 以 JPEG 格式编码并保存调整大小后的图片
	err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 80})
	if err != nil {
		return fmt.Errorf("编码输出图片失败: %v", err)
	}

	return nil
}

func UploadService(c *gin.Context) {

	uploadDir, host, port, compressUploadDir := getServerInfo()

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "get form err:"+err.Error(), nil)
		return
	}

	dst := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "upload file err:"+err.Error(), nil)
		return
	}

	if err := compressImage(dst, filepath.Join(compressUploadDir, file.Filename), 400); err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "compress file err:"+err.Error(), nil)
		return
	}

	utils.SendJsonResponse(c, http.StatusOK, "success", gin.H{
		"url": fmt.Sprintf("http://%s:%s/compress_uploads/%s", host, port, file.Filename),
	})
}

func GetFileDetail(c *gin.Context) {
	uploadDir, host, port, _ := getServerInfo()

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
		"url":      fmt.Sprintf("http://%s:%s/compress_uploads/%s", host, port, fileInfo.Name()),
	})

}

func ListPics(c *gin.Context) {
	_, host, port, compressUploadDir := getServerInfo()

	files, err := ioutil.ReadDir(compressUploadDir)
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
				"name":   file.Name(),
				"url":    fmt.Sprintf("http://%s:%s/compress_uploads/%s", host, port, file.Name()),
				"rawUrl": fmt.Sprintf("http://%s:%s/uploads/%s", host, port, file.Name()),
			})
		}
	}

	utils.SendJsonResponse(c, http.StatusOK, "success", gin.H{
		"images": images,
	})

}

func ListRawPics(c *gin.Context) {
	uploadDir, host, port, _ := getServerInfo()

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

	utils.SendJsonResponse(c, http.StatusOK, "success", gin.H{
		"images": images,
	})
}
