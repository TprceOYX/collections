package skipset_test

import (
	"reflect"
	"testing"

	"github.com/TprceOYX/collections/skipset"
)

func TestSkipSet(t *testing.T) {
	originData := []int{1, 3, 4, 5, 6, 9}
	shuffleData := []int{5, 9, 1, 3, 6, 4}
	set := skipset.NewSkipSet[int](false)
	for _, v := range shuffleData {
		if !set.Add(v) {
			t.Fatalf("expect add ele: %d success", v)
		}
	}
	reflect.DeepEqual(originData, set.TopN(100))
	for i, v := range originData {
		ele := set.Index(i)
		if ele != v {
			t.Fatalf("expect: %d, actural: %d", v, ele)
		}
	}
}
