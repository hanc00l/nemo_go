package server

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/constant"
	"github.com/hanc00l/nemo_go/v2/pkg/minichat/util"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// IsImageExt 判断是否是图片后缀
func IsImageExt(filename string) bool {
	ext := filepath.Ext(filename)

	return strings.EqualFold(ext, ".jpg") ||
		strings.EqualFold(ext, ".jpeg") ||
		strings.EqualFold(ext, ".png") ||
		strings.EqualFold(ext, ".gif") ||
		strings.EqualFold(ext, ".svg") ||
		strings.EqualFold(ext, ".bmp") ||
		strings.EqualFold(ext, ".webp")
}

type onlyWriter struct {
	io.Writer
}

var (
	copyBufferPool sync.Pool
)

func Upload(w http.ResponseWriter, r *http.Request) {
	//query := r.URL.Query()
	//roomNumber := query.Get("room_number")
	//userName := query.Get("username")
	//password := query.Get("password")

	// 确保请求方法是POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	values := r.URL.Query()
	roomNumber := values.Get("roomNumber")
	if roomNumber == "" {
		http.Error(w, "Invalid roomNumber", http.StatusBadRequest)
		return
	}
	identify := util.Md5Crypt(roomNumber)

	if r.ParseForm() == nil {
		formFile, moreFile, err := r.FormFile("editormd-image-file")
		query := r.URL.Query()
		if query.Get("identify") == "headimg" {
			identify = "headimg"
			if !IsImageExt(moreFile.Filename) {
				http.Error(w, "文件上传失败,请上传图片格式文件", http.StatusInternalServerError)
				return
			}
		}
		if err != nil {
			http.Error(w, "message.upload_file_size_limit", http.StatusBadRequest)
			return
		}
		//formFile, fileHeader, err = r.FormFile("file")
		if formFile != nil {
			if err != nil {
				log.Println(err)
			}
			if moreFile.Size > 50*1024*1024 {
				http.Error(w, "Upload File Size Limit < 50M", http.StatusBadRequest)
				return
			}
			ext := filepath.Ext(moreFile.Filename)
			fileName := "m_" + util.UniqueId() + "_r"
			// 上传文件的路径
			filePath := filepath.Join(constant.UploadSavePath, identify)
			//将图片和文件分开存放
			if IsImageExt(moreFile.Filename) {
				filePath = filepath.Join(filePath, "images", fileName+ext)
			} else {
				filePath = filepath.Join(filePath, "files", fileName+ext)
			}
			path := filepath.Dir(filePath)

			_ = os.MkdirAll(path, os.ModePerm)

			src, _, err := r.FormFile("editormd-image-file")
			if err != nil {
				log.Printf("保存文件失败 -> ", err)
				http.Error(w, "File Save Failed", http.StatusBadRequest)
				return
			}

			defer src.Close()

			dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
			if err != nil {
				log.Printf("保存文件失败 -> ", err)
				http.Error(w, "File Save Failed", http.StatusBadRequest)
				return
			}
			defer dst.Close()
			_, err = io.Copy(dst, formFile)
			if err != nil {
				http.Error(w, "文件上传失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err != nil {
				log.Printf("保存文件失败 -> ", err)
				http.Error(w, "File Save Failed", http.StatusBadRequest)
				return
			}
			var HttpPath string
			if IsImageExt(moreFile.Filename) {
				HttpPath = filepath.Join(constant.UploadHttpPath, identify, "images", fileName+ext)
			} else {
				HttpPath = filepath.Join(constant.UploadHttpPath, identify, "files", fileName+ext)
			}
			result := map[string]interface{}{
				"errcode":   0,
				"success":   1,
				"message":   "ok",
				"url":       HttpPath,
				"alt":       "",
				"is_attach": false,
				"attach":    "",
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			var content []byte
			content, err = json.MarshalIndent(result, "", "  ")
			if err != nil {
				http.Error(w, "JSON Parse Error", http.StatusBadRequest)
				return
			}
			w.Write(content)
			return
		}
	}

	http.Error(w, "No File Found", http.StatusBadRequest)
	return
}
