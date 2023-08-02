package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
	"strings"
)

func LoadCustomTaskWorkspace() (workspace map[string]struct{}) {
	workspace = make(map[string]struct{})
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/task_workspace.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		guidAndName := strings.Split(text, " ")
		guid := strings.TrimSpace(guidAndName[0])
		workspace[guid] = struct{}{}
	}
	inputFile.Close()

	return
}
