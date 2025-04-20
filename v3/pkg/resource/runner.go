package resource

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"os"
	"path/filepath"
)

func LoadFileResource(category string, name string) (*Resource, error) {
	var resource Resource
	var ok bool
	if resource, ok = Resources[category][name]; !ok {
		return nil, fmt.Errorf("resource %s:%s not found", category, name)
	}

	// 读取文件内容
	filePath := filepath.Join(conf.GetRootPath(), resource.Path, resource.Name)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 更新资源内容
	resource.Bytes = bytes
	resource.Hash = utils.MD5Bytes(bytes)

	return &resource, nil
}

func SaveFileResource(resource *Resource) (err error) {
	// 创建目录
	if !utils.MakePath(filepath.Join(conf.GetRootPath(), resource.Path)) {
		return fmt.Errorf("can not save resource %s", resource.Path)
	}
	// 保存文件
	filePath := filepath.Join(conf.GetRootPath(), resource.Path, resource.Name)
	fileMode := 0644
	if resource.Type == ExecuteFile {
		fileMode = 0755
	}
	err = os.WriteFile(filePath, resource.Bytes, os.FileMode(fileMode))
	if err != nil {
		return err
	}
	// 检验文件是否存在和MD5是否一致
	if !utils.CheckFileExist(filePath) {
		return fmt.Errorf("resource %s save error, file not exist", resource.Name)
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if utils.MD5Bytes(fileBytes) != resource.Hash {
		return fmt.Errorf("save resource %s error, hash not match", resource.Name)
	}
	return nil
}

func LoadDirResource(category string, name string) (*Resource, error) {
	var resource Resource
	var ok bool
	if resource, ok = Resources[category][name]; !ok {
		return nil, fmt.Errorf("resource %s:%s not found", category, name)
	}
	// 遍历目录，收集文件和目录信息
	resources, err := traverseDirectory(filepath.Join(conf.GetRootPath(), resource.Path, resource.Name))
	if err != nil {
		return nil, err
	}
	resource.Bytes, err = json.Marshal(resources)
	if err != nil {
		return nil, err
	}
	resource.Hash = utils.MD5Bytes(resource.Bytes)

	return &resource, nil

}

func SaveDirResource(resource *Resource) (err error) {
	// 创建目录
	if !utils.MakePath(filepath.Join(conf.GetRootPath(), resource.Path)) {
		return fmt.Errorf("can not save resource %s", resource.Path)
	}
	// 保存文件
	var resources []Resource
	err = json.Unmarshal(resource.Bytes, &resources)
	if err != nil {
		return err
	}
	for _, r := range resources {
		if !utils.MakePath(filepath.Join(conf.GetRootPath(), resource.Path, r.Path)) {
			return fmt.Errorf("can not save resource %s", r.Path)
		}
		filePath := filepath.Join(conf.GetRootPath(), resource.Path, r.Path, r.Name)
		fileMode := 0644
		if resource.Type == ExecuteFile {
			fileMode = 0755
		}
		err = os.WriteFile(filePath, r.Bytes, os.FileMode(fileMode))
		if err != nil {
			return err
		}
	}
	// 检验文件是否存在和MD5是否一致
	for _, r := range resources {
		filePath := filepath.Join(conf.GetRootPath(), resource.Path, r.Path, r.Name)
		if !utils.CheckFileExist(filePath) {
			return fmt.Errorf("resource %s save error, file not exist", r.Name)
		}
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if utils.MD5Bytes(fileBytes) != r.Hash {
			return fmt.Errorf("save resource %s error, hash not match", r.Name)
		}
	}
	return nil
}

// traverseDirectory 递归遍历目录并收集文件和目录信息
func traverseDirectory(rootPath string) ([]Resource, error) {
	var resources []Resource

	// 获取 rootPath 的绝对路径，用于计算相对路径
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 忽略目录本身，只处理文件
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// 计算相对路径
			relPath, err := filepath.Rel(absRootPath, path)
			if err != nil {
				return err
			}
			resource := Resource{
				Name:  info.Name(),
				Path:  filepath.Dir(relPath),
				Bytes: content,
				Hash:  utils.MD5Bytes(content),
			}
			resources = append(resources, resource)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}
