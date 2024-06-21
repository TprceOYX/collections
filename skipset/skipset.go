package skipset

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
)

const maxLevel = 16

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

type SkipSet[T ordered] struct {
	head         *skipsetNode[T]
	length       int64
	highestLevel uint32
	desc         bool
}

func NewSkipSet[T ordered](desc bool) *SkipSet[T] {
	return &SkipSet[T]{
		head: newSkipsetNode[T](maxLevel),
		desc: desc,
	}
}

type skipsetNode[T ordered] struct {
	value  T
	levels []*levelNode[T]
	mutex  sync.Mutex
}

type levelNode[T ordered] struct {
	next atomic.Pointer[skipsetNode[T]]
	span uint32
}

func newSkipsetNode[T ordered](nl int) *skipsetNode[T] {
	levels := make([]*levelNode[T], nl)
	for i := 0; i < nl; i++ {
		levels[i] = new(levelNode[T])
	}
	return &skipsetNode[T]{
		value:  *new(T),
		levels: levels,
		mutex:  sync.Mutex{},
	}
}

// getNext returns the next node at the given level.
func (n *skipsetNode[T]) getNext(level int) *skipsetNode[T] {
	return n.levels[level].next.Load()
}

func (n *skipsetNode[T]) getNextSpan(level int) int {
	return int(atomic.LoadUint32(&n.levels[level].span))
}

// setNext sets the next node at the given level.
func (n *skipsetNode[T]) setNext(level int, next *skipsetNode[T]) {
	n.levels[level].next.Store(next)
}

func (n *skipsetNode[T]) setNextStep(level int, step uint32) {
	atomic.StoreUint32(&n.levels[level].span, step)
}

func (n *skipsetNode[T]) incrNextStep(level int) {
	atomic.AddUint32(&n.levels[level].span, 1)
}

// Add adds a value to the skipset, returns true if the value was added,
// returns false if the value was already in the skipset.
func (list *SkipSet[T]) Add(value T) bool {
	var (
		newLevel                       = list.randomLevel()
		prevNodes, nextNodes           [maxLevel]*skipsetNode[T]
		prevNodesShift, nextNodesShift [maxLevel]uint32
	)
	for {
		// 找到插入位置
		var (
			node         = list.head
			highestLevel = int(atomic.LoadUint32(&list.highestLevel))
			shift        uint32 // 和head的距离
		)
		for i := highestLevel - 1; i >= 0; i-- { // 垂直方向向下寻找
			next := node.getNext(i)
			for next != nil && ((next.value < value && !list.desc) || (next.value > value && list.desc)) { // 水平方向向前寻找
				shift += uint32(node.getNextSpan(i)) // 记录向前查找的步进
				node = next
				next = node.getNext(i)
			}
			prevNodes[i], nextNodes[i] = node, next
			prevNodesShift[i], nextNodesShift[i] = shift, shift+uint32(node.getNextSpan(i))
			if next != nil && next.value == value { // 已经存在，不做插入操作
				return false
			}
		}
		// 自底向上给所有涉及到的node加锁
		var (
			lastPrevIndex int = -1 // 上一个加锁节点在链表中的位置
			stateChange       = false
		)
		for i := 0; i < newLevel; i++ {
			prev, next := prevNodes[i], nextNodes[i]
			prevIndex := int(prevNodesShift[i])
			if lastPrevIndex != prevIndex { // 给新节点加锁
				lastPrevIndex = prevIndex
				prev.mutex.Lock()
			}
			// 其它goroutine也在添加node，重新尝试
			stateChange = prev.getNext(i) != next
			if stateChange {
				break
			}
		}
		if stateChange {
			// release lock and retry
			lastPrevIndex = -1
			for i := newLevel - 1; i >= 0; i-- {
				prevIndex := int(prevNodesShift[i])
				if lastPrevIndex != prevIndex {
					lastPrevIndex = prevIndex
					prevNodes[i].mutex.Unlock()
				}
			}
			continue
		}
		newNode := newSkipsetNode[T](newLevel)
		newNode.value = value
		shift += 1
		// 从最底层向上，将新节点插入链表
		for i := 0; i < newLevel; i++ {
			prev, next := prevNodes[i], nextNodes[i]
			newNode.setNext(i, next)
			newNode.setNextStep(i, nextNodesShift[i]+1-shift)
			prev.setNextStep(i, shift-prevNodesShift[i])
			prev.setNext(i, newNode)
		}
		for i := newLevel; i < highestLevel; i++ {
			prev, next := prevNodes[i], nextNodes[i]
			if next == nil {
				break
			}
			prev.incrNextStep(newLevel)
		}
		lastPrevIndex = -1
		for i := newLevel - 1; i >= 0; i-- {
			prevIndex := int(prevNodesShift[i])
			if lastPrevIndex != prevIndex {
				lastPrevIndex = prevIndex
				prevNodes[i].mutex.Unlock()
			}
		}
		return true
	}
}

// Remove a value from the skipset, returns true if the value was found and removed.
// TODO implement this.
func (list *SkipSet[T]) Remove(value T) bool {
	return true
}

// Contains checks if the value is in the skipset.
func (list *SkipSet[T]) Contains(value T) bool {
	node := list.head

	for i := int(atomic.LoadUint32(&list.highestLevel)) - 1; i >= 0; i-- {
		next := node.getNext(i)
		for next != nil && next.value < value {
			node = next
			next = node.getNext(i)
		}

		if next != nil && next.value == value {
			return true
		}
	}
	return false
}

// TopN returns the top n elements in skipset in the order specified by desc
func (list *SkipSet[T]) TopN(n int) []T {
	res := make([]T, 0, n)
	node := list.head
	next := node.getNext(0)
	for i := 0; i < n && next != nil; i++ {
		res = append(res, next.value)
		node = next
		next = next.getNext(0)
	}
	return res
}

// Index returns the element at the given index, the index is 0-based.
func (list *SkipSet[T]) Index(index int) T {
	var (
		length int = -1
		node       = list.head
	)
	for i := int(atomic.LoadUint32(&list.highestLevel)) - 1; i >= 0; i-- {
		next := node.getNext(i)
		for next != nil {
			step := node.getNextSpan(i)
			if step+length > index {
				break
			}
			length += step
			node = next
			next = node.getNext(i)
		}
		if length == index {
			break
		}
	}
	return node.value
}

// ToSlice returns a slice of all the elements in the skipset.
func (list *SkipSet[T]) ToSlice() []T {
	return list.TopN(list.Len())
}

// Len returns the number of elements in the skipset.
func (list *SkipSet[T]) Len() int {
	return int(atomic.LoadInt64(&list.length))
}

// randomLevel returns a random level for a new node.
func (list *SkipSet[T]) randomLevel() int {
	const threshold = math.MaxInt32 / 4
	var level uint32 = 1
	for rand.Int31() < threshold {
		level++
	}
	for {
		highestLevel := atomic.LoadUint32(&list.highestLevel)
		if highestLevel >= level {
			break
		}
		if atomic.CompareAndSwapUint32(&list.highestLevel, highestLevel, level) {
			break
		}
	}
	return int(level)
}
