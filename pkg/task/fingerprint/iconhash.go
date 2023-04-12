package fingerprint

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/mat/besticon/besticon"
	"github.com/remeh/sizedwaitgroup"
	"github.com/twmb/murmur3"
	"hash"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type IconHash struct {
	ResultPortScan     portscan.Result
	ResultDomainScan   domainscan.Result
	IconHashInfoResult IconHashInfoResult
	DomainTargetPort   map[string]map[int]struct{}
}

type IconHashResult struct {
	Url       string
	Hash      string
	ImageData []byte
}

type IconHashInfo struct {
	Url       string
	Hash      string
	ImageData []byte
}

type IconHashInfoResult struct {
	sync.RWMutex
	Result []IconHashInfo
}

var (
	// used by besticon
	keepImageBytes = true
	// used by github.com/Becivells/iconhash
	isUint32 = false
	debug    = false
)

func NewIconHash() *IconHash {
	return &IconHash{}
}

func (i *IconHash) Do() {
	swg := sizedwaitgroup.New(fpIconHashThreadNumber[conf.WorkerPerformanceMode])

	if i.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range i.ResultPortScan.IPResult {
			for portNumber := range ipResult.Ports {
				if _, ok := blankPort[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					defer swg.Done()
					iconHashes := i.RunFetchIconHashes(u)
					if len(iconHashes) > 0 {
						for _, r := range iconHashes {
							par := portscan.PortAttrResult{
								Source:  "iconhash",
								Tag:     "favicon",
								Content: fmt.Sprintf("%s | %s", r.Hash, r.Url),
							}
							i.ResultPortScan.SetPortAttr(ip, port, par)
							if len(r.ImageData) > 0 {
								i.IconHashInfoResult.Lock()
								i.IconHashInfoResult.Result = append(i.IconHashInfoResult.Result, IconHashInfo{
									Url:       r.Url,
									Hash:      r.Hash,
									ImageData: r.ImageData,
								})
								i.IconHashInfoResult.Unlock()
							}
						}
					}
				}(ipName, portNumber, url)
			}
		}
	}
	if i.ResultDomainScan.DomainResult != nil {
		if i.DomainTargetPort == nil {
			i.DomainTargetPort = make(map[string]map[int]struct{})
		}
		blackDomain := domainscan.NewBlankDomain()
		for domain := range i.ResultDomainScan.DomainResult {
			if blackDomain.CheckBlank(domain) {
				continue
			}
			//如果无域名对应的端口，默认80和443
			if _, ok := i.DomainTargetPort[domain]; !ok || len(i.DomainTargetPort[domain]) == 0 {
				i.DomainTargetPort[domain] = make(map[int]struct{})
				i.DomainTargetPort[domain][80] = struct{}{}
				i.DomainTargetPort[domain][443] = struct{}{}
			}
			for port := range i.DomainTargetPort[domain] {
				if _, ok := blankPort[port]; ok {
					continue
				}
				url := fmt.Sprintf("%s:%d", domain, port)
				swg.Add()
				go func(d string, u string) {
					defer swg.Done()
					iconHashes := i.RunFetchIconHashes(u)
					if len(iconHashes) > 0 {
						for _, r := range iconHashes {
							dar := domainscan.DomainAttrResult{
								Source:  "iconhash",
								Tag:     "favicon",
								Content: fmt.Sprintf("%s | %s", r.Hash, r.Url),
							}
							i.ResultDomainScan.SetDomainAttr(d, dar)
							if len(r.ImageData) > 0 {
								i.IconHashInfoResult.Lock()
								i.IconHashInfoResult.Result = append(i.IconHashInfoResult.Result, IconHashInfo{
									Url:       r.Url,
									Hash:      r.Hash,
									ImageData: r.ImageData,
								})
								i.IconHashInfoResult.Unlock()
							}
						}
					}
				}(domain, url)
			}
		}
	}

	swg.Wait()
}

func (i *IconHash) RunFetchIconHashes(url string) (hashResult []IconHashResult) {
	var icons []besticon.Icon
	protocoll := utils.GetProtocol(url, 5)
	// 获取缺省的icon:
	u1 := fmt.Sprintf("%s://%s/favicon.ico", protocoll, url)
	icon1 := fetchIconDetails(u1)
	if icon1.Error == nil {
		icons = append(icons, icon1)
	}
	// 获取网页中定义的全部icon
	u2 := fmt.Sprintf("%s://%s", protocoll, url)
	finder := besticon.IconFinder{}
	icons2, err := finder.FetchIcons(u2)
	if err == nil {
		icons = append(icons, icons2...)
	}
	//计算哈希值
	for _, icon := range icons {
		hash := mmh3Hash32(standBase64(icon.ImageData))
		if isContainString(hashResult, hash) == false {
			hashResult = append(hashResult, IconHashResult{Url: icon.URL, Hash: hash, ImageData: icon.ImageData})
		}
	}
	return
}

// SaveFile 保存icon Image文件
func (i *IconHash) SaveFile(localSavePath string, result []IconHashInfo) string {
	count := 0
	for _, ihf := range result {
		if ihf.Url == "" || ihf.Hash == "" || len(ihf.ImageData) <= 0 {
			continue
		}
		fileSuffix := utils.GetFaviconSuffixUrl(ihf.Url)
		if fileSuffix == "" {
			continue
		}
		//文件名为md5(iconHash).后缀
		filePathName := filepath.Join(localSavePath, fmt.Sprintf("%s.%s", utils.MD5(ihf.Hash), fileSuffix))
		err := os.WriteFile(filePathName, ihf.ImageData, 0666)
		if err != nil {
			logging.RuntimeLog.Errorf("write file %s fail:%v", filePathName, err)
			continue
		}
		count++
	}
	return fmt.Sprintf("iconimage:%d", count)
}

// fetchIconDetails 根据URL获取并校验icon文件
// forked from github.com/mat/besticon/besticon/besticon.go
func fetchIconDetails(url string) besticon.Icon {
	i := besticon.Icon{URL: url}

	response, e := besticon.Get(url)
	if e != nil {
		i.Error = e
		return i
	}

	b, e := besticon.GetBodyBytes(response)
	if e != nil {
		i.Error = e
		return i
	}

	if isSVG(b) {
		// Special handling for svg, which golang can't decode with
		// image.DecodeConfig. Fill in an absurdly large width/height so SVG always
		// wins size contests.
		i.Format = "svg"
		i.Width = 9999
		i.Height = 9999
	} else {
		cfg, format, e := image.DecodeConfig(bytes.NewReader(b))
		if e != nil {
			i.Error = fmt.Errorf("besticon: unknown image format: %s", e)
			return i
		}

		// jpeg => jpg
		if format == "jpeg" {
			format = "jpg"
		}

		i.Width = cfg.Width
		i.Height = cfg.Height
		i.Format = format
	}

	i.Bytes = len(b)
	i.Sha1sum = sha1Sum(b)
	if keepImageBytes {
		i.ImageData = b
	}

	return i
}

// SVG detector. We can't use image.RegisterFormat, since RegisterFormat is
// limited to a simple magic number check. It's easy to confuse the first few
// bytes of HTML with SVG.
func isSVG(body []byte) bool {
	// is it long enough?
	if len(body) < 10 {
		return false
	}

	// does it start with something reasonable?
	switch {
	case bytes.Equal(body[0:2], []byte("<!")):
	case bytes.Equal(body[0:2], []byte("<?")):
	case bytes.Equal(body[0:4], []byte("<svg")):
	default:
		return false
	}

	// is there an <svg in the first 300 bytes?
	if off := bytes.Index(body, []byte("<svg")); off == -1 || off > 300 {
		return false
	}

	return true
}

func sha1Sum(b []byte) string {
	hash := sha1.New()
	hash.Write(b)
	bs := hash.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// mmh3Hash32 计算icon的hash
// forked from https://github.com/Becivells/iconhash
func mmh3Hash32(raw []byte) string {
	var h32 hash.Hash32 = murmur3.New32()
	h32.Write(raw)
	if isUint32 {
		return fmt.Sprintf("%d", h32.Sum32())
	}
	return fmt.Sprintf("%d", int32(h32.Sum32()))
}

// standBase64 计算 base64 的值
func standBase64(braw []byte) []byte {
	bckd := base64.StdEncoding.EncodeToString(braw)
	var buffer bytes.Buffer
	for i := 0; i < len(bckd); i++ {
		ch := bckd[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	if debug {
		fmt.Print("---------------------------start base64 content--------------------------------\n")
		fmt.Printf("====> base64:\n%s\n", buffer.String())
		defer fmt.Print("---------------------------end base64 content--------------------------------\n")
	}
	return buffer.Bytes()
}

// SplitChar76 按照 76 字符切分
func SplitChar76(braw []byte) []byte {
	// 去掉 data:image/vnd.microsoft.icon;base64
	if strings.HasPrefix(string(braw), "data:image/vnd.microsoft.icon;base64,") {
		braw = braw[37:]
	}

	var buffer bytes.Buffer
	for i := 0; i < len(braw); i++ {
		ch := braw[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')

	if debug {
		fmt.Print("---------------------------start base64 content--------------------------------\n")
		fmt.Printf("====> base64 split 76:\n %s\n", buffer.String())
		defer fmt.Print("---------------------------end base64 content--------------------------------\n")
	}

	return buffer.Bytes()
}

// isContainString 判断一个哈希值是否在数组中
func isContainString(arrays []IconHashResult, hash string) bool {
	for _, a := range arrays {
		if hash == a.Hash {
			return true
		}
	}
	return false
}
