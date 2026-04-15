// Package pattern implements the Notifier pattern, which extends the Visitor
// pattern to support notification mechanisms after a visit operation.
//
// The Notifier pattern is useful when you need to:
// - Decouple the notification logic from the visitor logic.
// - Trigger side effects (e.g., logging, updates) after a visit.
// - Chain multiple visitors with notification callbacks.
//
// Key components:
// - INotifiedModel: Extends IVisited to support notification after a visit.
// - NotifiedModel: Concrete implementation of INotifiedModel.
package notifier

import "evoGo/interfaces"
import "evoGo/patterns/visitor"

// Interfaces imported to ensure architectural consistency.
type (
	IVisitor = interfaces.IVisitor
	IVisited = interfaces.IVisited[IVisitor]
	VisitedModel = visitor.VisitedModel
)

/* NotifiedModel */

type (

	// NotifiedModel includes an implementation of IVisitedModel. This struct 
	// is used as the basis for all notifiable models.
	NotifiedModel struct {
		VisitedModel
	}
	
)

// Accept implements the IVisitable interface for NotifiedModel. If args are 
// provided, the first argument is treated as a notification callback. The 
// callback is executed after the visit, allowing for post-visit actions 
// (e.g., notifications).
func (nm *NotifiedModel) Accept(visitor IVisitor, args ...func() IVisited) {
	
	if len(args) > 0 {
		notify := args[0]()
		notify.DoAccept(visitor)
	} else {
		nm.DoAccept(visitor)
	}
	
}

// DoAccept delegates the visit to the underlying model (VisitedModel).
// This method can be overridden to add specific behaviors before/after 
// the visit.
func (nm *NotifiedModel) DoAccept(visitor IVisitor) {
    nm.VisitedModel.DoAccept(visitor)
}

/* Exports */

func NewNotifiedModel(vm VisitedModel) *NotifiedModel {
    return &NotifiedModel{vm}
}