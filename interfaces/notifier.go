// Package interfaces defines the interfaces for the Notifier pattern.
package interfaces

// INotifiedModel extends IVisited to allow chaining visitors with 
// notifications. This ensures compatibility with the Visitor Pattern 
// while adding a notification layer.
type INotifiedModel interface {
    IVisited[IVisitor]  // Inherits from IVisited (if generic) or uses IVisitor directly
    
    // Accept takes a visitor and optional callbacks (args). Callbacks are 
    // executed after the visit to trigger notifications or side effects.
    Accept(visitor IVisitor, args ...func() IVisited[IVisitor])
}