package queue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntSet_NewIntSet(t *testing.T) {
	s := newIntSet()
	assert.NotNil(t, s.Values)
	assert.Empty(t, s.Values)
}

func TestIntSet_Add(t *testing.T) {
	s := newIntSet()
	
	// First add should return true
	added := s.Add(42)
	assert.True(t, added)
	assert.True(t, s.Has(42))
	
	// Duplicate add should return false
	added = s.Add(42)
	assert.False(t, added)
	assert.True(t, s.Has(42))
}

func TestIntSet_Has(t *testing.T) {
	s := newIntSet()
	
	// Should not have non-existent value
	assert.False(t, s.Has(99))
	
	// Should have added value
	s.Add(99)
	assert.True(t, s.Has(99))
}

func TestIntSet_GetSlice(t *testing.T) {
	s := newIntSet()
	
	// Empty set should return empty slice
	slice := s.GetSlice()
	assert.Empty(t, slice)
	assert.NotNil(t, slice)
	
	// Add some values
	s.Add(1)
	s.Add(3)
	s.Add(2)
	
	slice = s.GetSlice()
	assert.Len(t, slice, 3)
	assert.Contains(t, slice, 1)
	assert.Contains(t, slice, 2)
	assert.Contains(t, slice, 3)
}

func TestIntSet_Clear(t *testing.T) {
	s := newIntSet()
	
	// Add some values
	s.Add(1)
	s.Add(2)
	s.Add(3)
	assert.True(t, s.Has(1))
	assert.Len(t, s.GetSlice(), 3)
	
	// Clear should remove all values
	s.Clear()
	assert.False(t, s.Has(1))
	assert.False(t, s.Has(2))
	assert.False(t, s.Has(3))
	assert.Empty(t, s.GetSlice())
}

func TestIntSet_ConcurrentAdd(t *testing.T) {
	s := newIntSet()
	var wg sync.WaitGroup
	
	// Concurrent adds of same value
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add(42)
		}()
	}
	
	wg.Wait()
	
	// Should only have one instance
	assert.True(t, s.Has(42))
	slice := s.GetSlice()
	assert.Len(t, slice, 1)
	assert.Equal(t, 42, slice[0])
}

func TestIntSet_ConcurrentAddDifferentValues(t *testing.T) {
	s := newIntSet()
	var wg sync.WaitGroup
	
	// Concurrent adds of different values
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			s.Add(val)
		}(i)
	}
	
	wg.Wait()
	
	// Should have all 50 values
	slice := s.GetSlice()
	assert.Len(t, slice, 50)
	
	// Verify all values are present
	for i := 0; i < 50; i++ {
		assert.True(t, s.Has(i))
	}
}

func TestIntSet_ConcurrentReadWrite(t *testing.T) {
	s := newIntSet()
	var wg sync.WaitGroup
	
	// Add initial values
	for i := 0; i < 10; i++ {
		s.Add(i)
	}
	
	// Concurrent readers and writers
	for i := 0; i < 20; i++ {
		wg.Add(2)
		
		// Reader
		go func(val int) {
			defer wg.Done()
			s.Has(val % 10)
			s.GetSlice()
		}(i)
		
		// Writer
		go func(val int) {
			defer wg.Done()
			s.Add(val + 100)
		}(i)
	}
	
	wg.Wait()
	
	// Should have at least the original 10 values
	slice := s.GetSlice()
	assert.GreaterOrEqual(t, len(slice), 10)
}

func TestIntSet_ConcurrentClear(t *testing.T) {
	s := newIntSet()
	var wg sync.WaitGroup
	
	// Add some initial values
	for i := 0; i < 10; i++ {
		s.Add(i)
	}
	
	// Concurrent operations with clear
	wg.Add(3)
	
	// Clear in one goroutine
	go func() {
		defer wg.Done()
		s.Clear()
	}()
	
	// Add in another goroutine
	go func() {
		defer wg.Done()
		for i := 20; i < 30; i++ {
			s.Add(i)
		}
	}()
	
	// Read in third goroutine
	go func() {
		defer wg.Done()
		s.GetSlice()
		s.Has(5)
	}()
	
	wg.Wait()
	
	// Should not panic and have some consistent state
	slice := s.GetSlice()
	assert.NotNil(t, slice)
}

func TestIntSet_ZeroValue(t *testing.T) {
	s := newIntSet()
	
	// Should be able to add zero
	added := s.Add(0)
	assert.True(t, added)
	assert.True(t, s.Has(0))
	
	slice := s.GetSlice()
	assert.Contains(t, slice, 0)
}

func TestIntSet_NegativeValues(t *testing.T) {
	s := newIntSet()
	
	// Should handle negative values
	s.Add(-1)
	s.Add(-42)
	
	assert.True(t, s.Has(-1))
	assert.True(t, s.Has(-42))
	assert.False(t, s.Has(-99))
	
	slice := s.GetSlice()
	assert.Contains(t, slice, -1)
	assert.Contains(t, slice, -42)
}
