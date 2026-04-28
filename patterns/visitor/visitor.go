// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Package pattern implements the Visitor design pattern in Go using generics.
//
// The Visitor pattern allows for adding new operations to existing object 
// structures without modifying those structures. This implementation uses Go 
// generics to provide type safety and flexibility.
//
// Key components:
// - IVisitable: Defines the contract for objects that can be visited.
// - IVisitor: Defines the contract for visitor objects.
// - IVisited: Extends IVisitable to support chaining and lazy evaluation.
// - VisitedModel: Concrete implementation of a visitable object.
// - ModelVisitor: A concrete visitor implementation using function 
//   composition.
package visitor

import (
	"fmt"

	"evoGo/interfaces"
)

// Interfaces imported to ensure architectural consistency.
type (
	IVisited = interfaces.IVisited[IVisitor]
	IVisitor = interfaces.IVisitor
	VisitorFunc = interfaces.VisitorFunc
)

/* VisitedModel */

// VisitedModel represents a concrete implementation of the IVisited interface.
// It serves as a minimal, extensible model that can be visited by an IVisitor.
// The empty struct is used here because the Visitor pattern typically operates
// on the behavior (methods) rather than the state (fields) of the visited 
// object. You can extend this struct with fields if your use case requires 
// stateful objects.
type VisitedModel struct {
	// Empty.	
}

// Create initializes and returns a VisitedModel instance. Must be 
// called on a pointer receiver (e.g., new(VisitedModel).Create() or 
// (&VisitedModel{}).Create()).
func (vm *VisitedModel) Create() *VisitedModel {
	return vm
}

// Accept implements the IVisitable interface. If no additional arguments 
// (args) are provided, it delegates the visit to DoAccept. Args can be used 
// to pass functions that return IVisited[IVisitor], enabling lazy evaluation
// or chaining.
func (vm *VisitedModel) Accept(
	visitor IVisitor, 
	args ...func() IVisited,
) {	

	if args == nil {
		vm.DoAccept(visitor)
	}

}

// DoAccept performs the actual visit operation using the provided visitor.
// It checks if the visitor is nil to avoid runtime errors. If the visitor 
// is nil, it logs an error message.
func (vm *VisitedModel) DoAccept(visitor IVisitor) {

	// if visitor != nil { visitor.Visit(vm) }
	algo := visitor

	if algo == nil {
		fmt.Println(`Error: Visitor cannot be nil: provide a concrete implementation of IVisitor`)
		// This check is necessary because we don't want to call the
		// Visit method on a nil visitor.
	}
	
	algo.Visit(vm)
}

/* ModelVisitor */

// ModelVisitor is a concrete implementation of IVisitor. It wraps a 
// VisitorFunc to provide a Visit method that executes the function on 
// the provided data.
type ModelVisitor struct {
	Algo VisitorFunc
}

// Create initializes the ModelVisitor with a VisitorFunc. It returns the 
// ModelVisitor instance for method chaining. Must be called on a pointer 
// receiver (e.g., new(ModelVisitor).Create() or (&ModelVisitor{}).Create()).
func (mv *ModelVisitor) Create(vf VisitorFunc) *ModelVisitor {
	mv.Algo = vf
	return mv
}

// Visit implements the IVisitor interface. It executes the wrapped 
// VisitorFunc on the provided data.
func (mv *ModelVisitor) Visit(data any) {
	mv.Algo(data)
}
