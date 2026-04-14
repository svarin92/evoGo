// Package interfaces define the base interfaces for the patterns used in 
// evoGo. These interfaces are designed to be stable and independent of 
// specific implementations. They allow for complete decoupling between 
// packages (model, algorithm, builder, notifier, visitor, etc.).
package interfaces

type (

	// IVisitable defines a generic contract for types that can be visited by a 
	// visitor of type T. The Accept method allows the visitable object to call 
	// the appropriate visitor method. Args of type U are optional and can be 
	// used to pass additional context.
	IVisitable[T, U any] interface {
		Accept(v T, args ...U)
	}

	// IVisited extends IVisitable to provide a concrete method (DoAccept) for 
	// accepting a visitor of type T. It enforces that the Accept method returns 
	// a function that produces an IVisited[T]. This is useful for chaining 
	// visitor operations or lazy evaluation.
	IVisited[T IVisitor] interface {
		IVisitable[T, func() IVisited[T]]
		DoAccept(v T)
	}

	// IVisitor defines the contract for visitor objects. The Visit method is 
	// called by visitable objects to perform operations on them. Using `any` 
	// as the parameter type allows this visitor to handle any visitable 
	// object.
	IVisitor interface {
		Visit(any)
	}

	// VisitorFunc allows to define visitors as anonymous functions, which 
	// simplifies the creation of visitors for one-off operations. It takes 
	// a single parameter of type `any` to allow flexibility in the types of 
	// data it can process.
	VisitorFunc func(data any)

)