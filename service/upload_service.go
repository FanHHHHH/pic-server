package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	service_utils "pic-server/service/utils"
	"pic-server/utils"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func getServerInfo() (uploadDir, host, port, compressUploadDir string) {
	uploadDir = viper.GetString("server.uploadDir")
	compressUploadDir = viper.GetString("server.compressUploadDir")
	host = viper.GetString("server.host")
	port = viper.GetString("server.port")
	return uploadDir, host, port, compressUploadDir
}

func UploadService(c *gin.Context) {

	uploadDir, host, port, compressUploadDir := getServerInfo()

	file, err := c.FormFile("file")
	if err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "get form err:"+err.Error(), nil)
		return
	}

	fullFilename := file.Filename
	dst := filepath.Join(uploadDir, fullFilename)

	//检查是否重名
	if service_utils.FileExists(dst) {
		ext := filepath.Ext(file.Filename)
		filename := file.Filename[0 : len(file.Filename)-len(ext)]

		fullFilename = filename + "_" + time.Now().Format("20060102150405") + filepath.Ext(file.Filename)
		dst = filepath.Join(uploadDir, fullFilename)
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "upload file err:"+err.Error(), nil)
		return
	}

	if err := service_utils.CompressImage(dst, filepath.Join(compressUploadDir, fullFilename), 400); err != nil {
		utils.SendJsonResponse(c, http.StatusBadRequest, "compress file err:"+err.Error(), nil)
		return
	}

	utils.SendJsonResponse(c, http.StatusOK, "success", gin.H{
		"name":   fullFilename,
		"url":    fmt.Sprintf("http://%s:%s/compress_uploads/%s", host, port, file.Filename),
		"rawUrl": fmt.Sprintf("http://%s:%s/uploads/%s", host, port, file.Filename),
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

func DeleteFile(c *gin.Context) {
	filename := c.Params.ByName("filename")

	uploadDir, _, _, compressUploadDir := getServerInfo()
	fmt.Println("filename:", filename, "uploadDir:", uploadDir, "filepath:", filepath.Join(uploadDir, filename))

	rawFilePath := filepath.Join(uploadDir, filename)
	tmpRawFilePath := filepath.Join(uploadDir, "tmp__"+filename)

	compressedFilePath := filepath.Join(compressUploadDir, filename)

	if _, err := os.Stat(rawFilePath); os.IsNotExist(err) {
		utils.SendJsonResponse(c, http.StatusInternalServerError, "file not found", nil)
		return
	}

	if _, err := os.Stat(compressedFilePath); os.IsNotExist(err) {
		utils.SendJsonResponse(c, http.StatusInternalServerError, "file not found", nil)
		return
	}

	// 原文件
	if err := service_utils.CopyFile(rawFilePath, tmpRawFilePath); err != nil {
		utils.SendJsonResponse(c, http.StatusInternalServerError, "delete file err:"+err.Error(), nil)
		return
	}

	if err := os.Remove(filepath.Join(uploadDir, filename)); err != nil {
		utils.SendJsonResponse(c, http.StatusInternalServerError, "delete file err:"+err.Error(), nil)
		return
	}

	//压缩文件
	if err := os.Remove(filepath.Join(compressUploadDir, filename)); err != nil {
		//恢复文件
		service_utils.CopyFile(tmpRawFilePath, rawFilePath)

		utils.SendJsonResponse(c, http.StatusInternalServerError, "delete file err:"+err.Error(), nil)
		return
	}

	// 删除临时文件
	err := os.Remove(tmpRawFilePath)
	if err != nil {
		utils.SendJsonResponse(c, http.StatusInternalServerError, "delete file err:"+err.Error(), nil)
		return
	}

	utils.SendJsonResponse(c, http.StatusOK, "success", nil)
}

// 同步压缩文件和原文件(根据原文件整理压缩文件列表)
func Sync(c *gin.Context) {
	uploadDir, _, _, compressUploadDir := getServerInfo()

	compressedFiles, err := ioutil.ReadDir(compressUploadDir)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "upload dir not found",
		})
		return
	}

	for _, file := range compressedFiles {
		if !service_utils.FileExists(filepath.Join(uploadDir, file.Name())) {
			err := os.Remove(filepath.Join(compressUploadDir, file.Name()))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed",
				})
				return
			}
		}
	}

}
