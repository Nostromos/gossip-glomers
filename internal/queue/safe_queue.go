package queue

import (
	"sync"
)

type Safe struct {
	MU     sync.RWMutex
	Values map[int]struct{}
}

func (q *Safe) Add(v int) bool {
	q.MU.Lock()
	defer q.MU.Unlock()

	if _, exists := q.Values[v]; exists {
		return false
	}
	q.Values[v] = struct{}{}
	return true
}

func (q *Safe) Has(v int) bool {
	q.MU.RLock()
	defer q.MU.RUnlock()

	for i := range q.Values {
		if i == v {
			return true
		}
	}

	return false
}

func (q *Safe) Snapshot() []int {
	q.MU.RLock()
	defer q.MU.RUnlock()

	out := make([]int, 0, len(q.Values))
	for v := range q.Values {
		out = append(out, v)
	}
	return out
}

func (q *Safe) Clear() {
	q.MU.Lock()
	defer q.MU.Unlock()

	q.Values = make(map[int]struct{})
}
