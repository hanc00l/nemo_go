package utils

import (
	"math/big"
	"sort"
	"strings"
)

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// SortMapByValue 对map[string]int按value排序，返回k,v的列表
func SortMapByValue(mapData map[string]int, descSort bool) (r PairList) {
	r = make(PairList, len(mapData))
	i := 0
	for k, v := range mapData {
		r[i] = Pair{k, v}
		i++
	}
	if descSort {
		//从大到小排序
		sort.Sort(sort.Reverse(r))
	} else {
		//从小到大排序
		sort.Sort(r)
	}
	return
}

// RemoveDuplicationElement 去除重复切片元素
func RemoveDuplicationElement(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}

// SetToSlice 将Set（由map模拟实现）结果转化为列表结果
func SetToSlice(setMap map[string]struct{}) (list []string) {
	list = make([]string, len(setMap))
	i := 0
	for k, _ := range setMap {
		list[i] = k
		i++
	}
	return
}

// SetToString 将Set（由map模拟实现）结果转化为拼结的字符
func SetToString(setMap map[string]struct{}) string {
	list := SetToSlice(setMap)
	return strings.Join(list, ",")
}

// SetToSliceInt 将Set（由map模拟实现）结果转化为列表结果
func SetToSliceInt(setMap map[int]struct{}) (list []int) {
	list = make([]int, len(setMap))
	i := 0
	for k, _ := range setMap {
		list[i] = k
		i++
	}
	return
}

// SetToSliceUInt 将Set（由map模拟实现）结果转化为列表结果
func SetToSliceUInt(setMap map[uint32]struct{}) (list []uint32) {
	list = make([]uint32, len(setMap))
	i := 0
	for k, _ := range setMap {
		list[i] = k
		i++
	}
	return
}

// SetToSliceBigInt 将Set（由map模拟实现）结果转化为列表结果
func SetToSliceBigInt(setMap map[*big.Int]struct{}) (list []*big.Int) {
	list = make([]*big.Int, len(setMap))
	i := 0
	for k, _ := range setMap {
		list[i] = k
		i++
	}
	return
}

// SetToSliceStringInt 将Set（由map模拟实现）结果转化为列表结果
func SetToSliceStringInt(setMap map[string]int) (list []string) {
	list = make([]string, len(setMap))
	i := 0
	for k, _ := range setMap {
		list[i] = k
		i++
	}
	return
}

func MergeMapStringInt(dstMap map[string]int, srcMap map[string]int) {
	for k, v := range srcMap {
		if _, ok := dstMap[k]; !ok {
			dstMap[k] = v
		} else {
			dstMap[k] += v
		}
	}
	return
}
