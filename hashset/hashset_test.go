package hashset_test

import (
	"testing"

	"github.com/TprceOYX/collections/hashset"
)

func TestHashSet(t *testing.T) {
	s1 := []string{"1", "2", "3", "4"}
	s2 := []string{"3", "4", "5", "6"}
	strSet := hashset.NewHashSetWithSize[string](4)
	strSet.Add(s1...)
	strSet.Merge(hashset.NewHashSetFromSlice(s2))
	if strSet.Len() != 6 {
		t.Fatalf("wrong length of hashset, expected:%d, actual:%d", 6, strSet.Len())
	}
	for _, v := range s1 {
		if !strSet.Contain(v) {
			t.Fatalf("hashset not contain expected element: %s", v)
		}
	}
	count := 0
	strSet.Range(func(string) bool {
		count += 1
		return count < 2
	})
	if count != 2 {
		t.Fatalf("wrong count, expected:%d, actual:%d", 2, count)
	}

	strSet = strSet.Intersection(hashset.NewHashSetFromSlice(s1).Intersection(hashset.NewHashSetFromSlice(s2)))
	if strSet.Len() != 2 {
		t.Fatalf("wrong length of hashset, expected:%d, actual:%d", 2, strSet.Len())
	}
	for _, v := range []string{"3", "4"} {
		if !strSet.Contain(v) {
			t.Fatalf("hashset not contain expected element: %s", v)
		}
	}
	slice := strSet.ToSlice()
	if len(slice) != strSet.Len() {
		t.Fatalf("wrong length of slice, expected:%d, actual:%d", strSet.Len(), len(slice))
	}
}
