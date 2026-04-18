// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

// Stack is a generic LIFO stack.
// Example of us:
//   stack := NewStack[int]()
//   stack.Push(1, 2, 3)
//   val := stack.Pop() // Returns 3
type Stack[T any] struct {
    elements []T
}

// IsEmpty returns true if the stack is empty.
// Example:
//   if stack.IsEmpty() { ... }
func (s *Stack[T]) IsEmpty() bool {
    return len(s.elements) == 0
}

// Size returns the number of elements in the stack.
// Example:
//   size := stack.Size()
func (s *Stack[T]) Size() int {
    return len(s.elements)
}

// Pop removes and returns the last element from the stack. Returns the value 
// of T to zero if the stack is empty.
// Example:
//   val := stack.Pop()
func (s *Stack[T]) Pop() T {

    if s.IsEmpty() {
        var zero T
        return zero
    }
    
	top := s.elements[len(s.elements)-1]
    s.elements = s.elements[:len(s.elements)-1]
    return top
}

// Push adds one or more elements to the top of the stack.
// Example:
//   stack.Push(1, 2, 3)
func (s *Stack[T]) Push(elements ...T) {
    s.elements = append(s.elements, elements...)
}

// NewStack creates a new empty stack.
// Example:
//   stack := NewStack[string]()
func NewStack[T any]() *Stack[T] {
    return &Stack[T]{elements: []T{}}
}