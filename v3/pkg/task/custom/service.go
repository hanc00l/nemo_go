package custom

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	nmapServiceData map[string]string
}

const UnknownService = "unknown"

// NewService 创建service对象
func NewService() Service {
	s := Service{
		nmapServiceData: make(map[string]string),
	}
	s.loadNmapService()

	return s
}

// loadNmapService 加载nmap的service定义
func (s *Service) loadNmapService() {
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/nmap/nmap-services"))
	if err != nil {
		logging.RuntimeLog.Info(err)
		logging.CLILog.Info(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			servicesListArray := strings.Split(txt, "\t")
			if len(servicesListArray) < 3 {
				continue
			}
			s.nmapServiceData[servicesListArray[1]] = servicesListArray[0]
		}
	}
}

// FindService 查找端口服务
// 否则从nmap定义的服务类型中进行查找
func (s *Service) FindService(portNumber int) string {
	var serviceName string
	var ok bool
	serviceName, ok = s.nmapServiceData[fmt.Sprintf("%d/tcp", portNumber)]
	if ok && serviceName != "" {
		return serviceName
	}

	return UnknownService
}
