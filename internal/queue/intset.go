package queue

import (
	"sync"
)

type IntSet interface {
	Add(int) bool
	Has(int) bool
	Clear()
}

type intSet struct {
	MU     sync.RWMutex
	Values map[int]struct{}
}

func newIntSet() intSet {
	return intSet{
		Values: make(map[int]struct{}),
	}
}

func (s *intSet) Has(v int) bool {
	s.MU.Lock()
	defer s.MU.Unlock()

	_, exists := s.Values[v]
	return exists
}

func (s *intSet) Add(v int) bool {
	s.MU.Lock()
	defer s.MU.Unlock()

	if s.Has(v) {
		return false
	}

	s.Values[v] = struct{}{}
	return true
}

func (s *intSet) GetSlice() []int {
	s.MU.Lock()
	defer s.MU.Unlock()

	out := make([]int, 0, len(s.Values))
	for v := range s.Values {
		out = append(out, v)
	}

	return out
}

func (s *intSet) Clear() {
	s.MU.Lock()
	defer s.MU.Unlock()

	s.Values = make(map[int]struct{})
}
