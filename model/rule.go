// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"fmt"

	"github.com/alecthomas/participle/v2/ebnf"
)

/* EBNFModel */

// EBNFModel represents a complete EBNF model, with notification support.
type EBNFModel struct {
	NotifiedModel  // NotifiedModel inheritance
	*ebnf.EBNF
}

// CreateFrom initializes an EBNFModel from an EBNF node.
func (grm *EBNFModel) CreateFrom(node ebnf.Node) IEBNFModel {
	ebnfNode, ok := node.(*ebnf.EBNF)

	if !ok {
		return nil
	}

	grm.EBNF = ebnfNode
	return grm
}

// DoAccept allows the model to accept a visitor.
func (grm *EBNFModel) DoAccept(visitor IVisitor) {
	
	switch v := visitor.(type) {
	case IParseEBNF:
		v.Visit(grm)
	default:
		grm.VisitedModel.DoAccept(v)
	}

}

// GetEBNF returns the underlying EBNF node.
func (grm *EBNFModel) GetEBNF() *ebnf.EBNF {
	return grm.EBNF
}

// GetText implements interfaces.TextProvider.GetText.
func (grm *EBNFModel) GetText() string {

	if grm.EBNF != nil {
		return grm.EBNF.String()
	}

	return ""
}

// IsValid checks if the model is valid.
func (grm *EBNFModel) IsValid() bool {
	return grm.EBNF != nil
}

/* ExpressionModel */

// ExpressionModel represents an EBNF expression.
type ExpressionModel struct {
	NotifiedModel  // NotifiedModel inheritance
	*ebnf.Expression
	Symbols [][]IRuleModel
}

// CreateFrom initializes an ExpressionModel from an EBNF node.
func (em *ExpressionModel) CreateFrom(node ebnf.Node) IExpressionModel {

	// Redundant check.
	if _, ok := node.(*ebnf.Production); !ok {
        return nil  // Should never happen with the particle/ebnf parser.
    }

	exprNode := node.(*ebnf.Production).Expression
	
	if exprNode == nil {
		return nil
	}
	
	em.Expression = exprNode
	em.Symbols = [][]IRuleModel{}
	return em
}

// DoAccept allows the model to accept a visitor.
func (em *ExpressionModel) DoAccept(visitor IVisitor) {
	
	switch v := visitor.(type) {
	case IParseExpression:
		v.Visit(em)
	default:
		em.VisitedModel.DoAccept(v)
	}

}

func (em *ExpressionModel) GetExpression() *ebnf.Expression {
	return em.Expression
}

// GetSymbols returns the symbols of the expression.
func (em *ExpressionModel) GetSymbols() [][]IRuleModel {
	return em.Symbols
}

// Returns a textual representation of the expression.
func (em *ExpressionModel) GetText() string {

	if em.Expression != nil {
		return em.Expression.String()
	}
	
	return ""
}

func (em *ExpressionModel) IsValid() bool {

    if em == nil || em.Expression == nil {
        return false
    }
    
	// Checks that each symbol group exists and is not empty.
	for _, symbolGroup := range em.Symbols {
    
		if len(symbolGroup) == 0 {
            return false
        }
    
		// Checks that each symbol in the group is valid.
		for _, symbol := range symbolGroup {
    
			if symbol == nil || !symbol.IsValid() {
                return false
            }
    
		}
    
	}
    
	return true
}

/* RuleModel */

// RuleModel represents an EBNF grammar rule.
type RuleModel struct {
	NotifiedModel  				// NotifiedModel inheritance
	*ebnf.Production
	count	   int				// Number of repetitions (e.g., 3 for [letter, letter, letter]).
	Symbol     string
	SymbolType SymbolType 		// Terminal or Non-Terminal.
	rhs        [][]IRuleModel	// Represents the alternatives to the rule: syntagms or codons
}

// CreateFrom initializes a RuleModel from an EBNF node.
func (rm *RuleModel) CreateFrom(node ebnf.Node) IRuleModel {
	prodNode, ok := node.(*ebnf.Production)
	
	if !ok {
		return nil
	}
	
	rm.Production = prodNode
	rm.rhs = [][]IRuleModel{}
	return rm
}

// Clone creates a deep copy of RuleModel, including rhs.
func (rm *RuleModel) Clone() IRuleModel {
    var rhsCopy [][]IRuleModel
    
	if rm.rhs != nil {
        rhsCopy = make([][]IRuleModel, len(rm.rhs))
    
		for i, prod := range rm.rhs {
            rhsCopy[i] = make([]IRuleModel, len(prod))
    
			for j, rule := range prod {
                rhsCopy[i][j] = rule.Clone()
            }

        }

    }
    
    return &RuleModel{
        Symbol:     rm.GetText(),
        SymbolType: rm.GetSymbolType(),
        rhs:        rhsCopy,
    }
}

// DoAccept allows the model to accept a visitor.
func (rm *RuleModel) DoAccept(visitor IVisitor) {
	
	switch v := visitor.(type) {
	case IParseRule:
		v.Visit(rm)
	default:
		rm.VisitedModel.DoAccept(v)
	}

}

func (rm *RuleModel) GetIdentifier() string {
	return rm.Production.Production
}

func (rm *RuleModel) GetSymbols() [][]IRuleModel {
    return rm.rhs
}

// GetSymbolType implements interfaces.IRuleModel.GetSymbolType.
func (rm *RuleModel) GetSymbolType() SymbolType {
    return rm.SymbolType
}

func (rm *RuleModel) GetText() string {
	return rm.Symbol
}

// IsValid checks if the rule is valid.
func (rm *RuleModel) IsValid() bool {

	if rm == nil {
		return false
	}

	return rm.Symbol != "" && (rm.SymbolType == 0 || rm.SymbolType == 1)
}

func (rm *RuleModel) SetSymbols(symbols [][]IRuleModel) {
    rm.rhs = symbols
}

func (rm *RuleModel) String() string {
	return fmt.Sprintf("%s %v", rm.Symbol, rm.SymbolType)
}

/* SequencModel */

// SequenceModel represents a sequence of EBNF expressions, modeling a syntagm.
// In the context of evoGo, a syntagm is treated as a codon: a basic unit for 
// building and evolving grammars. Each codon can be manipulated, evaluated, 
// or recombined during the natural selection operations simulated by the 
// Genomizer. In other words, SequenceModel represents an evolving unit.
// Example: a sequence like "a b c" is a codon that can be optimized or 
// mutated according to the rules of evolutionary grammar.
type SequenceModel struct {
	NotifiedModel         // NotifiedModel inheritance
	*ebnf.Sequence
	Symbols []IRuleModel  // List of rules (symbols) composing the codon.
}

// CreateFrom initializes a SequenceModel from an EBNF node.
func (sm *SequenceModel) CreateFrom(node ebnf.Node) ISequenceModel {
	seqNode, ok := node.(*ebnf.Sequence)
	
	if !ok {
		return nil
	}
	
	sm.Sequence = seqNode
	sm.Symbols = []IRuleModel{}
	return sm
}

// DoAccept allows the model to accept a visitor.
func (sm *SequenceModel) DoAccept(visitor IVisitor) {
	
	switch v := visitor.(type) {
	case IParseSequence:
		v.Visit(sm)
	default:
		sm.VisitedModel.DoAccept(v)
	}

}

func (sm *SequenceModel) GetSequence() *ebnf.Sequence {
	return sm.Sequence
}

// GetSymbols returns the symbols of the sequence.
func (sm *SequenceModel) GetSymbols() []IRuleModel {
	return sm.Symbols
}

// Returns a textual representation of the sequence.
func (sm *SequenceModel) GetText() string {

	if sm.Sequence != nil {
		return sm.Sequence.String()
	}

	return ""
}

func (sm *SequenceModel) IsValid() bool {

	if sm == nil {
		return false
	}

	// Checks if all symbols in the sequence are valid.
	for _, symbol := range sm.Symbols {

		if symbol == nil || !symbol.IsValid() {
			return false
		}
		
	}

	return true
}

/* Exports */

// NewRuleModel creates a new RuleModel instance. Function exported to allow 
// creation from other packages.
func NewRuleModel(symbol string, symbolType SymbolType, rhs [][]IRuleModel) IRuleModel {
    return &RuleModel{
        Symbol:     symbol,
        SymbolType: symbolType,
        rhs:        rhs,  // the private rhs field is exported
    }
}