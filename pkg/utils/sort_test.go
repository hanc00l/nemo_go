package utils

import "testing"

func TestSortMapByValue(t *testing.T) {
	test := map[string]int{"wang":1,"liang":4,"lin":2,"dd":2,"haha":10}
	t.Log(SortMapByValue(test,true))
}

func TestSetToSlince(t *testing.T) {
	test := map[string]struct{}{"wang":struct{}{},"liang":struct{}{},"lin":struct{}{},"dd":struct{}{},"haha":struct{}{}}
	t.Log(SetToSlice(test))
	t.Log(SetToString(test))
	test2 := map[string]struct{}{}
	t.Log(SetToSlice(test2))
	t.Log(SetToString(test2))

}