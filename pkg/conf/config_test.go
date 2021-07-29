package conf

import (
	"testing"
)

func TestConfig_LoadYamlConfig(t *testing.T) {
	t.Log(Nemo)
	t.Log(Nemo.API.ICP)
}

func TestConfig_WriteYamlConfig(t *testing.T) {
	Nemo.ReloadConfig()
	Nemo.WriteConfig()
}
