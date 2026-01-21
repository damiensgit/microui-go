package microui

// PoolItem represents an item in a microui-style pool.
// This is compatible with the original microui mu_PoolItem struct.
type PoolItem struct {
	ID         ID
	LastUpdate int // Frame number when last updated
}

// growStack is a simple stack with pre-allocated capacity.
type growStack[T any] struct {
	items []T
}

func (s *growStack[T]) Init(capacity int) {
	s.items = make([]T, 0, capacity)
}

func (s *growStack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *growStack[T]) Pop() T {
	n := len(s.items)
	if n == 0 {
		var zero T
		return zero
	}
	v := s.items[n-1]
	s.items = s.items[:n-1]
	return v
}

func (s *growStack[T]) Peek() T {
	n := len(s.items)
	if n == 0 {
		var zero T
		return zero
	}
	return s.items[n-1]
}

func (s *growStack[T]) Len() int {
	return len(s.items)
}

func (s *growStack[T]) Reset() {
	s.items = s.items[:0]
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
