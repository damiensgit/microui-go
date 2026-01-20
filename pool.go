package microui

import (
	"log"
	"sync/atomic"
)

// PoolItem represents an item in a microui-style pool.
// This is compatible with the original microui mu_PoolItem struct.
type PoolItem struct {
	ID         ID
	LastUpdate int // Frame number when last updated
}

// growPool provides fixed initial capacity with graceful growth.
type growPool[T any] struct {
	fixed []poolSlot[T]
	grow  []poolSlot[T]
	max   int
}

type poolSlot[T any] struct {
	value T
	used  atomic.Bool
}

// Init initializes the pool with fixed size and maximum size.
func (p *growPool[T]) Init(fixedSize, maxSize int) {
	p.fixed = make([]poolSlot[T], fixedSize)
	p.max = maxSize
}

// Alloc returns a pointer to a slot in the pool.
// Fixed pool is used first (zero alloc), then grows with warning.
func (p *growPool[T]) Alloc() *T {
	// Try fixed pool first (zero alloc)
	for i := range p.fixed {
		if !p.fixed[i].used.Load() {
			if p.fixed[i].used.CompareAndSwap(false, true) {
				return &p.fixed[i].value
			}
		}
	}

	// Check if we can grow
	currentTotal := len(p.fixed) + len(p.grow)
	if currentTotal >= p.max {
		log.Panicf("pool exhausted: %d >= %d", currentTotal, p.max)
	}

	// Fall back to growth with warning
	slot := poolSlot[T]{used: atomic.Bool{}}
	slot.used.Store(true)
	p.grow = append(p.grow, slot)
	if len(p.grow) == 1 {
		log.Printf("warning: microui pool started growing beyond fixed size")
	}
	return &p.grow[len(p.grow)-1].value
}

// Len returns the number of allocated slots.
func (p *growPool[T]) Len() int {
	count := 0
	for i := range p.fixed {
		if p.fixed[i].used.Load() {
			count++
		}
	}
	return count + len(p.grow)
}

// growStack is similar to growPool but stack-ordered.
type growStack[T any] struct {
	items []T
	count int
}

func (s *growStack[T]) Init(capacity int) {
	s.items = make([]T, 0, capacity)
}

func (s *growStack[T]) Push(v T) {
	s.items = append(s.items, v)
	s.count++
}

func (s *growStack[T]) Pop() T {
	if s.count == 0 {
		var zero T
		return zero
	}
	s.count--
	v := s.items[s.count]
	s.items = s.items[:s.count]
	return v
}

func (s *growStack[T]) Peek() T {
	if s.count == 0 {
		var zero T
		return zero
	}
	return s.items[s.count-1]
}

func (s *growStack[T]) Len() int {
	return s.count
}

func (s *growStack[T]) Reset() {
	s.items = s.items[:0]
	s.count = 0
}

// PoolInit initializes a pool slot for the given ID.
// Finds the least-recently-used slot and assigns the ID to it.
// Returns the index of the slot.
// This matches the original microui mu_pool_init function.
func (u *UI) PoolInit(items []PoolItem, id ID) int {
	// Find least recently updated slot
	minFrame := u.frame
	minIdx := 0
	for i := range items {
		if items[i].LastUpdate < minFrame {
			minFrame = items[i].LastUpdate
			minIdx = i
		}
	}

	items[minIdx].ID = id
	u.PoolUpdate(items, minIdx)
	return minIdx
}

// PoolGet finds an item by ID in the pool.
// Returns the index if found, -1 if not found.
// This matches the original microui mu_pool_get function.
func (u *UI) PoolGet(items []PoolItem, id ID) int {
	for i := range items {
		if items[i].ID == id {
			return i
		}
	}
	return -1
}

// PoolUpdate marks a pool item as used this frame.
// This matches the original microui mu_pool_update function.
func (u *UI) PoolUpdate(items []PoolItem, idx int) {
	items[idx].LastUpdate = u.frame
}
