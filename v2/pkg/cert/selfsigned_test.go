package cert

import (
	"testing"
)

func TestSelfSignedCertGenerator_Generate(t *testing.T) {
	t.Log(GenerateSelfSignedCert("server.crt", "server.key"))
}
