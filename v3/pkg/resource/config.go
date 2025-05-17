package resource

const (
	ExecuteFile = "executeFile"
	ConfigFile  = "configFile"
	DataFile    = "dataFile"
	Dir         = "dir"
)

const (
	WorkerCategory       = "worker"
	GogoCategory         = "gogo"
	DictCategory         = "dict"
	MassdnsCategory      = "massdns"
	SubfinderCategory    = "subfinder"
	HttpxCategory        = "httpx"
	FingerprintxCategory = "fingerprintx"
	NucleiCategory       = "nuclei"
	QqwryCategory        = "qqwry"
	Zxipv6wryCategory    = "zxipv6wry"
	Geolite2Category     = "geolite2"
	NmapCategory         = "nmap"
	PocFileCategory      = "pocfile"
)

type Resource struct {
	Name string
	Type string
	Path string

	Bytes []byte
	Hash  string
}

var Resources = make(map[string]map[string]Resource)

func init() {
	worker := make(map[string]Resource)
	worker["worker_darwin_amd64"] = Resource{Name: "worker_darwin_amd64", Type: ExecuteFile, Path: "."}
	worker["worker_linux_amd64"] = Resource{Name: "worker_linux_amd64", Type: ExecuteFile, Path: "."}
	worker["worker_windows_amd64.exe"] = Resource{Name: "worker_windows_amd64.exe", Type: ExecuteFile, Path: "."}
	Resources[WorkerCategory] = worker

	gogo := make(map[string]Resource)
	gogo["gogo_darwin_amd64"] = Resource{Name: "gogo_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/gogo"}
	gogo["gogo_linux_amd64"] = Resource{Name: "gogo_linux_amd64", Type: ExecuteFile, Path: "thirdparty/gogo"}
	gogo["gogo_windows_amd64.exe"] = Resource{Name: "gogo_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/gogo"}
	Resources[GogoCategory] = gogo

	dict := make(map[string]Resource)
	dict["dicc.txt"] = Resource{Name: "dicc.txt", Type: ConfigFile, Path: "thirdparty/dict"}
	dict["resolver.txt"] = Resource{Name: "resolver.txt", Type: ConfigFile, Path: "thirdparty/dict"}
	dict["subnames_medium.txt"] = Resource{Name: "subnames_medium.txt", Type: ConfigFile, Path: "thirdparty/dict"}
	dict["subnames.txt"] = Resource{Name: "subnames.txt", Type: ConfigFile, Path: "thirdparty/dict"}
	dict["web_fingerprint_v4.json"] = Resource{Name: "web_fingerprint_v4.json", Type: ConfigFile, Path: "thirdparty/dict"}
	dict["web_poc_map_v2.json"] = Resource{Name: "web_poc_map_v2.json", Type: ConfigFile, Path: "thirdparty/dict"}
	Resources[DictCategory] = dict

	massdns := make(map[string]Resource)
	massdns["massdns_darwin_amd64"] = Resource{Name: "massdns_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/massdns"}
	massdns["massdns_linux_amd64"] = Resource{Name: "massdns_linux_amd64", Type: ExecuteFile, Path: "thirdparty/massdns"}
	massdns["massdns_windows_amd64.exe"] = Resource{Name: "massdns_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/massdns"}
	Resources[MassdnsCategory] = massdns

	subfinder := make(map[string]Resource)
	subfinder["subfinder_darwin_amd64"] = Resource{Name: "subfinder_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/subfinder"}
	subfinder["subfinder_linux_amd64"] = Resource{Name: "subfinder_linux_amd64", Type: ExecuteFile, Path: "thirdparty/subfinder"}
	subfinder["subfinder_windows_amd64.exe"] = Resource{Name: "subfinder_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/subfinder"}
	Resources[SubfinderCategory] = subfinder

	httpx := make(map[string]Resource)
	httpx["httpx_darwin_amd64"] = Resource{Name: "httpx_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/httpx"}
	httpx["httpx_linux_amd64"] = Resource{Name: "httpx_linux_amd64", Type: ExecuteFile, Path: "thirdparty/httpx"}
	httpx["httpx_windows_amd64.exe"] = Resource{Name: "httpx_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/httpx"}
	Resources[HttpxCategory] = httpx

	fingerprintx := make(map[string]Resource)
	fingerprintx["fingerprintx_darwin_amd64"] = Resource{Name: "fingerprintx_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/fingerprintx"}
	fingerprintx["fingerprintx_linux_amd64"] = Resource{Name: "fingerprintx_linux_amd64", Type: ExecuteFile, Path: "thirdparty/fingerprintx"}
	fingerprintx["fingerprintx_windows_amd64.exe"] = Resource{Name: "fingerprintx_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/fingerprintx"}
	Resources[FingerprintxCategory] = fingerprintx
	//IP归属地数据库文件
	qqwry := make(map[string]Resource)
	qqwry["qqwry.dat"] = Resource{Name: "qqwry.dat", Type: DataFile, Path: "thirdparty/qqwry"}
	Resources[QqwryCategory] = qqwry

	zxipv6wry := make(map[string]Resource)
	zxipv6wry["ipv6wry.db"] = Resource{Name: "ipv6wry.db", Type: DataFile, Path: "thirdparty/zxipv6wry"}
	Resources[Zxipv6wryCategory] = zxipv6wry

	//CDN查询数据库文件
	geolite2 := make(map[string]Resource)
	geolite2["GeoLite2-ASN.mmdb"] = Resource{Name: "GeoLite2-ASN.mmdb", Type: DataFile, Path: "thirdparty/geolite2"}
	Resources[Geolite2Category] = geolite2

	nmap := make(map[string]Resource)
	nmap["nmap-services"] = Resource{Name: "nmap-services", Type: ConfigFile, Path: "thirdparty/nmap"}
	Resources[NmapCategory] = nmap

	nuclei := make(map[string]Resource)
	nuclei["nuclei_darwin_amd64"] = Resource{Name: "nuclei_darwin_amd64", Type: ExecuteFile, Path: "thirdparty/nuclei"}
	nuclei["nuclei_linux_amd64"] = Resource{Name: "nuclei_linux_amd64", Type: ExecuteFile, Path: "thirdparty/nuclei"}
	nuclei["nuclei_windows_amd64.exe"] = Resource{Name: "nuclei_windows_amd64.exe", Type: ExecuteFile, Path: "thirdparty/nuclei"}
	Resources[NucleiCategory] = nuclei

	//poc文件
	pocfile := make(map[string]Resource)
	pocfile["nuclei-templates"] = Resource{Name: "nuclei-templates", Type: Dir, Path: "thirdparty/nuclei"}
	pocfile["some_nuclei_templates"] = Resource{Name: "some_nuclei_templates", Type: Dir, Path: "thirdparty/nuclei"}
	Resources[PocFileCategory] = pocfile
}
