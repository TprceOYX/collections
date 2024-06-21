package internal

type number interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

func Min[T number](a, b T) T {
	if a > b {
		return b
	}
	return a
}
