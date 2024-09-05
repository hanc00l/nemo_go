package custom

import (
	"fmt"
	"testing"
)

func TestBlackTargetCheck_Check(t *testing.T) {
	ips := []string{"127.0.0.1", "172.16.3.1", "114.114.114.115", "192.168.1.65"}
	domains := []string{"www.qq.com", "smtpqq.com", "kk.gov.cn", "2.gov.cnn"}

	ipCheck := NewBlackTargetCheck(CheckIP)
	for _, ip := range ips {
		t.Log(ip, ipCheck.CheckBlack(ip))
	}

	domainCheck := NewBlackTargetCheck(CheckDomain)
	for _, d := range domains {
		t.Log(d, domainCheck.CheckBlack(d))
	}

	fmt.Println()

	domainCheckAll := NewBlackTargetCheck(CheckAll)
	allTarget := append(ips, domains...)
	for _, d := range allTarget {
		t.Log(d, domainCheckAll.CheckBlack(d))
	}
}
