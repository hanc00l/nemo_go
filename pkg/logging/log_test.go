package logging

import (
	"testing"
)


func Test3(t *testing.T){
	RuntimeLog.Error("Error")
	RuntimeLog.Info("info")

	CLILog.Info("cli hello")
	CLILog.Errorf("name %s","hacker")
}
