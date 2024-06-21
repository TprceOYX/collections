package hashset

type HashSet[T comparable] map[T]struct{}

// NewHashSet return an empty hashset
func NewHashSet[T comparable]() HashSet[T] {
	return make(map[T]struct{})
}

// NewHashSet return an empty hashset with given initial size
func NewHashSetWithSize[T comparable](size int) HashSet[T] {
	return make(map[T]struct{}, size)
}

// NewHashSet returns a hashset containing the unique elements in the given slice.
func NewHashSetFromSlice[T comparable](slice []T) HashSet[T] {
	set := make(map[T]struct{}, len(slice))
	for _, v := range slice {
		set[v] = struct{}{}
	}
	return set
}

// Len returns the number of elements in this hashset
func (s HashSet[T]) Len() int {
	return len(s)
}

// Add adds a new element to this set
func (s HashSet[T]) Add(elements ...T) {
	for _, v := range elements {
		s[v] = struct{}{}
	}
}

func (s HashSet[T]) Contain(ele T) bool {
	_, exist := s[ele]
	return exist
}

// Range calls f sequentially for each element present in the hashset.
// If f returns false, range stops the iteration.
func (s HashSet[T]) Range(f func(ele T) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Merge sequentially traverses the given hashsets and merges them into this set one by one
// Will directly update this set instead of creating a new hashset to store the results
func (s HashSet[T]) Merge(others ...HashSet[T]) HashSet[T] {
	for _, set := range others {
		for k := range set {
			s[k] = struct{}{}
		}
	}
	return s
}

// Intersection returns a new hashset containing the same elements as this hashset and the given hashset.
func (s HashSet[T]) Intersection(other HashSet[T]) HashSet[T] {
	var res, rangeSet, compareSet HashSet[T]

	if s.Len() > other.Len() {
		res = NewHashSetWithSize[T](other.Len())
		rangeSet, compareSet = other, s
	} else {
		res = NewHashSetWithSize[T](s.Len())
		rangeSet, compareSet = s, other
	}
	for k := range rangeSet {
		if _, exist := compareSet[k]; exist {
			res[k] = struct{}{}
		}
	}
	return res
}

// ToSlice converts hashset to slice
func (s HashSet[T]) ToSlice() []T {
	res := make([]T, 0, s.Len())
	for k := range s {
		res = append(res, k)
	}
	return res
}
