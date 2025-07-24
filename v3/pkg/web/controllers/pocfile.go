package controllers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type PocFileController struct {
	BaseController
}

const (
	allowedBaseDir     = "thirdparty/nuclei"                        // 允许管理的基准目录
	allowedUploadTypes = ".yaml,.yml,.json,.txt,.xml,.csv,.js,.svg" // 允许上传的文件类型
	maxUploadSize      = 10 << 20                                   // 10MB
)

// FileInfo 文件信息结构体
type FileInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	IsDir     bool   `json:"is_dir"`
	Extension string `json:"extension"`
	MD5       string `json:"md5"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	StatusResponseData
	Data []FileInfo `json:"data"`
}

// FileContentResponse 文件内容响应
type FileContentResponse struct {
	StatusResponseData
	Content string `json:"content"`
}

// sanitizePath 检查并清理路径，防止目录穿越
func sanitizePath(path string) (string, error) {
	// 清理路径中的..和.
	cleanPath := filepath.Clean(path)
	allowedAbsDir := filepath.Join(conf.GetAbsRootPath(), allowedBaseDir)
	// 转换为绝对路径
	absPath, err := filepath.Abs(filepath.Join(allowedAbsDir, cleanPath))
	if err != nil {
		return "", err
	}

	// 检查路径是否在允许的基准目录下
	if !strings.HasPrefix(absPath, allowedAbsDir) {
		return "", os.ErrPermission
	}

	return absPath, nil
}

// checkFileType 检查文件类型是否允许
func checkFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := strings.Split(allowedUploadTypes, ",")

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

func (c *PocFileController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "pocfile-list.html"
}

// ListFilesAction 列出目录下的文件
func (c *PocFileController) ListFilesAction() {
	defer func(c *PocFileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	path := c.GetString("path", "")
	sanitizedPath, err := sanitizePath(path)
	if err != nil {
		c.FailedStatus("路径不合法")
		return
	}

	fileInfos := make([]FileInfo, 0)

	files, err := os.ReadDir(sanitizedPath)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		absPath := filepath.Join(sanitizedPath, file.Name())

		var md5Str string
		if !file.IsDir() {
			md5Str, _ = c.calculateFileMD5(absPath)
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:      file.Name(),
			Path:      filePath,
			Size:      info.Size(),
			ModTime:   FormatDateTime(info.ModTime()),
			IsDir:     file.IsDir(),
			Extension: strings.ToLower(filepath.Ext(file.Name())),
			MD5:       md5Str,
		})
	}

	c.Data["json"] = FileListResponse{
		StatusResponseData: StatusResponseData{Status: Success},
		Data:               fileInfos,
	}

}

// calculateFileMD5 计算文件MD5
func (c *PocFileController) calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// UploadFileAction 上传文件
func (c *PocFileController) UploadFileAction() {
	defer func(c *PocFileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	path := c.GetString("path", "")
	sanitizedPath, err := sanitizePath(path)
	if err != nil {
		c.FailedStatus("路径不合法")
		return
	}

	// 检查是否是目录
	fileInfo, err := os.Stat(sanitizedPath)
	if err != nil || !fileInfo.IsDir() {
		c.FailedStatus("目标路径不是目录或不存在")
		return
	}

	// 解析多文件上传
	files, err := c.GetFiles("files")
	if err != nil {
		c.FailedStatus("获取上传文件失败")
		return
	}

	for _, fileHeader := range files {
		// 检查文件名安全性
		if !IsSafeFilename(fileHeader.Filename) {
			c.FailedStatus("文件名包含非法字符")
			return
		}
		// 检查文件大小
		if fileHeader.Size > maxUploadSize {
			c.FailedStatus("文件大小超过限制")
			return
		}

		// 检查文件类型
		if !checkFileType(fileHeader.Filename) {
			c.FailedStatus("不允许上传此类型文件")
			return
		}

		// 保存文件
		// 使用清理后的文件名
		safeName := SanitizeFilename(fileHeader.Filename)
		dstPath := filepath.Join(sanitizedPath, safeName)
		err := c.SaveToFile("files", dstPath)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}

	c.SucceededStatus("文件上传成功")
}

// DownloadFileAction 下载文件
func (c *PocFileController) DownloadFileAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	path := c.GetString("path", "")
	sanitizedPath, err := sanitizePath(path)
	if err != nil {
		c.FailedStatus("路径不合法")
		return
	}

	fileInfo, err := os.Stat(sanitizedPath)
	if err != nil {
		c.FailedStatus("文件不存在")
		return
	}
	if fileInfo.IsDir() {
		c.FailedStatus("不能下载目录")
		return
	}

	file, err := os.Open(sanitizedPath)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	defer file.Close()

	// 设置下载头
	fileName := filepath.Base(sanitizedPath)
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Ctx.Output.Header("Content-Type", contentType)
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Ctx.Output.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	_, err = io.Copy(c.Ctx.ResponseWriter, file)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
}

// DeleteFileAction 删除文件或目录
func (c *PocFileController) DeleteFileAction() {
	defer func(c *PocFileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	paths := c.GetStrings("paths[]")

	for _, path := range paths {
		sanitizedPath, err := sanitizePath(path)
		if err != nil {
			c.FailedStatus("路径不合法")
			return
		}

		fileInfo, err := os.Stat(sanitizedPath)
		if err != nil {
			c.FailedStatus("文件或目录不存在")
			return
		}

		if fileInfo.IsDir() {
			err = os.RemoveAll(sanitizedPath)
		} else {
			err = os.Remove(sanitizedPath)
		}

		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}

	c.SucceededStatus("删除成功")
}

// SearchFilesAction 搜索文件
func (c *PocFileController) SearchFilesAction() {
	defer func(c *PocFileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	keyword := c.GetString("keyword", "")
	mode := c.GetString("mode", "name") // name or content

	if keyword == "" {
		c.FailedStatus("请输入搜索关键词")
		return
	}

	results := make([]FileInfo, 0)

	allowedAbsBaseDir := filepath.Join(conf.GetAbsRootPath(), allowedBaseDir)

	err := filepath.Walk(allowedAbsBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(allowedAbsBaseDir, path)
		if err != nil {
			return nil
		}

		// 文件名搜索
		if mode == "name" && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(keyword)) {
			md5Str, _ := c.calculateFileMD5(path)
			results = append(results, FileInfo{
				Name:      info.Name(),
				Path:      relPath,
				Size:      info.Size(),
				ModTime:   FormatDateTime(info.ModTime()),
				IsDir:     false,
				Extension: strings.ToLower(filepath.Ext(info.Name())),
				MD5:       md5Str,
			})
			return nil
		}

		// 文件内容搜索
		if mode == "content" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if bytes.Contains(bytes.ToLower(content), bytes.ToLower([]byte(keyword))) {
				md5Str, _ := c.calculateFileMD5(path)
				results = append(results, FileInfo{
					Name:      info.Name(),
					Path:      relPath,
					Size:      info.Size(),
					ModTime:   FormatDateTime(info.ModTime()),
					IsDir:     false,
					Extension: strings.ToLower(filepath.Ext(info.Name())),
					MD5:       md5Str,
				})
			}
		}

		return nil
	})

	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	c.Data["json"] = FileListResponse{
		StatusResponseData: StatusResponseData{Status: Success},
		Data:               results,
	}
}

// CreateFolderAction 创建新目录
func (c *PocFileController) CreateFolderAction() {
	defer func(c *PocFileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	path := c.GetString("path", "")
	name := c.GetString("name", "")

	if name == "" {
		c.FailedStatus("目录名称不能为空")
		return
	}

	// 检查目录名称是否合法
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		c.FailedStatus("目录名称包含非法字符")
		return
	}

	sanitizedPath, err := sanitizePath(path)
	if err != nil {
		c.FailedStatus("路径不合法")
		return
	}

	// 创建完整路径
	fullPath := filepath.Join(sanitizedPath, name)

	// 检查目录是否已存在
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		c.FailedStatus("目录已存在")
		return
	}

	// 创建目录
	err = os.Mkdir(fullPath, 0755)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus("创建目录失败: " + err.Error())
		return
	}

	c.SucceededStatus("目录创建成功")
}

// IsSafeFilename 检查文件名是否安全
func IsSafeFilename(filename string) bool {
	// 空文件名不安全
	if filename == "" {
		return false
	}

	// 检查是否包含路径分隔符
	if strings.ContainsAny(filename, `/\`) {
		return false
	}

	// 检查是否包含特殊字符
	// 禁止: \ / : * ? " < > |
	// 以及控制字符(0x00-0x1F, 0x7F)
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1F\x7F]`)
	if invalidChars.MatchString(filename) {
		return false
	}

	// 检查Windows保留文件名
	windowsReserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	base := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range windowsReserved {
		if base == reserved {
			return false
		}
	}

	// 检查文件名长度限制(255是大多数文件系统的限制)
	if len(filename) > 255 {
		return false
	}

	// 检查以点开头或结尾的文件名
	if strings.HasPrefix(filename, ".") || strings.HasSuffix(filename, ".") {
		return false
	}

	// 检查空格开头或结尾
	if strings.HasPrefix(filename, " ") || strings.HasSuffix(filename, " ") {
		return false
	}

	return true
}

// SanitizeFilename 清理文件名，返回安全的文件名
func SanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// 替换特殊字符为下划线
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1F\x7F]`)
	filename = invalidChars.ReplaceAllString(filename, "_")

	// 处理Windows保留文件名
	windowsReserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	base := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range windowsReserved {
		if base == reserved {
			filename = "_" + filename
			break
		}
	}

	// 限制文件名长度
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		name := filename[:len(filename)-len(ext)]
		if len(ext) > 0 && len(name) > 255-len(ext) {
			name = name[:255-len(ext)]
		}
		filename = name + ext
	}

	// 去除开头和结尾的点
	filename = strings.Trim(filename, ".")

	// 去除开头和结尾的空格
	filename = strings.TrimSpace(filename)

	// 如果处理后为空，返回默认文件名
	if filename == "" {
		return "unnamed_file"
	}

	return filename
}
