package custom

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	nmapServiceData   map[string]string
	customServiceData map[string]string
}

const UnknownService = "unknown"

// NewService 创建service对象
func NewService() Service {
	s := Service{
		nmapServiceData:   make(map[string]string),
		customServiceData: make(map[string]string),
	}
	s.loadNmapService()
	s.loadCustomService()

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

// loadCustomService 加载自定义service定义文件
func (s *Service) loadCustomService() {
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/services-custom.txt"))
	if err != nil {
		logging.RuntimeLog.Info(err)
		logging.CLILog.Info(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			servicesListArray := strings.Split(txt, " ")
			if len(servicesListArray) < 2 {
				continue
			}

			s.customServiceData[strings.TrimSpace(servicesListArray[0])] = strings.TrimSpace(servicesListArray[1])
		}
	}
}

// FindService 查找端口服务
// 如果给定了IP，并且IP是自定义归属地IP就先查找自定义端口服务
// 否则从nmap定义的服务类型中进行查找
func (s *Service) FindService(portNumber int, ip string) string {
	var serviceName string
	var ok bool
	if ip != "" {
		ipl := &Ipv4Location{}
		ipLocation := ipl.FindCustomIP(ip)
		if ipLocation != "" {
			serviceName, ok = s.customServiceData[fmt.Sprintf("%d/tcp", portNumber)]
			if ok && serviceName != "" {
				return serviceName
			}
		}
	}
	serviceName, ok = s.nmapServiceData[fmt.Sprintf("%d/tcp", portNumber)]
	if ok && serviceName != "" {
		return serviceName
	}

	return UnknownService
}
