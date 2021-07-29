package custom

import "testing"

func TestCustomService_FindService(t *testing.T) {
	service := Service{}

	t.Log(service.FindService(443,""))
	t.Log(service.FindService(3306,""))
	t.Log(service.FindService(18080,""))
	t.Log(service.FindService(18080,"10.1.2.3"))
}
