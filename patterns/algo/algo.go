// The algo package provides a factory for creating specialized parsing 
// algorithms for EBNF (Extended Backus-Naur Form) models. This package 
// extends the Visitor pattern to offer configurable and reusable algorithms 
// for traversing, validating, or transforming EBNF rule structures.
//
// The algorithms created by this package are designed to be used with the 
// models defined in the `model` package (e.g., RuleModel, ExpressionModel) 
// and can be integrated with other patterns such as Builder or Notifier for 
// advanced processing. 
package algo

import (
	"evoGo/interfaces"
	"evoGo/patterns/visitor"
)

// Interfaces imported to ensure architectural consistency.
type (
	IParseEBNF = interfaces.IParseEBNF
	IParseExpression = interfaces.IParseExpression
	IParseRule = interfaces.IParseRule
	IParseSequence = interfaces.IParseSequence
	IParseSubExpression = interfaces.IParseSubExpression
	IParseTerm = interfaces.IParseTerm
//
	IAlgo = interfaces.IAlgo
//
	VisitorFunc = interfaces.VisitorFunc
	ModelVisitor = visitor.ModelVisitor
)

/* ParseEBNF */

// ParseEBNF is a concrete implementation of IParseEBNF. It encapsulates a 
// `visitor.ModelVisitor` and allows you to apply a custom visit function to 
// EBNF models.
type ParseEBNF struct {
	ModelVisitor  // Composition with a generic visitor
}

/* ParseExpression */

// ParseExpression is a concrete implementation of IParseExpression. It allows 
// you to apply custom visit logic to EBNF expressions.
type ParseExpression struct {
	ModelVisitor  // Composition with a generic visitor 
}

/* ParseRule */

// ParseRule is a concrete implementation of IParseRule. It allows you to 
// apply custom logic when an EBNF rule is visited, such as validation, 
// transformation, or information extraction.
type ParseRule struct {
	ModelVisitor  // Composition with a generic visitor
}

/* ParseSequence */

// ParseSequence is a concrete implementation of IParseSequence. It allows 
// you to apply custom visit logic to EBNF sequences.
type ParseSequence struct {
	ModelVisitor  // Composition with a generic visitor
}

/* ParseSubExpression */

// ParseSubExpression is a concrete implementation of IParseSuExpression.
// It allows you to apply custom visit logic to EBNF subexpressions.
type ParseSubExpression struct {
	ModelVisitor  // Composition with a generic visitor
}

/* ParseTerm */

// ParseTerm is a concrete implementation of IParseTerm. It allows you to 
// apply custom visit logic to EBNF terms.
type ParseTerm struct {
	ModelVisitor  // Composition with a generic visitor
}

/* AlgoFactory */

// AlgoMaker is a factory for parsing algorithms for EBNF models. It 
// centralizes the creation of specialized visitors, which facilitates
// their reuse and configuration within the project.
type AlgoMaker struct{}

// Create initializes an AlgoMaker.
func (am *AlgoMaker) Create() *AlgoMaker {
	return am
}

// MakeExpressionCase creates an algorithm for EBNF expressions.
func (am *AlgoMaker) MakeExpressionCase(vf VisitorFunc) IParseExpression {
	return new(ParseExpression).Create(vf)
}

// MakeRuleCase creates an algorithm for EBNF rules.
func (am *AlgoMaker) MakeRuleCase(vf VisitorFunc) IParseRule {
	return new(ParseRule).Create(vf)
}

// MakeRulesCase creates an algorithm for EBNF grammar models.
func (am *AlgoMaker) MakeRulesCase(vf VisitorFunc) IParseEBNF {
	return new(ParseEBNF).Create(vf)
}

// MakeSequenceCase creates an algorithm for EBNF sequences.
func (am *AlgoMaker) MakeSequenceCase(vf VisitorFunc) IParseSequence {
	return new(ParseSequence).Create(vf)
}

// MakeSubExpressionCase creates an algorithm for EBNF subexpressions.
func (am *AlgoMaker) MakeSubExpressionCase(vf VisitorFunc) IParseSubExpression {
	return new(ParseSubExpression).Create(vf)
}

// MakeTermCase creates an algorithm for EBNF terms.
func (am *AlgoMaker) MakeTermCase(vf VisitorFunc) IParseTerm {
	return new(ParseTerm).Create(vf)
}

/* Exports */

// AlgoFactory is a functional type for creating an instance of IAlgo. 
// This allows for flexible initialization and dependency injection.
type AlgoFactory func() IAlgo

// NewAlgo creates a new instance of IAlgo via AlgoMaker.
func NewAlgo() IAlgo {
	return new(AlgoMaker).Create()
}
