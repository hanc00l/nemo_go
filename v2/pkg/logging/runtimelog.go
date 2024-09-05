package logging

import (
	"github.com/sirupsen/logrus"
	"io"
)

type RuntimeLogMessage struct {
	Source  string `json:"source"`
	File    string `json:"file"`
	Func    string `json:"func"`
	Level   string `json:"level"`
	Message string `json:"msg"`
}

type RuntimeLogWriter struct {
}

type RuntimeLogHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (w RuntimeLogWriter) Write(p []byte) (n int, err error) {
	if RuntimeLogChan != nil {
		RuntimeLogChan <- p
	}
	return len(p), nil
}

func (hook *RuntimeLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.Bytes()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(line)
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *RuntimeLogHook) Levels() []logrus.Level {
	return hook.LogLevels
}
