package custom

// forked from https://github.com/zu1k/nali
// ipv6db数据使用http://ip.zxinc.org的免费离线数据（更新到2021年）

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/zu1k/nali/pkg/wry"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Ipv6Location struct {
	wry.IPDB[uint64]
}

func NewIPv6Location() (*Ipv6Location, error) {
	var fileData []byte
	dbFilePath := filepath.Join(conf.GetRootPath(), "thirdparty/zxipv6wry/ipv6wry.db")
	_, err := os.Stat(dbFilePath)
	if err != nil && os.IsNotExist(err) {
		if err != nil {
			return nil, errors.New("数据文件不存在")
		}
	} else {
		fileBase, err := os.OpenFile(dbFilePath, os.O_RDONLY, 0400)
		if err != nil {
			return nil, err
		}
		defer fileBase.Close()

		fileData, err = io.ReadAll(fileBase)
		if err != nil {
			return nil, err
		}
	}

	if !CheckFile(fileData) {
		log.Fatalln("ZX IPv6数据库存在错误，请重新下载")
	}

	header := fileData[:24]
	offLen := header[6]
	ipLen := header[7]

	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11

	return &Ipv6Location{
		IPDB: wry.IPDB[uint64]{
			Data: fileData,

			OffLen:   offLen,
			IPLen:    ipLen,
			IPCnt:    counts,
			IdxStart: start,
			IdxEnd:   end,
		},
	}, nil
}

func (db *Ipv6Location) FindByZX(query string, _ ...string) (result *wry.Result, err error) {
	ip := net.ParseIP(query)
	if ip == nil {
		return nil, errors.New("query should be IPv6")
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return nil, errors.New("query should be IPv6")
	}
	ip6 = ip6[:8]
	ipu64 := binary.BigEndian.Uint64(ip6)

	offset := db.SearchIndexV6(ipu64)
	reader := wry.NewReader(db.Data)
	reader.Parse(offset)

	return &reader.Result, nil
}

func (db *Ipv6Location) Find(query string) string {
	result, err := db.FindByZX(query)
	if err != nil || result == nil {
		return ""
	}
	r := strings.ReplaceAll(result.Country, "\t", " ")
	if len(result.Area) > 0 {
		for _, v := range cloudNameList {
			if strings.Index(result.Area, v) >= 0 {
				return fmt.Sprintf("%s [%s]", r, result.Area)
			}
		}
	}
	return r
}

func CheckFile(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if string(data[:4]) != "IPDB" {
		return false
	}

	if len(data) < 24 {
		return false
	}
	header := data[:24]
	start := binary.LittleEndian.Uint64(header[16:24])
	counts := binary.LittleEndian.Uint64(header[8:16])
	end := start + counts*11
	if start >= end || uint64(len(data)) < end {
		return false
	}

	return true
}
