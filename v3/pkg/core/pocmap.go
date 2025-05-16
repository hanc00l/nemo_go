package core

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MapInfo struct {
	Name        string   `yaml:"name"`
	Category    string   `yaml:"category"`
	Description string   `yaml:"description"`
	Updated     string   `yaml:"updated"`
	Fingerprint []string `yaml:"fingerprint"`
	Poc         []struct {
		Source string   `yaml:"source"`
		Path   []string `yaml:"path"`
	} `yaml:"poc"`
}

func MatchAssetPoc(docs []db.AssetDocument, pocscanConfig execute.PocscanConfig) (targetPocMapResult map[string]execute.PocscanConfig) {
	if len(docs) == 0 {
		return
	}
	mapInfo, err := loadPocMapJson()
	if err != nil {
		logging.RuntimeLog.Errorf("加载poc_map失败: %s", err)
		return
	}
	targetPocMapResult = make(map[string]execute.PocscanConfig)
	for _, doc := range docs {
		// match fingerprint：app、service
		var fingerprint []string
		if len(doc.App) > 0 {
			fingerprint = append(fingerprint, doc.App...)
		}
		if len(doc.Service) > 0 {
			fingerprint = append(fingerprint, doc.Service)
		}
		if len(fingerprint) == 0 {
			continue
		}
		pocList := matchSingle(fingerprint, mapInfo)
		if len(pocList) == 0 {
			continue
		}
		logging.RuntimeLog.Infof("指纹：%v 匹配到poc: %s", fingerprint, pocList)
		c := pocscanConfig
		c.Target = doc.Authority
		c.PocFile = strings.Join(pocList, ",")
		targetPocMapResult[c.Target] = c
	}

	return
}

func loadPocMapJson() (pocMap []MapInfo, err error) {
	jsonFilePath := filepath.Join(conf.GetAbsRootPath(), "thirdparty/dict/web_poc_map_v2.json")
	file, err := os.OpenFile(jsonFilePath, os.O_RDONLY, 0400)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &pocMap)
	if err != nil {
		return nil, err
	}
	return pocMap, nil
}

func matchSingle(app []string, pocMap []MapInfo) (pocList []string) {
	for _, a := range app {
		// match fingerprint
		appFingerprint := strings.ToLower(a)
		for _, poc := range pocMap {
			for _, f := range poc.Fingerprint {
				pocMapFingerprint := strings.ToLower(f)
				if strings.Contains(appFingerprint, pocMapFingerprint) {
					pocList = append(pocList, getPocList(poc)...)
					break
				}
			}
		}
	}

	return
}

func getPocList(mapInfo MapInfo) (pocList []string) {
	sourcePaths := map[string]string{
		"some_nuclei_templates": "some_nuclei_templates",
		"nuclei-template":       "nuclei-templates",
	}
	for _, poc := range mapInfo.Poc {
		pocPath, ok := sourcePaths[poc.Source]
		if !ok {
			logging.RuntimeLog.Errorf("poc_map中不正确的source：%s", poc.Source)
			continue
		}
		for _, pocFile := range poc.Path {
			pocList = append(pocList, filepath.Join(pocPath, pocFile))
		}
	}
	return
}
