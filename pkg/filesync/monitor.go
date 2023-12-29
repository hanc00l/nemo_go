package filesync

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
)

type NotifyFile struct {
	watch            *fsnotify.Watcher
	ChNeedWorkerSync chan string
}

func NewNotifyFile() *NotifyFile {
	w := new(NotifyFile)
	var err error
	w.watch, err = fsnotify.NewWatcher()
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	w.ChNeedWorkerSync = make(chan string)

	return w
}

// WatchDir 监控目录中文件的变化
func (n *NotifyFile) WatchDir(srcPath string) {
	f, fErr := os.Lstat(srcPath)
	if fErr != nil {
		logging.RuntimeLog.Error(fErr)
		logging.CLILog.Error(fErr)
		return
	}
	var dir string
	var base string
	if f.IsDir() {
		dir = srcPath
		base = "."
	} else {
		dir = filepath.Dir(srcPath)
		base = filepath.Base(srcPath)
	}
	fErr = os.Chdir(dir)
	if fErr != nil {
		msg := fmt.Sprintf("traverse monitor dir failure:%s", fErr.Error())
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)
		return
	}
	defer os.Chdir(cwd)
	var err error
	//通过Walk来遍历目录下的所有子目录
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if !checkFileIsMonitorWhileList(path) {
			return nil
		}
		//判断是否为目录，监控目录,目录下文件也在监控范围内，不需要加
		if info.IsDir() {
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			err = n.watch.Add(path)
			if err != nil {
				return err
			}
			//fmt.Println("监控 : ", path)
		}
		return nil
	})
	if err != nil {
		msg := fmt.Sprintf("traverse monitor dir failure:%s", err.Error())
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)
		return
	}

	go n.WatchEvent()
}

// WatchFile 监控指定文件的变化
func (n *NotifyFile) WatchFile(srcPathFile []string) {
	for _, f := range srcPathFile {
		_, fErr := os.Lstat(f)
		if fErr != nil {
			logging.RuntimeLog.Error(fErr)
			logging.CLILog.Error(fErr)
			continue
		}
		err := n.watch.Add(f)
		if err != nil {
			logging.RuntimeLog.Error(fErr)
			logging.CLILog.Error(fErr)
			continue
		}
	}
	go n.WatchEvent()
}

// WatchEvent 文件变化事件通知
func (n *NotifyFile) WatchEvent() {
	for {
		select {
		case ev := <-n.watch.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					//fmt.Println("创建文件 : ", ev.Name)
					//获取新创建文件的信息，如果是目录，则加入监控中
					file, err := os.Stat(ev.Name)
					if err == nil && file.IsDir() {
						n.watch.Add(ev.Name)
						//fmt.Println("添加监控 : ", ev.Name)
					}
					n.ChNeedWorkerSync <- ev.Name
				}

				if ev.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Println("写入文件 : ", ev.Name)
					n.ChNeedWorkerSync <- ev.Name
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					//fmt.Println("删除文件 : ", ev.Name)
					//如果删除文件是目录，则移除监控
					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						n.watch.Remove(ev.Name)
						//fmt.Println("删除监控 : ", ev.Name)
					}
				}

				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					//如果重命名文件是目录，则移除监控 ,注意这里无法使用os.Stat来判断是否是目录了
					//因为重命名后，go已经无法找到原文件来获取信息了,所以简单粗爆直接remove
					//fmt.Println("重命名文件 : ", ev.Name)
					n.watch.Remove(ev.Name)
					n.ChNeedWorkerSync <- ev.Name
				}
				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					//fmt.Println("修改权限 : ", ev.Name)
				}
			}
		case err := <-n.watch.Errors:
			{
				logging.CLILog.Errorf("moniter error : %v", err)
				logging.RuntimeLog.Errorf("moniter error : %v", err)
				return
			}
		}
	}
}
