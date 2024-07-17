package es

import "testing"

func TestImportIpAssets(t *testing.T) {
	ShowBulkIndexStatistic = true
	//本地测试第一个
	workspaceResult := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	//a := NewAssets(workspaceResult[1])
	//a.DeleteIndex()

	ImportIpAssets(workspaceResult)
}

func TestImportDomainAssets(t *testing.T) {
	ShowBulkIndexStatistic = true
	//本地测试第一个
	workspaceResult := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	a := NewAssets(workspaceResult[1])
	a.DeleteIndex()

	ImportDomainAssets(workspaceResult)
}
