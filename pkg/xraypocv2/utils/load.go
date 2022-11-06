package utils

import (
	"embed"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/xraypocv2/pkg/xray/structs"
	"gopkg.in/yaml.v2"
	"strings"
)

func LoadMultiPoc(Pocs embed.FS, pocname string) []*structs.Poc {
	var pocs []*structs.Poc
	for _, f := range SelectPoc(Pocs, pocname) {
		if p, err := loadPoc(f, Pocs); err == nil {
			pocs = append(pocs, p)
		}
	}
	return pocs
}
func LoadPocByBytes(pocBytes []byte) (*structs.Poc, error) {
	p := &structs.Poc{}
	//yamlFile, err := Pocs.ReadFile("xrayFiles/" + fileName)
	//
	//if err != nil {
	//	return nil, err
	//}
	err := yaml.Unmarshal(pocBytes, p)
	if err != nil {
		return nil, err
	}
	return p, err
}
func loadPoc(fileName string, Pocs embed.FS) (*structs.Poc, error) {
	p := &structs.Poc{}
	yamlFile, err := Pocs.ReadFile("xrayFiles/" + fileName)

	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, p)
	if err != nil {
		return nil, err
	}
	return p, err
}

func SelectPoc(Pocs embed.FS, pocname string) []string {
	entries, err := Pocs.ReadDir("xrayFiles")
	if err != nil {
		fmt.Println(err)
	}
	var foundFiles []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), pocname+"-") {
			foundFiles = append(foundFiles, entry.Name())
		}
	}
	return foundFiles
}
