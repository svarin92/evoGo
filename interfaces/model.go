// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Package interfaces defines the interfaces for EBNF models. 
package interfaces

import "github.com/alecthomas/participle/v2/ebnf"

// SymbolType defines the type of a symbol (Terminal or NonTerminal).
type SymbolType int

const (
    Terminal    SymbolType = 0
    NonTerminal SymbolType = 1
)

// ITextProvider is an interface for objects that provide text.
type ITextProvider interface {
    GetText() string
}

// IEBNFModel defines the methods for complete EBNF models.
type IEBNFModel interface {
	ITextProvider
	INotifiedModel  
	GetEBNF() *ebnf.EBNF
	IsValid() bool
}

// IExpressionModel defines the methods for EBNF expressions.
type IExpressionModel interface {
	ITextProvider
	INotifiedModel
	GetExpression() *ebnf.Expression
	
	// GetSymbols returns the alternatives for the expression.
	GetSymbols() [][]IRuleModel

	IsValid() bool
}

// IIdentifierModel extends TermModel for identifiers.
type IIdentifierModel interface {
	ITermModel
	GetIdentifier() string
}

// ILiteralModel extends TermModel for literals.
type ILiteralModel interface {
	ITermModel
}

// IRuleModel defines the methods common to all rule models. 
type IRuleModel interface {
    ITextProvider
	INotifiedModel
	Clone() IRuleModel
	GetIdentifier() string
	GetSymbols() [][]IRuleModel  // Returns the symbols (abstraction of rhs)
	GetSymbolType() SymbolType 
    IsValid() bool
	SetSymbols(symbols [][]IRuleModel)
}

// ISequenceModel defines the methods for EBNF sequences. In evoGo, a sequence
// is treated as a codon: an evolving unit for grammars.
type ISequenceModel interface {
	ITextProvider
	INotifiedModel
	GetSequence() *ebnf.Sequence

	// GetSymbols returns the terms of the sequence.
	GetSymbols() []IRuleModel

	IsValid() bool
}

// ISubExpressionModel defines the methods for EBNF subexpressions. Inherits 
// from ITermModel to reuse the GetLexeme() method and ensure consistency with
// the SubExpressionModel structure.
type ISubExpressionModel interface {
	ITermModel
	GetSubExpression() *ebnf.SubExpression
	GetSymbols() [][]IRuleModel
}

// ITermModel defines the methods common to all EBNF terms.
type ITermModel interface {
	ITextProvider
	INotifiedModel
	GetLexeme() IRuleModel
	IsValid() bool
}