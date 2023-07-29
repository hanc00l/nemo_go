package logging

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"strings"
)

const RuntimeLogChanMax = 100

var (
	//RuntimeLog 系统运行时日志，记录发生的异常和错误
	RuntimeLog = logrus.New()
	// CLILog 控制台日志
	CLILog = logrus.New()
	// RuntimeLogChan 系统运行日志的chan，用于将日志通过chan的方式发送到其它地方
	RuntimeLogChan chan []byte
)

func init() {
	RuntimeLog.SetLevel(logrus.InfoLevel)
	RuntimeLog.SetReportCaller(true)
	RuntimeLog.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:03:04",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File)
			functions := strings.Split(frame.Function, "/")

			return functions[len(functions)-1], fileName
		},
	})

	if file, err := os.OpenFile(path.Join(conf.GetRootPath(), "log/runtime.log"),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666); err == nil {
		RuntimeLog.SetOutput(file)
	}

	RuntimeLog.AddHook(&RuntimeLogHook{
		Writer: RuntimeLogWriter{},
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		},
	})

	CLILog.SetFormatter(GetCustomLoggerFormatter())
}

// GetCustomLoggerFormatter 日定义日志格式
func GetCustomLoggerFormatter() logrus.Formatter {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.FullTimestamp = true                    // 显示完整时间
	customFormatter.TimestampFormat = "2006-01-02 15:04:05" // 时间格式
	customFormatter.DisableTimestamp = false                // 禁止显示时间
	customFormatter.DisableColors = false                   // 禁止颜色显示

	return customFormatter
}
