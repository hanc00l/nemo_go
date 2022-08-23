package filesync

import (
	"os"
	"path/filepath"
)

// Traverse walk要同步的文件, 生成md5, 并返回列表
func Traverse(path string) ([]string, error) {
	f, fErr := os.Lstat(path)
	if fErr != nil {
		return nil, fErr
	}
	var dir string
	var base string
	if f.IsDir() {
		dir = path
		base = "."
	} else {
		dir = filepath.Dir(path)
		base = filepath.Base(path)
	}
	// 相对目录遍历，需要cd到指定目录下；遍历完后必须返回原来的当前位置
	fErr = os.Chdir(dir)
	if fErr != nil {
		return nil, fErr
	}
	defer os.Chdir(cwd)

	md5List := make([]string, 10)
	var md5Str string
	WalkFunc := func(path string, info os.FileInfo, err error) error {
		if checkFileIsSyncWhileList(path) == false {
			return nil
		}
		if info.IsDir() {
			md5Str = path + ",,Directory"
			md5List = append(md5List, md5Str)
		}
		if info.Mode().IsRegular() {
			// if !info.IsDir() {
			md5Str, fErr = Md5OfAFile(path)
			if fErr != nil {
				PrintInfor(fErr)
				return fErr
			}
			md5Str = path + ",," + md5Str
			md5List = append(md5List, md5Str)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			filename, err := os.Readlink(path)
			if err != nil {
				return err
			}
			md5Str = "symbolLink&&" + filename
			md5Str = path + ",," + md5Str
			md5List = append(md5List, md5Str)
		}
		return nil
	}
	fErr = filepath.Walk(base, WalkFunc)
	if fErr != nil {
		PrintInfor(fErr)
		return nil, fErr
	}
	// DebugInfor(md5List)
	for i, v := range md5List {
		// DebugInfor(v)
		if v == "" || v == ".,,Directory" {
			continue
		}
		md5List = md5List[i:]
		break
	}
	if md5List[len(md5List)-1] == ".,,Directory" {
		return []string{}, nil
	}
	return md5List, nil
}
