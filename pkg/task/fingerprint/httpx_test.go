package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestHttpx_Run(t *testing.T) {
	domainConfig := domainscan.Config{Target: "800best.com"}
	subdomain := domainscan.NewSubFinder(domainConfig)
	subdomain.Do()
	t.Log(subdomain.Result)

	httpx := NewHttpx()
	httpx.ResultDomainScan = subdomain.Result
	httpx.Do()
	t.Log(httpx.ResultDomainScan)
	for d, da := range httpx.ResultDomainScan.DomainResult {
		t.Log(d, da)
	}
	subdomain.Result.SaveResult(subdomain.Config)
}

func TestHttpx_Run2(t *testing.T) {
	nmapConfig := portscan.Config{
		Target: "47.98.181.116",
		Port:   "80,443",
		Rate:   1000,
		IsPing: false,
		Tech:   "-sS",
		CmdBin: "nmap",
	}
	nmap := portscan.NewNmap(nmapConfig)
	nmap.Do()

	httpx := NewHttpx()
	httpx.ResultPortScan = nmap.Result
	httpx.Do()
	for ip, r := range httpx.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
	//httpx.ResultPortScan.SaveResult(nmap.Config)
}

func TestHttpx_ParseJSONContentResult(t *testing.T) {
	data := `{"timestamp":"2022-07-19T19:02:35.330689+08:00","hashes":{"body-md5":"20e77a2cdde432fb1c330007299f19f2","body-mmh3":"-1537041958","body-sha256":"c8428b27215e7edcebae356a4405b12d148bd1ad7030cc63284a180ef9c7e7be","body-simhash":"14425645466038018950","header-md5":"8223778d8d832043a1320fe156a55615","header-mmh3":"-1306485643","header-sha256":"1e567ffbcddc909ff5e96d25e6417d00b781d59e9189f4304d2f2f67bc4790ea","header-simhash":"9814387088565714943"},"port":"8000","url":"http://127.0.0.1:8000","input":"127.0.0.1:8000","title":"Directory listing for /","scheme":"http","webserver":"SimpleHTTP/0.6 Python/3.9.13","content-type":"text/html","method":"GET","host":"127.0.0.1","path":"/","response-time":"1.514799ms","a":["127.0.0.1"],"technologies":["Python:3.9.13","SimpleHTTP:0.6"],"words":17,"lines":16,"status-code":200,"content-length":336,"failed":false}
{"timestamp":"2022-07-19T19:02:35.335754+08:00","hashes":{"body-md5":"4c9bd066572c8a29c3a35bf97e9c5472","body-mmh3":"-241930836","body-sha256":"e959fa807ef1ee875d6876fe5eed02095ea62222b7636166f2b471af544e61e1","body-simhash":"18057126888269405110","header-md5":"6d757f335eb333195d4225b1069b087c","header-mmh3":"-1604855004","header-sha256":"b075a17f7e4b3196a62d1244555e73ef7a982c61b0d7ae03bb0dafd7c01519c7","header-simhash":"9814066098791485551"},"port":"49157","url":"http://127.0.0.1:49157","input":"127.0.0.1:49157","scheme":"http","content-type":"application/json","method":"GET","host":"127.0.0.1","path":"/","response-time":"1.121499ms","a":["127.0.0.1"],"words":4,"lines":1,"status-code":200,"content-length":58,"failed":false}
{"timestamp":"2022-07-19T19:02:35.355728+08:00","hashes":{"body-md5":"4c9bd066572c8a29c3a35bf97e9c5472","body-mmh3":"-241930836","body-sha256":"e959fa807ef1ee875d6876fe5eed02095ea62222b7636166f2b471af544e61e1","body-simhash":"18057126888269405110","header-md5":"6d757f335eb333195d4225b1069b087c","header-mmh3":"-1604855004","header-sha256":"b075a17f7e4b3196a62d1244555e73ef7a982c61b0d7ae03bb0dafd7c01519c7","header-simhash":"9814066098791485551"},"port":"49157","url":"http://localhost:49157","input":"localhost:49157","scheme":"http","content-type":"application/json","method":"GET","host":"127.0.0.1","path":"/","response-time":"1.191473ms","a":["127.0.0.1","::1"],"words":4,"lines":1,"status-code":200,"content-length":58,"failed":false}
{"timestamp":"2022-07-19T19:02:35.362846+08:00","hashes":{"body-md5":"20e77a2cdde432fb1c330007299f19f2","body-mmh3":"-1537041958","body-sha256":"c8428b27215e7edcebae356a4405b12d148bd1ad7030cc63284a180ef9c7e7be","body-simhash":"14425645466038018950","header-md5":"8223778d8d832043a1320fe156a55615","header-mmh3":"-1306485643","header-sha256":"1e567ffbcddc909ff5e96d25e6417d00b781d59e9189f4304d2f2f67bc4790ea","header-simhash":"9814387088565714943"},"port":"8000","url":"http://172.16.222.1:8000","input":"172.16.222.1:8000","title":"Directory listing for /","scheme":"http","webserver":"SimpleHTTP/0.6 Python/3.9.13","content-type":"text/html","method":"GET","host":"172.16.222.1","path":"/","response-time":"1.08643ms","a":["172.16.222.1"],"technologies":["Python:3.9.13","SimpleHTTP:0.6"],"words":17,"lines":16,"status-code":200,"content-length":336,"failed":false}
{"timestamp":"2022-07-19T19:02:35.369681+08:00","hashes":{"body-md5":"20e77a2cdde432fb1c330007299f19f2","body-mmh3":"-1537041958","body-sha256":"c8428b27215e7edcebae356a4405b12d148bd1ad7030cc63284a180ef9c7e7be","body-simhash":"14425645466038018950","header-md5":"8223778d8d832043a1320fe156a55615","header-mmh3":"-1306485643","header-sha256":"1e567ffbcddc909ff5e96d25e6417d00b781d59e9189f4304d2f2f67bc4790ea","header-simhash":"9814387088565714943"},"port":"8000","url":"http://localhost:8000","input":"localhost:8000","title":"Directory listing for /","scheme":"http","webserver":"SimpleHTTP/0.6 Python/3.9.13","content-type":"text/html","method":"GET","host":"127.0.0.1","path":"/","response-time":"1.107331ms","a":["127.0.0.1","::1"],"technologies":["Python:3.9.13","SimpleHTTP:0.6"],"words":17,"lines":16,"status-code":200,"content-length":336,"failed":false}`
	httpx := NewHttpx()
	httpx.ParseJSONContentResult([]byte(data))
	for k, ip := range httpx.ResultPortScan.IPResult {
		t.Log(k)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
}
