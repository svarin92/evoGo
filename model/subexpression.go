// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"github.com/alecthomas/participle/v2/ebnf"
)

/* SubExpressionModel */

// SubExpressionModel represents a subexpression within an EBNF expression. 
// Inherits from TermModel to reuse the GetLexeme() method and ensure 
// consistency with the ISubExpressionModel interface.
type SubExpressionModel struct {
	TermModel
	*ebnf.SubExpression
	Symbols [][]IRuleModel
}

// CreateFrom initializes a SubExpressionModel from an EBNF node. In the EBNF 
// grammar, a Group (subexpression) is encapsulated within a Term. We therefore
// extract the Group from node.(*ebnf.Term).Group.
func (sem *SubExpressionModel) CreateFrom(node ebnf.Node) ISubExpressionModel {

	// Redundant check.
    if _, ok := node.(*ebnf.Term); !ok {
        return nil  // Should never happen with the particle/ebnf parser.
    }

	subNode := node.(*ebnf.Term).Group

	if subNode == nil {
		return nil
	}

	sem.SubExpression = subNode
	sem.Symbols = [][]IRuleModel{}
	return sem
}

func (sem *SubExpressionModel) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case IParseSubExpression:
		v.Visit(sem)
	default:
		sem.VisitedModel.DoAccept(v)
	}

}

func (sem *SubExpressionModel) GetSubExpression() *ebnf.SubExpression {
	return sem.SubExpression
}

func (sem *SubExpressionModel) GetLexeme() IRuleModel {
	return sem.TermModel.GetLexeme()
}

// GetSymbols returns the symbols of the subexpression.
func (sem *SubExpressionModel) GetSymbols() [][]IRuleModel {
	return sem.Symbols
}

func (sem *SubExpressionModel) IsValid() bool {

	if sem == nil || sem.SubExpression == nil {
		return false
	}
	
	for _, symbolGroup := range sem.Symbols {
	
		if len(symbolGroup) == 0 {
			return false
		}
	
		for _, symbol := range symbolGroup {
	
			if symbol == nil || !symbol.IsValid() {
				return false
			}
	
		}
	
	}
	
	return true
}