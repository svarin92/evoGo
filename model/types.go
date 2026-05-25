// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Package model provides the data structures to represent EBNF rules and 
// expressions.
package model

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

import "evoGo/interfaces"

// Constants for symbol types.
const (
	Terminal    SymbolType = interfaces.Terminal
	NonTerminal SymbolType = interfaces.NonTerminal
)