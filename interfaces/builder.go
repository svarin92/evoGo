// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

import "github.com/alecthomas/participle/v2/ebnf"

// IBuilder defines the interface for building rule models from an EBNF AST.
// Each method allows you to build a specific type of model (e.g., expression, 
// sequence, rule) and apply a Visitor for transformations or validations.
type IBuilder interface {	

	// Getter to retrieve the constructed terms.
	GetTerms() [][]IRuleModel

	// Methods for building models from an AST EBNF node. Each method uses a 
	// Visitor to apply transformations or validations.
	BuildEBNF(node ebnf.Node, visitor IVisitor) error
	BuildExpression(node ebnf.Node, visitor IVisitor) ([][]IRuleModel, error)
	BuildIdentifier(node ebnf.Node, visitor IVisitor) (IRuleModel, error)
	BuildLiteral(node ebnf.Node, visitor IVisitor) (IRuleModel, error)
	BuildRule(node ebnf.Node, visitor IVisitor) (string, error)
	BuildSequence(node ebnf.Node, visitor IVisitor) ([]IRuleModel, error)
	BuildSubExpression(node ebnf.Node, visitor IVisitor) ([][]IRuleModel, error)	
	BuildTerm(node ebnf.Node, visitor IVisitor) (IRuleModel, error)

	// Modify the terms constructed via a transformation function.
	ModifyTerms(func(terms [][]IRuleModel) [][]IRuleModel)
}
