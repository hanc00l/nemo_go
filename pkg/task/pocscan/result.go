package pocscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"sync"
	"time"
)

const (
	pocMaxThreadNumber      = 2
	ipMaxPocForHoneypot     = 40
	portMaxPocForHoneypot   = 10
	domainMaxPocForHoneypot = 40
)

type Config struct {
	Target           string `json:"target"`
	PocFile          string `json:"pocFile"`
	CmdBin           string `json:"cmdBin"`
	IsLoadOpenedPort bool   `json:"loadOpenedPort"`
}

type Result struct {
	Target  string `json:"target"`
	Url     string `json:"url"`
	PocFile string `json:"pocFile"`
	Source  string `json:"source"`
	Extra   string `json:"extra"`
}

type xrayJSONResult struct {
	Target struct {
		Url string `json:"url"`
	} `json:"target"`
	Plugin string `json:"plugin"`
	Detail struct {
		Addr     string     `json:"addr"`
		Payload  string     `json:"payload"`
		Snapshot [][]string `json:"snapshot"`
	} `json:"detail"`
}
type nucleiJSONResult struct {
	// Template is the relative filename for the template
	Template string `json:"template,omitempty"`
	// TemplateURL is the URL of the template for the result inside the nuclei
	// templates repository if it belongs to the repository.
	TemplateURL string `json:"template-url,omitempty"`
	// TemplateID is the ID of the template for the result.
	TemplateID string `json:"template-id"`
	// MatcherName is the name of the matcher matched if any.
	MatcherName string `json:"matcher-name,omitempty"`
	// Type is the type of the result event.
	Type string `json:"type"`
	// Host is the host input on which match was found.
	Host string `json:"host,omitempty"`
	// Path is the path input on which match was found.
	Path string `json:"path,omitempty"`
	// Matched contains the matched input in its transformed form.
	Matched string `json:"matched-at,omitempty"`
	// Request is the optional, dumped request for the match.
	Request string `json:"request,omitempty"`
	// Response is the optional, dumped response for the match.
	Response string `json:"response,omitempty"`
	// IP is the IP address for the found result event.
	IP string `json:"ip,omitempty"`
	// Timestamp is the time the result was found at.
	Timestamp time.Time `json:"timestamp"`
	// Interaction is the full details of interactsh interaction.
}

// PortResult 端口结果
type PortResult struct {
	Vuls []string
}

// IPResult IP的端口结果
type IPResult struct {
	Ports map[int]*PortResult
}

// PortscanVulResult ip结果
type PortscanVulResult struct {
	sync.RWMutex `json:"-"`
	IPResult     map[string]*IPResult
}

func (r *PortscanVulResult) HasIP(ip string) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.IPResult[ip]
	return ok
}

func (r *PortscanVulResult) SetIP(ip string) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip] = &IPResult{Ports: make(map[int]*PortResult)}
}

func (r *PortscanVulResult) HasPort(ip string, port int) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.IPResult[ip].Ports[port]
	return ok
}

func (r *PortscanVulResult) SetPort(ip string, port int) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip].Ports[port] = &PortResult{}
}

func (r *PortscanVulResult) SetPortVul(ip string, port int, vul string) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip].Ports[port].Vuls = append(r.IPResult[ip].Ports[port].Vuls, vul)
}

// DomainResult 域名结果
type DomainResult struct {
	Vuls []string
}

// DomainscanVulResult 域名结果
type DomainscanVulResult struct {
	sync.RWMutex `json:"-"`
	DomainResult map[string]*DomainResult
}

func (r *DomainscanVulResult) HasDomain(domain string) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.DomainResult[domain]
	return ok
}

func (r *DomainscanVulResult) SetDomain(domain string) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain] = &DomainResult{}
}

func (r *DomainscanVulResult) SetDomainVul(domain string, vul string) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].Vuls = append(r.DomainResult[domain].Vuls, vul)
}

// SaveResult 保存结果
func SaveResult(result []Result) string {
	var resultCount int
	for _, r := range result {
		target := utils.HostStrip(r.Target)
		extra := r.Extra
		if len(r.Extra) > 2000 {
			extra = r.Extra[:2000] + "..."
		}
		vul := db.Vulnerability{
			Target:  target,
			Url:     r.Url,
			PocFile: r.PocFile,
			Source:  r.Source,
			Extra:   extra,
		}
		if vul.SaveOrUpdate() {
			resultCount++
		}
	}
	return fmt.Sprintf("vulnerability:%d", resultCount)
}
