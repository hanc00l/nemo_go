package core

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/projectdiscovery/mapcidr"
	"math/big"
	"net"
	"sort"
	"strings"
)

type TaskSlice struct {
	Targets     string
	TaskMode    int
	SliceNumber int
}

// NewTaskSlice 创建一个新的TaskSlice实例
func NewTaskSlice(targets string, taskMode, sliceNumber int) *TaskSlice {
	return &TaskSlice{
		Targets:     targets,
		TaskMode:    taskMode,
		SliceNumber: sliceNumber,
	}
}

// SplitTargets 根据TaskMode和SliceNumber拆分Targets
func (ts *TaskSlice) SplitTargets() []string {
	// 按逗号分割目标字符串
	targetsSlice := strings.Split(ts.Targets, ",")

	switch ts.TaskMode {
	case 0:
		// 不拆分，直接返回包含整个Targets的切片
		return []string{ts.Targets}
	case 1:
		// 按行数拆分
		return splitByLineCount(targetsSlice, ts.SliceNumber)
	case 2:
		// 按IP数拆分
		return sliceByIP(targetsSlice, ts.SliceNumber)
	default:
		return []string{ts.Targets}
	}
}

// splitByLineCount 按行数拆分目标切片
func splitByLineCount(targets []string, sliceNumber int) []string {
	var result []string
	length := len(targets)

	for i := 0; i < length; i += sliceNumber {
		end := i + sliceNumber
		if end > length {
			end = length
		}
		// 将当前切片部分合并为逗号分隔的字符串
		slice := strings.Join(targets[i:end], ",")
		result = append(result, slice)
	}

	return result
}

// parseAllIP 解析所有的IP
func parseAllIP(targetList []string) (ipv4IntMap map[int]struct{}, ipv6BigIntMap map[*big.Int]struct{}) {
	ipv4IntMap = make(map[int]struct{})
	ipv6BigIntMap = make(map[*big.Int]struct{})
	for _, v := range targetList {
		ips := utils.ParseIP(v)
		for _, ip := range ips {
			if utils.CheckIPV4(ip) {
				ipv4Int := int(utils.IPV4ToUInt32(ip))
				if _, ok := ipv4IntMap[ipv4Int]; !ok {
					ipv4IntMap[ipv4Int] = struct{}{}
				}
			} else if utils.CheckIPV6(ip) {
				ipv6BigInt := utils.IPV6ToBigInt(ip)
				if _, ok := ipv6BigIntMap[ipv6BigInt]; !ok {
					ipv6BigIntMap[ipv6BigInt] = struct{}{}
				}
			}
		}
	}
	return
}

// sliceByIP 按等量对ip进行切分，同时将IP进行cidr聚合
func sliceByIP(targetsSlice []string, sliceNumber int) (ips []string) {
	ipv4Map, ipv6Map := parseAllIP(targetsSlice)
	//ipv4
	if len(ipv4Map) > 0 {
		ipIntList := utils.SetToSliceInt(ipv4Map)
		sort.Ints(ipIntList)
		segments := splitArrayInt(ipIntList, sliceNumber)
		for _, v := range segments {
			var ipList []string
			for _, ipInt := range v {
				ip := utils.UInt32ToIPV4(uint32(ipInt))
				ipList = append(ipList, ip)
			}
			ipCidrs, _ := aggregateCIDRs(ipList)
			ips = append(ips, ipCidrs)
		}
	}
	//ipv6:
	if len(ipv6Map) > 0 {
		ipBigIntList := utils.SetToSliceBigInt(ipv6Map)
		sort.Slice(ipBigIntList, func(i, j int) bool {
			return ipBigIntList[i].Cmp(ipBigIntList[j]) < 0
		})
		segmentsBigInt := splitArrayBigInt(ipBigIntList, sliceNumber)
		for _, v := range segmentsBigInt {
			var ipList []string
			for _, ipBigInt := range v {
				ip := utils.BigIntToIPV6(ipBigInt)
				ipList = append(ipList, ip)
			}
			_, ipCidrs := aggregateCIDRs(ipList)
			ips = append(ips, ipCidrs)
		}
	}
	return
}

// splitArrayInt 对数组分组
func splitArrayInt(arr []int, num int) (segments [][]int) {
	segments = make([][]int, 0)
	maxInt := len(arr)
	if maxInt <= num {
		segments = append(segments, arr)
		return
	}
	quantity := maxInt / num
	for i := 0; i <= quantity; i++ {
		start := i * num
		end := (i + 1) * num
		if i != quantity {
			segments = append(segments, arr[start:end])
		} else {
			if maxInt%num != 0 {
				segments = append(segments, arr[start:])
			}
		}
	}
	return
}

// splitArrayBigInt 对数组分组
func splitArrayBigInt(arr []*big.Int, num int) (segments [][]*big.Int) {
	segments = make([][]*big.Int, 0)
	m := len(arr)
	if m <= num {
		segments = append(segments, arr)
		return
	}
	quantity := m / num
	for i := 0; i <= quantity; i++ {
		start := i * num
		end := (i + 1) * num
		if i != quantity {
			segments = append(segments, arr[start:end])
		} else {
			if m%num != 0 {
				segments = append(segments, arr[start:])
			}
		}
	}
	return
}

// aggregateCIDRs 对IP进行cidr聚合，调用的是mapcidr
func aggregateCIDRs(ips []string) (ipv4Cidr, ipv6Cidr string) {
	var allCidrs []*net.IPNet
	for _, ip := range ips {
		var cidr string
		if utils.CheckIPV4(ip) {
			cidr = fmt.Sprintf("%s/32", ip)
		} else {
			cidr = fmt.Sprintf("%s/128", ip)
		}
		_, pCidr, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		allCidrs = append(allCidrs, pCidr)
	}
	cCidrsIPV4, cCidrsIPV6 := mapcidr.CoalesceCIDRs(allCidrs)
	var outputV4, outputV6 []string
	for _, cidrIPV4 := range cCidrsIPV4 {
		s := strings.ReplaceAll(cidrIPV4.String(), "/32", "")
		outputV4 = append(outputV4, s)
	}
	ipv4Cidr = strings.Join(outputV4, ",")
	for _, cidrIPV6 := range cCidrsIPV6 {
		s := strings.ReplaceAll(cidrIPV6.String(), "/128", "")
		outputV6 = append(outputV6, s)
	}
	ipv6Cidr = strings.Join(outputV6, ",")

	return
}
