package pocscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"sync"
	"time"
)

type Result struct {
	sync.RWMutex
	VulResult []db.VulDocument
}
type Info struct {
	Name     string                 `json:"name,omitempty"`
	Severity string                 `json:"severity,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
type NucleiJSONResult struct {
	// Template is the relative filename for the template
	Template string `json:"template,omitempty"`
	// TemplateURL is the URL of the template for the result inside the nuclei
	// templates repository if it belongs to the repository.
	TemplateURL string `json:"template-url,omitempty"`
	// TemplateID is the ID of the template for the result.
	TemplateID string `json:"template-id"`
	// MatcherName is the name of the matcher matched if any.
	MatcherName string `json:"matcher-name,omitempty"`
	// Info contains the information about the template.
	Info Info `json:"info,omitempty"`
	// Type is the type of the result event.
	Type string `json:"type"`
	// Host is the host input on which match was found.
	Host string `json:"host,omitempty"`
	// Port is port of the host input on which match was found (if applicable).
	Port string `json:"port,omitempty"`
	// Scheme is the scheme of the host input on which match was found (if applicable).
	Scheme string `json:"scheme,omitempty"`
	// URL is the Base URL of the host input on which match was found (if applicable).
	URL string `json:"url,omitempty"`
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

func ParseResult(config execute.ExecutorTaskInfo, result *Result) (docs []db.VulDocument) {
	for _, vul := range result.VulResult {
		doc := vul
		doc.TaskId = config.MainTaskId
		docs = append(docs, doc)
	}
	return
}
