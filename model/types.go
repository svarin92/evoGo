// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Package model provides the data structures to represent EBNF rules and 
// expressions.
package model

import (
	"evoGo/interfaces"
	"evoGo/patterns/notifier"
)

/*
	The self-referential EBNF is:

	EBNF = Production* .
	Production = <ident> "=" Expression+ "." .
	Expression = Sequence ("|" Sequence)* .
	Sequence = Term+ .
	Term = <ident> | Literal | Range | Group | LookaheadGroup | EBNFOption | Repetition | Negation .
	Literal = <string> .
	Range = <string> "…" <string> .
	Group = "(" Expression ")" .
	LookaheadGroup = "(" "?" ("=" | "!") Expression ")" .
	EBNFOption = "[" Expression "]" .
	Repetition = "{" Expression "}" .
	Negation = "!" Expression .

	This EBNF grammar is mapped to the following structures:

	- EBNF → EBNFModel
	- Production → RuleModel
	- Expression → ExpressionModel
	- Sequence → SequenceModel
	- Term → TermModel
	- Litearl → LiteralModel
	- Identifier → IdentifierModel
	- Group → SubExpressionModel
	- etc.
*/

// Interfaces imported to ensure architectural consistency.
type (
	IVisitor = interfaces.IVisitor

	NotifiedModel = notifier.NotifiedModel

	IParseEBNF = interfaces.IParseEBNF
	IParseExpression = interfaces.IParseExpression
	IParseRule = interfaces.IParseRule
	IParseSequence = interfaces.IParseSequence
	IParseSubExpression = interfaces.IParseSubExpression
	IParseTerm = interfaces.IParseTerm

	ITextProvider = interfaces.ITextProvider

	SymbolType = interfaces.SymbolType
	IEBNFModel = interfaces.IEBNFModel
	IExpressionModel = interfaces.IExpressionModel
	IIdentifierModel = interfaces.IIdentifierModel
	ILiteralModel = interfaces.ILiteralModel
	IRuleModel = interfaces.IRuleModel
	ISequenceModel = interfaces.ISequenceModel
	ISubExpressionModel = interfaces.ISubExpressionModel
	ITermModel = interfaces.ITermModel
)

// Constants for symbol types.
const (
	Terminal    SymbolType = interfaces.Terminal
	NonTerminal SymbolType = interfaces.NonTerminal
)