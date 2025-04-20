package core

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"strings"
)

type Honeypot struct {
	ipAndDomainMap map[string]struct{}
}

func NewHoneypot(workspaceId string) *Honeypot {
	h := Honeypot{
		ipAndDomainMap: make(map[string]struct{}),
	}
	if len(workspaceId) > 0 {
		h.loadHoneypot(workspaceId)
	}
	return &h
}

func (h *Honeypot) loadHoneypot(workspaceId string) bool {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	defer db.CloseClient(mongoClient)

	cd := db.NewCustomData(workspaceId, mongoClient)
	docs, _ := cd.Find(db.CategoryHoneypot)
	for _, doc := range docs {
		lines := strings.Split(doc.Data, "\n")
		for _, data := range lines {
			ipOrDomain := strings.TrimSpace(data)
			if strings.HasPrefix(ipOrDomain, "#") {
				continue
			}
			domainArray := strings.Split(ipOrDomain, " ")
			if len(domainArray) > 1 {
				ipOrDomain = domainArray[0]
			}
			h.ipAndDomainMap[ipOrDomain] = struct{}{}
		}
	}
	return true
}

func (h *Honeypot) IsHoneypot(ipOrDomain string) bool {
	_, ok := h.ipAndDomainMap[ipOrDomain]
	return ok
}
