package utils

import (
	"archive/zip"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetRandomString2 生成指定长度的随机字符串
func GetRandomString2(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

// GetTempPathFileName 获取一个临时文件名
func GetTempPathFileName() (pathFileName string) {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.tmp", GetRandomString2(16)))
}

// GetTempPNGPathFileName 获取一个临时文件名，后缀为PNG
func GetTempPNGPathFileName() (pathFileName string) {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.png", GetRandomString2(16)))
}


//Unzip 解压zip文件
func Unzip(archiveFile, dstPath string) error {
	reader, err := zip.OpenReader(archiveFile)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return errors.WithStack(err)
	}
	for _, file := range reader.File {
		unzipped := filepath.Join(dstPath, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(unzipped, file.Mode())
			if err != nil {
				return errors.WithStack(err)
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return errors.WithStack(err)
		}

		targetFile, err := os.OpenFile(unzipped, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			fileReader.Close()
			return errors.WithStack(err)
		}

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			fileReader.Close()
			targetFile.Close()
			return errors.WithStack(err)
		}

		fileReader.Close()
		targetFile.Close()
	}

	return nil
}

// DownloadFile 下载文件
func DownloadFile(url, dstPathFile string) (bool, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// 创建一个文件用于保存
	out, err := os.Create(dstPathFile)
	if err != nil {
		return false, err
	}
	defer out.Close()
	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}
	return true, nil
}

// MakePath 创建目录，如果目录存在则直接返回
func MakePath(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		if err = os.Mkdir(filepath, 0777); err == nil {
			return true
		}
	}
	return false
}

// ReSizePicture 对图片文件尺寸缩放
func ReSizePicture(srcFile, dstFile string, width, height int) bool {
	src, err := imaging.Open(srcFile)
	if err != nil {
		logging.RuntimeLog.Errorf("failed to open image: %v", err)
		return false
	}
	dst := imaging.Resize(src, width, height, imaging.CatmullRom)
	err = imaging.Save(dst, dstFile)
	if err != nil {
		logging.RuntimeLog.Errorf("failed to save image: %v", err)
		return false
	}
	return true
}
