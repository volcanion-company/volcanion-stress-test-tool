package engine

import "sync"

// RingBuffer is a fixed-size circular buffer for storing latencies
type RingBuffer struct {
	data  []float64
	size  int
	head  int
	count int
	mu    sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with the specified size
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data:  make([]float64, size),
		size:  size,
		head:  0,
		count: 0,
	}
}

// Add adds a value to the ring buffer
func (rb *RingBuffer) Add(value float64) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = value
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// GetAll returns all values currently in the buffer
func (rb *RingBuffer) GetAll() []float64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]float64, rb.count)
	if rb.count == 0 {
		return result
	}

	if rb.count < rb.size {
		// Buffer not full yet, return from start to count
		copy(result, rb.data[:rb.count])
	} else {
		// Buffer is full, need to reorder
		copy(result, rb.data[rb.head:])
		copy(result[rb.size-rb.head:], rb.data[:rb.head])
	}

	return result
}

// Count returns the number of elements in the buffer
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// IsFull returns true if the buffer is at capacity
func (rb *RingBuffer) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count == rb.size
}
