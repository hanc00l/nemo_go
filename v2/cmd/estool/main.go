package main

import (
	"flag"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/db"
	"github.com/hanc00l/nemo_go/v2/pkg/es"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
)

var (
	JsonPathFile  string
	WorkspaceGUID string
	IsList        bool
	IsImport      bool
	IsCreate      bool
	IsDelete      bool
)

func parseOption() {
	flag.StringVar(&JsonPathFile, "f", "", "import by json file path")
	flag.StringVar(&WorkspaceGUID, "w", "", "import index name by workspace guid")
	flag.BoolVar(&IsList, "l", false, "list all es index")
	flag.BoolVar(&IsImport, "i", false, "import data to es")
	flag.BoolVar(&IsCreate, "c", false, "create index")
	flag.BoolVar(&IsDelete, "d", false, "delete index")

	flag.Parse()

}

func main() {
	parseOption()

	if !IsList && !IsImport && !IsCreate && !IsDelete {
		flag.Usage()
		fmt.Println("Usage: " +
			"\n\testools -l\t\t\t\tlist all es index" +
			"\n\testools -i -w workspace_guid\t\timport data from nemo database to elasticsearch" +
			"\n\testools -c -w workspace_guid\t\tcreate index" +
			"\n\testools -d -w workspace_guid\t\tdelete index" +
			"\n\testools -i -w workspace_guid -f json_file_path\t\timport data from json file to elasticsearch")
		return
	}
	if IsList {
		a := es.NewAssets("")
		result := a.ListAllIndices()
		fmt.Printf("%s\t%s\n", "docCount", "indexName")
		for indexName, count := range result {
			fmt.Printf("%s\t\t%s\n", count, indexName)
		}
		return
	}
	if WorkspaceGUID == "" {
		logging.CLILog.Error("workspace guid is required")
		return
	}
	if IsCreate {
		assets := es.NewAssets(WorkspaceGUID)
		if assets.CreateIndex() {
			logging.CLILog.Info("create index success")
		}
	} else if IsImport {
		var confirm string
		fmt.Print("Are you sure to import? (y/n): ")
		_, err := fmt.Scan(&confirm)
		if err != nil || confirm != "y" {
			return
		}
		// 显示导入的数量
		es.ShowBulkIndexStatistic = true
		if len(JsonPathFile) > 0 {
			logging.CLILog.Info(es.ImportAssetsFromFile(WorkspaceGUID, JsonPathFile))
		} else {
			workspace := db.Workspace{WorkspaceGUID: WorkspaceGUID}
			if !workspace.GetByGUID() {
				logging.CLILog.Error("workspace guid is not exist")
				return
			}
			ws := map[int]string{workspace.Id: workspace.WorkspaceGUID}
			logging.CLILog.Info("start to import ip assets...")
			es.ImportIpAssets(ws)
			logging.CLILog.Info("start to import domain assets...")
			es.ImportDomainAssets(ws)
		}
	} else if IsDelete {
		var confirm string
		fmt.Print("Are you sure to delete? (y/n): ")
		_, err := fmt.Scan(&confirm)
		if err != nil || confirm != "y" {
			return
		}
		assets := es.NewAssets(WorkspaceGUID)
		if assets.DeleteIndex() {
			logging.CLILog.Info("delete index success")
		}
	}
}
