// Package interfaces defines the specialized interfaces for EBNF parsing 
// algorithms. These interfaces extend IVisitor to allow specialization by 
// model type.
package interfaces

// Interfaces for specialized parsing algorithms. Each interface extends 
// `visitor.IVisitor` to ensure compatibility with the Visitor pattern, 
// while allowing specialization for each type of pattern (EBNF, Expression, 
// Rule, etc.).
type (

	// IParseEBNF defines a specialized visitor for full EBNF models. This 
	// type of visitor is used to traverse and process higher-level nodes of 
	// an EBNF grammar, such as productions or global rules.
	IParseEBNF          interface { IVisitor }
	
	// IParseExpression defines a specialized visitor for EBNF expressions.
	// Expressions represent sequences or alternatives in a grammar. This 
	// visitor allows you to browse and process these structures in a 
	// specific  way.
	IParseExpression    interface { IVisitor }
	
	// IParseRule defines a specialized visitor for EBNF rules. A rule 
	// represents a production in a grammar, for example: `rule = "a" | "b"`.
	// This visitor allows you to browse and process rules individually.	
	IParseRule          interface { IVisitor }
	
	// IParseSequence defines a specialized visitor for EBNF sequences. A 
	// sequence represents a series of terms in an expression, for example:
	// `sequence = term1 term2 term3`. This visitor allows you to browse and 
	// process sequences in a specific way.
	IParseSequence      interface { IVisitor }
	
	// IParseSubExpression defines a specialized visitor for EBNF 
	// subexpressions. A subexpression represents a nested expression, 
	// for example: `subexpr = (expression)`. This visitor allows you to 
	// iterate over and process these substructures.
	IParseSubExpression interface { IVisitor }
	
	// IParseTerm defines a specialized visitor for EBNF terms. A term 
	// represents a basic unit in an expression, such as an identifier
	// or a literal. For example: `term = "a"` or `term = identifier`.
	// This visitor allows you to browse and process terms individually.
	IParseTerm          interface { IVisitor }
)

// IAlgo defines the interface for a parsing algorithm factory. This 
// interface allows you to create specialized visitors for each type of 
// EBNF model (rules, expressions, sequences, etc.), which facilitates 
// the extension and reuse of algorithms.
type IAlgo interface {

	// MakeExpressionCase creates a parsing algorithm for EBNF expressions.
	MakeExpressionCase(vf VisitorFunc) IParseExpression

	// MakeRuleCase creates a parsing algorithm for EBNF rules.
	MakeRuleCase(vf VisitorFunc) IParseRule
	
	// MakeRulesCase creates a parsing algorithm for complete EBNF models.
	MakeRulesCase(vf VisitorFunc) IParseEBNF
	
	// MakeSequenceCase creates a parsing algorithm for EBNF sequences.
	MakeSequenceCase(vf VisitorFunc) IParseSequence
	
	// MakeSubExpressionCase creates a parsing algorithm for EBNF 
	// subexpressions.
	MakeSubExpressionCase(vf VisitorFunc) IParseSubExpression
	
	// MakeTermCase creates a parsing algorithm for EBNF terms.
	MakeTermCase(vf VisitorFunc) IParseTerm
}
