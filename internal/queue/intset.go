package queue

import (
	"sync"
)

// IntSet defines the interface for thread-safe integer set operations.
// Provides basic set functionality with concurrent access guarantees
// for distributed systems message tracking and deduplication.
type IntSet interface {
	// Add inserts an integer into the set, returns true if newly added
	Add(int) bool
	
	// Has checks if an integer exists in the set
	Has(int) bool
	
	// Clear removes all elements from the set
	Clear()
}

// intSet implements a thread-safe integer set using a map and RWMutex.
// Forms the foundation for both Messages and Peer queue implementations
// through composition, providing shared concurrent access patterns.
type intSet struct {
	// MU provides read-write mutex protection for concurrent access
	// Uses RWMutex to allow multiple concurrent readers when appropriate
	MU sync.RWMutex
	
	// Values stores integers as map keys with empty struct values
	// Empty struct{} uses zero memory, making this memory-efficient
	Values map[int]struct{}
}

// newIntSet creates a new thread-safe integer set with initialized map.
// Returns intSet by value since it will be embedded in other structs.
// The mutex is zero-initialized and ready for concurrent use.
func newIntSet() intSet {
	return intSet{
		Values: make(map[int]struct{}),
	}
}

// Has checks if the given integer exists in the set.
// Uses write lock for simplicity, though read lock could be used here.
// Returns true if the value exists, false otherwise.
func (s *intSet) Has(v int) bool {
	s.MU.RLock()
	defer s.MU.RUnlock()

	_, exists := s.Values[v]
	return exists
}

// Add inserts an integer into the set if it doesn't already exist.
// Returns true if the value was newly added, false if it already existed.
// Uses write lock to ensure atomic check-and-insert operation.
func (s *intSet) Add(v int) bool {
	s.MU.Lock()
	defer s.MU.Unlock()

	if _, exists := s.Values[v]; exists {
		return false
	}

	s.Values[v] = struct{}{}
	return true
}

// GetSlice returns all integers in the set as a slice.
// Creates a new slice with pre-allocated capacity for efficiency.
// Order of elements is not guaranteed due to map iteration semantics.
func (s *intSet) GetSlice() []int {
	s.MU.Lock()
	defer s.MU.Unlock()

	out := make([]int, 0, len(s.Values))
	for v := range s.Values {
		out = append(out, v)
	}

	return out
}

// Clear removes all elements from the set by creating a new empty map.
// More efficient than iterating and deleting individual elements.
// Uses write lock to ensure atomic operation during map replacement.
func (s *intSet) Clear() {
	s.MU.Lock()
	defer s.MU.Unlock()
	
	s.Values = make(map[int]struct{})
}