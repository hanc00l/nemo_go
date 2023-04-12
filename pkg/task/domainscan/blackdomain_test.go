package domainscan

import "testing"

func TestBlackDomain_CheckBlank(t *testing.T) {
	domains := []string{"www.qq.com", "smtpqq.com", "kk.gov.cn", "2.gov.cnn"}
	b := NewBlankDomain()
	for _, d := range domains {
		t.Log(d, b.CheckBlank(d))
	}
}
