package utils

import (
	"embed"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/xraypocv2/pkg/xray/structs"
	"gopkg.in/yaml.v2"
	"strings"
	"sync"
)

// 在poc.go中的var ORDER变量是全局共享变量
// yaml的序列化在在多个线程中会冲突，导致rule出错
// 因此通过加锁来避免

var yamlUnmarshalMutex sync.Mutex

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
	yamlUnmarshalMutex.Lock()
	defer yamlUnmarshalMutex.Unlock()

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
	yamlUnmarshalMutex.Lock()
	defer yamlUnmarshalMutex.Unlock()

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
