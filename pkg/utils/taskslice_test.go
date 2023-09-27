package utils

import (
	"testing"
)

func TestNewTaskSlice(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = 2
	ts.Port = "80,8080,443"
	ts.IpTarget = []string{"1.1.1.1", "172.16.80.0/30", "192.168.2.0/25", "10.185.192.0/24", "223.224.225.226"}
	ts.IpSliceNumber = 100
	target, port := ts.DoIpSlice()
	for _, v := range target {
		t.Log(v)
	}
	t.Log(port)
}

func TestNewTaskSlice1(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = 2
	ts.Port = "80,8080,443"
	ts.IpTarget = []string{"10.185.192.0/24", "223.224.225.226"}
	ts.IpSliceNumber = 64
	target, port := ts.DoIpSlice()
	for _, v := range target {
		t.Log(v)
	}
	t.Log(port)
}

func TestNewTaskSlice2(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = SliceByPort
	ts.Port = "--top-ports 1000"
	ts.IpTarget = []string{"1.1.1.1", "172.16.80.0/30", "192.168.2.0/25", "10.185.192.0/24", "223.224.225.226"}
	ts.PortSliceNumber = 100
	target, port := ts.DoIpSlice()
	for _, v := range port {
		t.Log(v)
	}
	t.Log(target)
}

func TestNewTaskSlice3(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = SliceByIPAndPort
	ts.Port = "--top-ports 100"
	ts.IpTarget = []string{"1.1.1.1", "172.16.80.0/30", "192.168.2.0/25", "10.185.192.0/24", "223.224.225.226"}
	ts.PortSliceNumber = 50
	ts.IpSliceNumber = 128
	target, port := ts.DoIpSlice()
	for _, i := range target {
		for _, p := range port {
			t.Log("----------------------")
			t.Log(i)
			t.Log(p)
		}
	}
}

func TestNewTaskSliceIPV6(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = 2
	ts.Port = "80,8080,443"
	ts.IpTarget = []string{"2409:8929:42d:bf31:1840:27ba:d669::/120"}
	ts.IpSliceNumber = 64
	target, port := ts.DoIpSlice()
	for _, v := range target {
		t.Log(v)
	}
	t.Log(port)
}

func TestNewTaskSliceIPV46(t *testing.T) {
	ts := NewTaskSlice()
	ts.TaskMode = 2
	ts.Port = "80,8080,443"
	ts.IpTarget = []string{"10.185.192.0/24", "2409:8929:42d:bf31:1840:27ba:d669::/120"}
	ts.IpSliceNumber = 64
	target, port := ts.DoIpSlice()
	for _, v := range target {
		t.Log(v)
	}
	t.Log(port)
}
