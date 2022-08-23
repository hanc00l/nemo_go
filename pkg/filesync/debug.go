package filesync

import (
	"fmt"
	"runtime"
)

var DebugFlag bool

// 用于debug时输出文件号及行号
func DebugInfor(a ...interface{}) {
	if DebugFlag {
		getInfor(2, a...)
	}
}

// 输出信息
func PrintInfor(a ...interface{}) {
	// getInfor(2, a...) // 输出带行号的信息, 有性能损失, 还是直接用fmt吧

	// 仅输出必要信息, 不带文件名和行号
	for _, v := range a {
		fmt.Printf("%v", v)
	}
	fmt.Println()
}

func getInfor(depth int, a ...interface{}) {
	var file string
	var line int
	var ok bool
	_, file, line, ok = runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	fmt.Printf("* %s:%d: ", file, line)
	for _, v := range a {
		fmt.Printf("%v", v)
	}
	fmt.Println()
}
