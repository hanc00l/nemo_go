package conf

import (
	"strings"
	"testing"
)

func Test1(t *testing.T) {
	s1 := "github.com/1/s.1"
	s2 := ""
	a1:= strings.Split(s1,"/")
	t.Log(a1[len(a1)-1])
	a2:= strings.Split(s2,"/")
	t.Log(a2[len(a2)-1])
}