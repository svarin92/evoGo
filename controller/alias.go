// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"evoGo/interfaces"
	"evoGo/model"
)

// Interfaces imported to ensure architectural consistency.
type (
	IBuilder             = interfaces.IBuilder
	ISerializer          = interfaces.ISerializer

	IAlgo                = interfaces.IAlgo
	IParseEBNF           = interfaces.IParseEBNF
	IParseExpression     = interfaces.IParseExpression
	IParseRule           = interfaces.IParseRule
	IParseSequence       = interfaces.IParseSequence
	IParseSubExpression  = interfaces.IParseSubExpression
	IParseTerm           = interfaces.IParseTerm
	IRuleModel           = interfaces.IRuleModel
)

// Alias ​​for concrete types.
type (
	EBNFModel            = model.EBNFModel
	ExpressionModel      = model.ExpressionModel
	IdentifierModel      = model.IdentifierModel
	LiteralModel         = model.LiteralModel
	RuleModel            = model.RuleModel
	SubExpressionModel   = model.SubExpressionModel
	SequenceModel        = model.SequenceModel
	SymbolType           = model.SymbolType
	TermModel            = model.TermModel
)

// Constants for symbol types.
const (
	Terminal    SymbolType = model.Terminal
	NonTerminal SymbolType = model.NonTerminal
)