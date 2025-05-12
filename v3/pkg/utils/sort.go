package utils

import (
	"math/big"
	"sort"
	"strings"
	"time"
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

func SortTimeMap(m map[time.Time]string, isDesc bool) []string {
	var keys []time.Time
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if isDesc {
			return keys[i].After(keys[j])
		} else {
			return keys[i].Before(keys[j])
		}
	})
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = m[key]
	}

	return result
}

func CompareStringSlices(s1, s2 []string, isSort bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	if isSort {
		// 对两个切片进行排序
		sort.Strings(s1)
		sort.Strings(s2)
	}
	// 比较排序后的切片
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// RemoveDuplicatesAndSort 将字符串数组转换为小写，去除重复元素并排序
func RemoveDuplicatesAndSort(arr []string) []string {
	// 使用map来记录已经出现过的字符串
	seen := make(map[string]bool)
	var result []string

	// 遍历数组，将每个字符串转换为小写
	for _, str := range arr {
		lowerStr := strings.ToLower(str)
		if !seen[lowerStr] {
			seen[lowerStr] = true
			result = append(result, lowerStr)
		}
	}

	// 对结果数组进行排序
	sort.Strings(result)
	return result
}
