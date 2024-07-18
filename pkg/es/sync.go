package es

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"strconv"
	"strings"
)

func getOrganization(workspaceId int) (orgResult map[int]string) {
	org := db.Organization{}
	searchMap := make(map[string]interface{})
	searchMap["workspace_id"] = workspaceId
	orgs := org.Gets(searchMap, 0, 0)
	orgResult = make(map[int]string)
	for _, v := range orgs {
		orgResult[v.Id] = v.OrgName
	}
	return
}

// mapIpToAssets 将ip资产映射为es文档
func mapIpToAssets(workspaceId int, ips []db.Ip) (docList []Document) {
	orgResult := getOrganization(workspaceId)
	for _, ipData := range ips {
		port := db.Port{IpId: ipData.Id}
		ports := port.GetsByIPId()
		// port
		for _, portData := range ports {
			doc := Document{
				Host:       fmt.Sprintf("%s:%d", ipData.IpName, portData.PortNum),
				Ip:         []string{ipData.IpName},
				Port:       portData.PortNum,
				Location:   []string{ipData.Location},
				CreateTime: portData.CreateDatetime,
				UpdateTime: portData.UpdateDatetime,
			}
			if ipData.OrgId != nil {
				if orgName, ok := orgResult[*ipData.OrgId]; ok {
					doc.Org = orgName
				}
			}
			if portData.Status != "" {
				if status, err := strconv.ParseInt(portData.Status, 10, 64); err == nil && status > 0 {
					doc.Status = int(status)
				}
			}
			// portAttr
			portAttr := db.PortAttr{RelatedId: portData.Id}
			portAttrs := portAttr.GetsByRelatedIdByDateAsc()
			source := make(map[string]struct{})
			for _, portAttrData := range portAttrs {
				switch portAttrData.Tag {
				case "service":
					doc.Service = portAttrData.Content
				case "banner":
					doc.Banner = portAttrData.Content
				case "server":
					doc.Server = portAttrData.Content
				case "title":
					doc.Title = portAttrData.Content
				case "tlsdata":
					doc.Cert = portAttrData.Content
				case "favicon":
					ihs := strings.Split(portAttrData.Content, " | ")
					if ihsData, err := strconv.ParseInt(ihs[0], 10, 64); err == nil && ihsData != 0 {
						doc.IconHash = ihsData
					}
				}
				source[portAttrData.Source] = struct{}{}
			}
			// http header and body
			htp := db.IpHttp{RelatedId: portData.Id}
			htps := htp.GetsByRelatedId()
			for _, htpData := range htps {
				if htpData.Tag == "header" {
					doc.Header = htpData.Content
				} else if htpData.Tag == "body" {
					doc.Body = htpData.Content
				}
			}
			doc.Source = utils.SetToSlice(source)
			doc.Id = SID(doc.Host)
			docList = append(docList, doc)
		}
	}
	return
}

// mapDomainToAssets 将domain资产映射为es文档
func mapDomainToAssets(workspaceId int, domains []db.Domain) (docList []Document) {
	locationV4 := custom.NewIPv4Location()
	locationV6, _ := custom.NewIPv6Location()
	tld := domainscan.NewTldExtract()
	orgResult := getOrganization(workspaceId)
	for _, domainData := range domains {
		doc := Document{
			Host:       domainData.DomainName,
			CreateTime: domainData.CreateDatetime,
			UpdateTime: domainData.UpdateDatetime,
		}
		if doc.Domain = tld.ExtractFLD(domainData.DomainName); doc.Domain == "" {
			logging.RuntimeLog.Errorf("ExtractFLD failed,host:%s", doc.Host)
			continue
		}
		if domainData.OrgId != nil {
			if orgName, ok := orgResult[*domainData.OrgId]; ok {
				doc.Org = orgName
			}
		}
		// domainAttr
		domainAttr := db.DomainAttr{RelatedId: domainData.Id}
		domainAttrs := domainAttr.GetsByRelatedIdByDateAsc()
		source := make(map[string]struct{})
		ips := make(map[string]struct{})
		locations := make(map[string]struct{})
		for _, domainAttrData := range domainAttrs {
			source[domainAttrData.Source] = struct{}{}
			switch domainAttrData.Tag {
			case "service":
				doc.Service = domainAttrData.Content
			case "banner":
				doc.Banner = domainAttrData.Content
			case "server":
				doc.Server = domainAttrData.Content
			case "title":
				doc.Title = domainAttrData.Content
			case "tlsdata":
				doc.Cert = domainAttrData.Content
			case "favicon":
				ihs := strings.Split(domainAttrData.Content, " | ")
				if ihsData, err := strconv.ParseInt(ihs[0], 10, 64); err == nil && ihsData != 0 {
					doc.IconHash = ihsData
				}
			case "A":
				ips[domainAttrData.Content] = struct{}{}
				if loc := locationV4.FindPublicIP(domainAttrData.Content); loc != "" {
					locations[loc] = struct{}{}
				}
			case "AAAA":
				ips[domainAttrData.Content] = struct{}{}
				if loc := locationV6.Find(domainAttrData.Content); loc != "" {
					locations[loc] = struct{}{}
				}
			}
		}
		// http header and body
		htp := db.DomainHttp{RelatedId: domainData.Id}
		htps := htp.GetsByRelatedId()
		for _, htpData := range htps {
			if htpData.Tag == "header" {
				doc.Header = htpData.Content
			} else if htpData.Tag == "body" {
				doc.Body = htpData.Content
			}
		}
		doc.Source = utils.SetToSlice(source)
		doc.Ip = utils.SetToSlice(ips)
		doc.Location = utils.SetToSlice(locations)
		doc.Id = SID(doc.Host)
		docList = append(docList, doc)
	}

	return
}

// SyncIpAssets 同步IP资产到es
func SyncIpAssets(workspaceId int, ips []db.Ip) bool {
	workspace := db.Workspace{Id: workspaceId}
	if !workspace.Get() {
		logging.RuntimeLog.Errorf("workspace id:%d not found", workspaceId)
		return false
	}
	assets := NewAssets(workspace.WorkspaceGUID)
	if !assets.CreateIndex() {
		logging.RuntimeLog.Errorf("Create index for workspace id:%d,GUID:%s failed", workspaceId, workspace.WorkspaceGUID)
		return false
	}
	// mapping
	docList := mapIpToAssets(workspaceId, ips)
	// bulk导入
	logging.RuntimeLog.Infof("start to bulk import,total:%d", len(docList))
	return assets.BulkIndexDoc(docList)
}

// SyncDomainAssets 同步Domain资产到es
func SyncDomainAssets(workspaceId int, domains []db.Domain) bool {
	workspace := db.Workspace{Id: workspaceId}
	if !workspace.Get() {
		logging.RuntimeLog.Errorf("workspace id:%d not found", workspaceId)
		return false
	}
	assets := NewAssets(workspace.WorkspaceGUID)
	if !assets.CreateIndex() {
		logging.RuntimeLog.Errorf("Create index for workspace id:%d,GUID:%s failed", workspaceId, workspace.WorkspaceGUID)
		return false
	}
	// mapping
	docList := mapDomainToAssets(workspaceId, domains)
	// bulk导入
	logging.RuntimeLog.Infof("start to bulk import,total:%d", len(docList))
	return assets.BulkIndexDoc(docList)
}

// ImportIpAssets 从mysql中导入指定workspace的IP资产
func ImportIpAssets(workspaceResult map[int]string) {
	for wsId, wsGuid := range workspaceResult {
		logging.RuntimeLog.Infof("start to import workspace id:%d,GUID:%s...", wsId, wsGuid)
		// ip
		ip := db.Ip{WorkspaceId: wsId}
		ips, ipCount := ip.Gets(map[string]interface{}{}, 0, 0, false)
		logging.RuntimeLog.Infof("total ip count:%d", ipCount)
		SyncIpAssets(wsId, ips)
	}
}

// ImportDomainAssets 从mysql中导入指定workspace的Domain资产
func ImportDomainAssets(workspaceResult map[int]string) {
	for wsId, wsGuid := range workspaceResult {
		logging.RuntimeLog.Infof("start to import workspace id:%d,GUID:%s...", wsId, wsGuid)
		// domain
		domain := db.Domain{WorkspaceId: wsId}
		domains, domainCount := domain.Gets(map[string]interface{}{}, 0, 0, false)
		logging.RuntimeLog.Infof("total domain count:%d", domainCount)
		SyncDomainAssets(wsId, domains)
	}
}

// ImportAssetsFromFile 导入JSON格式的资产数据
func ImportAssetsFromFile(indexName string, filename string) bool {
	/*
		从elasticdump的数据，每一行一个文档。
		elasticdump --input='https://user:password@localhost:9200/b0c79065-7ff7-32ae-cc18-864ccd8f7717' --output=nemo.json --type=data
			{
			    "_index": "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
			    "_id": "4236e91509189f00d24919800c34b9a423240048",
			    "_score": 1,
			    "_source": {
			        "id": "4236e91509189f00d24919800c34b9a423240048",
			        "host": "23.46.63.43:443",
			        "ip": [
			            "23.46.63.43"
			        ],
			        "port": 443,
			        "domain": "",
			        "location": [
			            "澳大利亚"
			        ],
			        "status": 0,
			        "service": "",
			        "banner": "",
			        "server": "",
			        "title": "Investors | Ocugen, Inc.",
			        "header": "",
			        "body": "",
			        "cert": "",
			        "icon_hash": 0,
			        "org": "",
			        "source": [
			            "hunter"
			        ],
			        "comment": "",
			        "create_time": "2023-08-09T10:59:38+08:00",
			        "update_time": "2023-08-09T10:59:38+08:00"
			    }
			}
	*/
	type elasticJsonData struct {
		Source Document `json:"_source"`
	}
	inputFile, err := os.Open(filename)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return false
	}
	defer inputFile.Close()

	assets := NewAssets(indexName)
	if !assets.CreateIndex() {
		logging.RuntimeLog.Errorf("Create index for workspace id:%s failed", indexName)
		logging.CLILog.Errorf("Create index for workspace id:%s failed", indexName)
		return false
	}
	scanner := bufio.NewScanner(inputFile)
	var docList []Document
	for scanner.Scan() {
		var data elasticJsonData
		err = json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			continue
		}
		docList = append(docList, data.Source)
	}
	return assets.BulkIndexDoc(docList)
}
