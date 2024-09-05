package logging

import (
	"fmt"
	"testing"
	"time"
)

func Test3(t *testing.T) {
	RuntimeLog.Error("Error")
	RuntimeLog.Info("info")

	CLILog.Info("cli hello")
	CLILog.Errorf("name %s", "hacker")

}

func Test4(t *testing.T) {
	RuntimeLogChan = make(chan []byte, RuntimeLogChanMax)
	quitChan := make(chan int)

	go outputLog(RuntimeLogChan, quitChan)

	RuntimeLog.Error("Runtime Error")
	RuntimeLog.Info("Runtime info")
	CLILog.Info("cli hello")
	CLILog.Errorf("name %s", "hacker")

	time.Sleep(3 * time.Second)
	quitChan <- 1
}

func outputLog(c chan []byte, quit chan int) {
	for {
		select {
		case msg := <-c:
			fmt.Println("received msg:", string(msg))
		case <-quit:
			return
		}
	}
}
