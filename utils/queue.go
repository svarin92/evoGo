// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

// A generic FIFO queue.
// Example usage:
//   queue := NewQueue[int]()
//   queue.Enqueue(1, 2, 3)
//   val := queue.Dequeue() // Returns 1
type Queue[T any] struct {
    elements []T
}

// Enqueue adds one or more elements to the end of the queue.
// Example:
//   queue.Enqueue(1, 2, 3)
func (q *Queue[T]) Enqueue(elements ...T) {
	q.elements = append(q.elements, elements...)
}

// Dequeue removes and returns the first element from the queue.
// Returns the value T as zero if the queue is empty.
// Example:
//   val := queue.Dequeue()
func (q *Queue[T]) Dequeue() T {

	if q.IsEmpty() {
		var zero T
		return zero
	}

	first := q.elements[0]
	q.elements = q.elements[1:]
	return first
}

// IsEmpty returns true if the queue is empty.
// Example :
//   if queue.IsEmpty() { ... }
func (q *Queue[T]) IsEmpty() bool {
	return len(q.elements) == 0
}

// Size returns the number of elements in the queue.
// Example :
//   size := queue.Size()
func (q *Queue[T]) Size() int {
	return len(q.elements)
}

// NewQueue creates a new empty queue.
// Example:
//   queue := NewQueue[string]()
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{elements: make([]T, 0)}
}