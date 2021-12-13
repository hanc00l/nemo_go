package fingerprint

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/mat/besticon/besticon"
	"github.com/remeh/sizedwaitgroup"
	"github.com/twmb/murmur3"
	"hash"
	"image"
	"strings"
)

type IconHash struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
}

type IconHashResult struct {
	Url  string
	Hash string
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
	swg := sizedwaitgroup.New(fpIconHashThreadNumber)

	if i.ResultPortScan.IPResult != nil {
		bport := make(map[int]struct{})
		for _, p := range IgnorePort {
			bport[p] = struct{}{}
		}
		for ipName, ipResult := range i.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := bport[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					iconHashes := i.RunFetchIconHashes(u)
					if len(iconHashes) > 0 {
						for _, r := range iconHashes {
							par := portscan.PortAttrResult{
								Source:  "iconhash",
								Tag:     "favicon",
								Content: fmt.Sprintf("%s | %s", r.Hash, r.Url),
							}
							i.ResultPortScan.SetPortAttr(ip, port, par)
						}
					}
					swg.Done()
				}(ipName, portNumber, url)
			}
		}
	}
	if i.ResultDomainScan.DomainResult != nil {
		for domain, _ := range i.ResultDomainScan.DomainResult {
			swg.Add()
			go func(d string) {
				iconHashes := i.RunFetchIconHashes(d)
				if len(iconHashes) > 0 {
					for _, r := range iconHashes {
						dar := domainscan.DomainAttrResult{
							Source:  "iconhash",
							Tag:     "favicon",
							Content: fmt.Sprintf("%s | %s", r.Hash, r.Url),
						}
						i.ResultDomainScan.SetDomainAttr(d, dar)
					}
				}
				swg.Done()
			}(domain)
		}
	}
	swg.Wait()
}

func (i *IconHash) RunFetchIconHashes(url string) (hashResult []IconHashResult) {
	var icons []besticon.Icon
	//遍历http与https两种协议
	for _, protol := range []string{"http", "https"} {
		// 获取缺省的icon:
		u1 := fmt.Sprintf("%s://%s/favicon.ico", protol, url)
		icon1 := fetchIconDetails(u1)
		if icon1.Error == nil {
			icons = append(icons, icon1)
		}
		// 获取网页中定义的全部icon
		u2 := fmt.Sprintf("%s://%s", protol, url)
		finder := besticon.IconFinder{}
		icons2, err := finder.FetchIcons(u2)
		if err == nil {
			icons = append(icons, icons2...)
		}
	}
	//计算哈希值
	for _, icon := range icons {
		hash := mmh3Hash32(standBase64(icon.ImageData))
		if isContainString(hashResult, hash) == false {
			hashResult = append(hashResult, IconHashResult{Url: icon.URL, Hash: hash})
		}
	}
	return
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
