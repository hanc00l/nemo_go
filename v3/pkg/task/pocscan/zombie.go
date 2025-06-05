package pocscan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
)

type Zombie struct {
	Config  execute.PocscanConfig
	IsProxy bool
}

type Task struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
	Scheme   string `json:"scheme"`
	//Param    map[string]string  `json:"-"`
	//Mod      TaskMod            `json:"-"`
	//Timeout  int                `json:"-"`
	//Context  context.Context    `json:"-"`
	//Canceler context.CancelFunc `json:"-"`
	//Locker   *sync.Mutex        `json:"-"`
}
type ZombieResult struct {
	*Task
	//Vulns common.Vulns
	//Extracteds parsers.Extracteds
	OK  bool
	Err error
}

func (z Zombie) IsExecuteFromCmd() bool {
	return true
}

func (z Zombie) GetExecuteCmd() string {
	return filepath.Join(conf.GetAbsRootPath(), "thirdparty/zombie", utils.GetThirdpartyBinNameByPlatform(utils.Zombie))
}

func (z Zombie) GetDir() string {
	return ""
}

func (z Zombie) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	cmdArgs = append(
		cmdArgs,
		"--quiet",
		"-I", inputTempFile, "-f", outputTempFile, "-O", "json",
		"--no-unauth", "--no-honeypot", // 减少误报
	)
	// 暂不支持proxy
	return
}

func (z Zombie) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.ZombieCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Zombie),
	})
	return
}

func (z Zombie) Run(target []string) (result Result) {
	//TODO implement me
	panic("implement me")
}

func (z Zombie) ParseContentResult(content []byte) (result Result) {
	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var zr ZombieResult
		err := json.Unmarshal(line, &zr)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			continue
		}
		if !zr.OK || len(zr.IP) == 0 || len(zr.IP) == 0 {
			continue
		}
		result.VulResult = append(result.VulResult, db.VulDocument{
			Authority: fmt.Sprintf("%s:%s", zr.IP, zr.Port),
			Host:      zr.IP,
			Url:       fmt.Sprintf("%s://%s:%s", zr.Scheme, zr.IP, zr.Port),
			PocFile:   fmt.Sprintf("%s弱口令 %s:%s", zr.Service, zr.Username, zr.Password),
			Source:    "zombie",
			Name:      fmt.Sprintf("%s弱口令", zr.Service),
			Severity:  "high",
			Extra:     string(line),
		})
	}
	return
}

func (z Zombie) LoadPocFiles() (pocFiles []string) {
	//TODO implement me
	panic("implement me")
}
