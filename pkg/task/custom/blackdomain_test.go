package custom

import "testing"

func TestBlackDomain_CheckBlack(t *testing.T) {
	domains := []string{"www.qq.com", "smtpqq.com", "kk.gov.cn", "2.gov.cnn"}
	b := NewBlackDomain()
	for _, d := range domains {
		t.Log(d, b.CheckBlack(d))
	}
}
