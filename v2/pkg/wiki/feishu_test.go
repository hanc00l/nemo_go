package wiki

import (
	"testing"
	"time"
)

func TestListDocument(t *testing.T) {
	spaceId := ""
	w := NewFeishuWiki()
	err, docs := w.GetDocuments(spaceId)
	if err != nil {
		t.Log(err)
	}
	for _, doc := range docs {
		t.Log(doc)
	}
}

func TestGetUserAccessTokenByCode(t *testing.T) {
	code := ""
	w := NewFeishuWiki()
	t.Log(w.GetUserAccessTokenByCode(code))
	t.Log(w.userAccessRefreshToken)
}

func TestNewDocument(t *testing.T) {
	spaceId := ""
	w := NewFeishuWiki()
	t.Log(w.NewDocument(spaceId, "测试1111", "test", 0))
}

func TestGetDocument(t *testing.T) {
	nodeToken := ""
	w := NewFeishuWiki()
	t.Log(w.GetDocument(nodeToken))
}

func TestFeishuWiki_ExportDocument(t *testing.T) {
	objType := "docx"
	objToken := ""
	w := NewFeishuWiki()
	err, ticket := w.CreateExportTask(objType, objToken)
	t.Log(err, ticket)
	var fileToken string
	var d int
	for {
		d++
		t.Log(d)
		if fileToken != "" || d > 10 {
			break
		}
		time.Sleep(1 * time.Second)
		err, fileToken = w.QueryExportTask(objToken, ticket)
		t.Log(err, fileToken)
	}
	err = w.DownloadExportTask(fileToken, "/tmp/1.docx")
	t.Log(err)
}

func TestFeishuWiki_GetDocumentContent(t *testing.T) {
	objToken := ""
	w := NewFeishuWiki()
	t.Log(w.GetDocumentContent(objToken))
}
