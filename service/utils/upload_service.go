package utils

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func CompressImage(inputFile, outputFile string, newWidth uint) error {
	// 打开输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return fmt.Errorf("不支持的文件扩展名: %s", ext)
	}

	var img image.Image

	if ext == ".png" {
		img, err = png.Decode(file)
		if err != nil {
			return fmt.Errorf("解码图片失败: %v", err)
		}
	}

	if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
		if err != nil {
			return fmt.Errorf("解码图片失败: %v", err)
		}
	}

	if ext == ".gif" {
		img, err = gif.Decode(file)
		if err != nil {
			return fmt.Errorf("解码图片失败: %v", err)
		}
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
