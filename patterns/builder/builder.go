// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Package patterns implements the Builder design pattern for constructing rule
// models from an EBNF AST. This Builder is designed for use with the Visitor
// and Notifier patterns, to enable transformations, validations, or notifications
// during rule construction.
// It is a key component of the evoGo project, particularly for managing evolving
// grammars.
package builder

import (
	"fmt"

	"github.com/alecthomas/participle/v2/ebnf"

	"evoGo/interfaces"
	"evoGo/model"
)

// Interfaces imported to ensure architectural consistency.
type (
	IBuilder = interfaces.IBuilder
	IVisitor = interfaces.IVisitor
	IVisited = interfaces.IVisited[IVisitor]

	IRuleModel = interfaces.IRuleModel
	EBNFModel = model.EBNFModel
	ExpressionModel = model.ExpressionModel
	IdentifierModel = model.IdentifierModel
	LiteralModel = model.LiteralModel
	RuleModel = model.RuleModel
	SequenceModel = model.SequenceModel
	SubExpressionModel = model.SubExpressionModel
	TermModel = model.TermModel
)

/* BuilderFactory */

// A struct representing the rule builder. RuleBuilder is a concrete 
// implementation of IBuilder.
type RuleBuilder struct{
	terms [][]*RuleModel // Stores term groups as a slice of slices
}

// Create initializes a new RuleBuilder with empty terms.
func (rb *RuleBuilder) Create() *RuleBuilder {

	// terms: Stores rule groups constructed as [][]*RuleModel. Each 
	// subslice represents a rule group (e.g., an alternative in an 
	// expression).
	rb.terms = [][]*RuleModel{}
	
	return rb
}

// AddTermGroup adds a term group to the builder.
func (rb *RuleBuilder) AddTermGroup(terms []*model.RuleModel) {
    rb.terms = append(rb.terms, terms)
}

// BuildEBNF builds an EBNF model from an AST node. The Visitor is used to 
// apply transformations or validations to the model.
func (rb *RuleBuilder) BuildEBNF(
	node ebnf.Node, 
	visitor IVisitor,
) error {
	ebnf := new(EBNFModel).CreateFrom(node)

	if ebnf == nil {
		return fmt.Errorf("failed to build grammar from node: %v", node)
	}

	ebnf.Accept(visitor, func() IVisited {
		return ebnf
	})
	return nil
}

// BuildExpression constructs an expression model from an AST node. The 
// Visitoris used to apply transformations or validations to the expression. 
// Returns a list of rule groups (each group represents an alternative in 
// the expression). .
func (rb *RuleBuilder) BuildExpression(
	node ebnf.Node, 
	visitor IVisitor,
) ([][]IRuleModel, error) {
	expr := new(ExpressionModel).CreateFrom(node)

	if expr == nil {
		return nil, fmt.Errorf("failed to build expression from node: %v", node)
	}

	expr.Accept(visitor, func() IVisited {
		return expr
	})
	return expr.GetSymbols(), nil
}

// BuildIdentifier constructs an identifier model from an AST node. The Visitor
// is used to apply transformations or validations to the identifier. Returns 
// the identifier's lexeme.
func (rb *RuleBuilder) BuildIdentifier(
	node ebnf.Node, 
	visitor IVisitor,
) (IRuleModel, error) {
	ident := new(IdentifierModel).CreateFrom(node)

	if ident == nil {
		return nil, fmt.Errorf("failed to build identifier from node: %v", node)
	}

	ident.Accept(visitor, func() IVisited {
		return ident
	})
	return ident.GetLexeme(), nil
}

// BuildLiteral constructs a literal model from an AST node. The Visitor is 
// used to apply transformations or validations to the literal. Returns the 
// literal's lexeme.
func (rb *RuleBuilder) BuildLiteral(
	node ebnf.Node, 
	visitor IVisitor,
) (IRuleModel, error) {
	text := new(LiteralModel).CreateFrom(node)

	if text == nil {
		return nil, fmt.Errorf("failed to build literal from node: %v", node)
	}

	text.Accept(visitor, func() IVisited {
		return text
	})
	return text.GetLexeme(), nil
}

// BuildRule constructs a rule template from an AST node. The Visitor is 
// used to apply transformations or validations to the rule. Returns the 
// rule symbol.
func (rb *RuleBuilder) BuildRule(
	node ebnf.Node, 
	visitor IVisitor,
) (string, error) {
	rule := new(RuleModel).CreateFrom(node)

	if rule == nil {
		return "", fmt.Errorf("failed to build rule from node: %v", node)
	}

	rule.Accept(visitor, func() IVisited {
		return rule
	})
	return rule.GetText(), nil
}

// BuildSequence constructs a sequence model from an AST node. The Visitor is
// used to apply transformations or validations to the sequence. Returns the 
// symbols of the sequence.
func (rb *RuleBuilder) BuildSequence(
	node ebnf.Node, 
	visitor IVisitor,
) ([]IRuleModel, error) {
	seq := new(SequenceModel).CreateFrom(node)

	if seq == nil {
		return nil, fmt.Errorf("failed to build sequence from node: %v", node)
	}

	seq.Accept(visitor, func() IVisited {
		return seq
	})
	return seq.GetSymbols(), nil
}

// BuildSubExpression constructs a subexpression model from an AST node. 
// The Visitor is used to apply transformations or validations to the 
// subexpression. Returns the symbols of the subexpression as groups of 
// terms.
func (rb *RuleBuilder) BuildSubExpression(
	node ebnf.Node, 
	visitor IVisitor,
) ([][]IRuleModel, error) {
	group := new(SubExpressionModel).CreateFrom(node)

	if group == nil {
		return nil, fmt.Errorf("failed to build group from node: %v", node)
	}

	group.Accept(visitor, func() IVisited {
		return group
	})
	return group.GetSymbols(), nil
}

// BuildTerm constructs a term model from an AST node. The Visitor is used to 
// apply transformations or validations to the term. Returns the term's lexeme.
func (rb *RuleBuilder) BuildTerm(
	node ebnf.Node, 
	visitor IVisitor,
) (IRuleModel, error) {
	term := new(TermModel).CreateFrom(node)

	if term == nil {
		return nil, fmt.Errorf("failed to build term from node: %v", node)
	}

	term.Accept(visitor, func() IVisited {
		return term
	})
	return term.GetLexeme(), nil
}

// ModifyConcreteTerms allows you to modify internal term groups by applying
// a transformation function. Invalid (empty) term groups are filtered out.
func (rb *RuleBuilder) GetConcreteTerms() [][]*RuleModel {
    return rb.terms  // Returns the field directly
}

// GetTerms returns the term groups (for a public API).
func (rb *RuleBuilder) GetTerms() [][]IRuleModel {

	// Go does not allow implicit conversion between [][]*RuleModel and 
	// [][]IRuleModel, even if *RuleModel implements IRuleModel.
	result := make([][]IRuleModel, len(rb.terms))
    
	for i, group := range rb.terms {
        result[i] = make([]IRuleModel, len(group))
    
		for j, rule := range group {

			// Safe conversion because *RuleModel implements IRuleModel.
            result[i][j] = rule  // rule is of type *RuleModel, which implements IRuleModel
        }
    
	}
    
	return result
}

// IsValidTermGroup checks if a group of terms is valid (not empty and valid 
// rules).
func (rb *RuleBuilder) IsValidTermGroup(termGroup []*RuleModel) bool {
    
	// return len(termGroup) > 0
	if len(termGroup) == 0 {
        return false
    }

    for _, rule := range termGroup {
    
		if rule == nil || !rule.IsValid() {  // Vérifie aussi la validité de chaque règle
            return false
        }
    
	}
    
	return true
}

// ModifyTerms allows you to modify term groups by applying a transformation 
// function. This method is designed for public use (e.g., by other packages 
// like `algo` or `visitor`). The modification function can add, delete, or 
// modify term groups. 
// Warning: This method does NOT filter invalid term groups. It is the caller's 
// responsibility to ensure that the modified terms are valid: No term group is
// empty and all rules within the groups are valid (see IRuleModel.IsValid()). 
// For a version with automatic filtering, use ModifyConcreteTerms (internal 
// use).
func (rb *RuleBuilder) ModifyTerms(modifier func([][]IRuleModel) [][]IRuleModel) {

	// Retrieve the current terms as interfaces (for the public API).
    terms := rb.GetTerms()

	// Apply the modification function.
    modifiedTerms := modifier(terms)

	// Convert the modified terms (interfaces) into concrete terms 
	// (*RuleModel).
    rb.terms = make([][]*RuleModel, len(modifiedTerms))
    
	for i, group := range modifiedTerms {
        rb.terms[i] = make([]*RuleModel, len(group))
    
		for j, rule := range group {

			// Safe conversion because the rules come from rb.GetTerms() 
			// (which returns *RuleModel).
            rb.terms[i][j] = rule.(*RuleModel)
        }
    
	}

}

// ModifyConcreteTerms allows modifying the internal terms by providing a 
// closure. It takes a modifier function that can transform the current terms 
// into new ones, potentially adding, removing, or altering term groups as 
// needed.
func (rb *RuleBuilder) ModifyConcreteTerms(modifier func(terms [][]*RuleModel) [][]*RuleModel) {
   
    // Apply the modification function to the current terms.
    modifiedTerms := modifier(rb.terms)

	// Filter out invalid term groups.
	var validTerms [][]*RuleModel

    for _, termGroup := range modifiedTerms {
    
		// Filter out any invalid term groups from the modified terms.
		if rb.IsValidTermGroup(termGroup) {
            validTerms = append(validTerms, termGroup)
        }
    
	}	

    // Update the builder's terms with the filtered and modified terms.
    rb.terms = validTerms
}

/* Exports */

// BuilderFactory is a function that returns a new instance of IBuilder.
type BuilderFactory func() IBuilder

// NewBuilder creates and returns a new instance of RuleBuilder.
func NewBuilder() IBuilder {
	return new(RuleBuilder).Create()
}
