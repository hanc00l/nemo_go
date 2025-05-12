package fingerprint

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
	"testing"
)

func TestFingerPrintHub(t *testing.T) {
	engine := NewFingerprintHubEngine()
	finerJsonFile := filepath.Join(conf.GetAbsRootPath(), "thirdparty/dict/web_fingerprint_v4.json")
	if err := engine.LoadFromFile(finerJsonFile); err != nil {
		t.Logf("Failed to load fingerprints: %v\n", err)
		return
	}

	faviconHash := "419828698"
	header := `Content-Type: text/html\nX-Powered-By: ThinkPHP`
	body := `<html><title>log in Test Page</title></html>`

	results := engine.Match(faviconHash, header, body)
	t.Log(results)
	results = utils.RemoveDuplicatesAndSort(results)

	for _, result := range results {
		t.Logf("Matched: %s \n", result)
	}
}
