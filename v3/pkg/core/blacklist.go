package core

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/pkg/errors"
	"github.com/yl2chen/cidranger"
	"net"
	"strings"
	"testing"
)

type Blacklist struct {
	ipRanger   cidranger.Ranger
	domainTrie *TrieNode
	blackCIDRs []*net.IPNet
}

type TrieNode struct {
	children map[string]*TrieNode
	isEnd    bool
}

func NewBlacklist() *Blacklist {
	return &Blacklist{
		ipRanger:   cidranger.NewPCTrieRanger(),
		domainTrie: &TrieNode{children: make(map[string]*TrieNode)},
		blackCIDRs: make([]*net.IPNet, 0),
	}
}

func IsCIDR(ip string) bool {
	_, _, err := net.ParseCIDR(ip)
	return err == nil
}

func (b *Blacklist) LoadBlacklist(workspaceId string) bool {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	defer db.CloseClient(mongoClient)

	cd := db.NewCustomData(workspaceId, mongoClient)
	docs, _ := cd.Find(db.CategoryBlacklist)
	for _, doc := range docs {
		lines := strings.Split(doc.Data, "\n")
		for _, data := range lines {
			ipOrDomain := strings.TrimSpace(data)
			if len(ipOrDomain) == 0 {
				continue
			}
			if strings.HasPrefix(ipOrDomain, "#") {
				continue
			}
			domainArray := strings.Split(ipOrDomain, " ")
			if len(domainArray) > 1 {
				ipOrDomain = domainArray[0]
			}
			if utils.CheckIPOrSubnet(ipOrDomain) {
				err = b.AddIP(ipOrDomain)
				if err != nil {
					logging.RuntimeLog.Error(err)
				}
			} else {
				b.AddDomain(ipOrDomain)
			}
		}
	}
	return true
}

func (b *Blacklist) AddIP(ip string) error {
	if IsCIDR(ip) {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return err
		}
		b.ipRanger.Insert(cidranger.NewBasicRangerEntry(*ipNet))
		b.blackCIDRs = append(b.blackCIDRs, ipNet)
	} else {
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			return errors.Errorf("invalid IP address: %s", ip)
		}
		var cidr string
		if ipAddr.To4() != nil {
			cidr = fmt.Sprintf("%s/32", ip)
		} else {
			cidr = fmt.Sprintf("%s/128", ip)
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		b.ipRanger.Insert(cidranger.NewBasicRangerEntry(*ipNet))
		b.blackCIDRs = append(b.blackCIDRs, ipNet)
	}
	return nil
}

func (b *Blacklist) AddDomain(domain string) {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	node := b.domainTrie
	for _, part := range parts {
		if _, exists := node.children[part]; !exists {
			node.children[part] = &TrieNode{children: make(map[string]*TrieNode)}
		}
		node = node.children[part]
	}
	node.isEnd = true
}

func (b *Blacklist) IsIPBlocked(ip string) bool {
	if IsCIDR(ip) {
		_, inputIPNet, err := net.ParseCIDR(ip)
		if err != nil || inputIPNet == nil {
			logging.RuntimeLog.Warningf("Invalid CIDR: %s", ip)
			return false
		}
		for _, blackNet := range b.blackCIDRs {
			onesInput, _ := inputIPNet.Mask.Size()
			onesBlack, _ := blackNet.Mask.Size()
			if blackNet.Contains(inputIPNet.IP) && onesBlack <= onesInput {
				return true
			}
		}
		return false
	} else {
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			return false
		}
		isBlocked, _ := b.ipRanger.Contains(ipAddr)
		return isBlocked
	}
}

func (b *Blacklist) IsDomainBlocked(domain string) bool {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	node := b.domainTrie
	for _, part := range parts {
		if node.children[part] == nil {
			return false
		}
		if node.children[part].isEnd {
			return true
		}
		node = node.children[part]
	}
	return node.isEnd
}

func (b *Blacklist) IsHostBlocked(host string) bool {
	if len(host) == 0 {
		return false
	}
	if utils.CheckIPOrSubnet(host) {
		return b.IsIPBlocked(host)
	} else {
		return b.IsDomainBlocked(host)
	}
}

func Test(t *testing.T) {
	blacklist := NewBlacklist()
	err := blacklist.AddIP("192.168.1.0/25")
	err = blacklist.AddIP("192.168.2.1")
	err = blacklist.AddIP("2001:db8::/32")
	err = blacklist.AddIP("192.168.120.1/25")
	if err != nil {
		t.Log("Error adding IP to blacklist:", err)
		return
	}
	blacklist.AddDomain("example.com")

	// 检查IP是否被封锁
	t.Log(blacklist.IsIPBlocked("192.168.1.1"))      // true
	t.Log(blacklist.IsIPBlocked("192.168.1.16/28"))  // true
	t.Log(blacklist.IsIPBlocked("192.168.2.2"))      //  false
	t.Log(blacklist.IsIPBlocked("192.168.120.1/28")) //true
	t.Log(blacklist.IsIPBlocked("2001:db8::/40"))    // true
	t.Log(blacklist.IsIPBlocked("2001:db9::/32"))    // false
	// 检查域名是否被封锁
	t.Log(blacklist.IsDomainBlocked("example.com"))        // true
	t.Log(blacklist.IsDomainBlocked("sub.example.com"))    // true
	t.Log(blacklist.IsDomainBlocked("sub.example.com.cn")) // false
	t.Log(blacklist.IsDomainBlocked("anotherdomain.com"))  // false
}
