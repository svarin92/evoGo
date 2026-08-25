// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"evoGo/config"
	"evoGo/interfaces"
	"evoGo/model"
)

// Aliases for functional types.
type (
	FitnessFunc             = interfaces.FitnessFunc
	TemplateFunc            = interfaces.TemplateFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IBuilder                = interfaces.IBuilder
	IGenomizer              = interfaces.IGenomizer
	IGrammar                = interfaces.IGrammar
	ISerializer             = interfaces.ISerializer
	ISuccessfulProduction   = interfaces.ISuccessfulProduction

	IAlgo                   = interfaces.IAlgo
	IParseEBNF              = interfaces.IParseEBNF
	IParseExpression        = interfaces.IParseExpression
	IParseRule              = interfaces.IParseRule
	IParseSequence          = interfaces.IParseSequence
	IParseSubExpression     = interfaces.IParseSubExpression
	IParseTerm              = interfaces.IParseTerm

	IIndividual             = interfaces.IIndividual
	IRuleModel              = interfaces.IRuleModel
)

// Alias ​​for concrete types.
type (
	EBNFModel               = model.EBNFModel
	ExpressionModel         = model.ExpressionModel
	IdentifierModel         = model.IdentifierModel
	LiteralModel            = model.LiteralModel
	RuleModel               = model.RuleModel
	SubExpressionModel      = model.SubExpressionModel
	SequenceModel           = model.SequenceModel
	TermModel               = model.TermModel
	Individual              = model.Individual   
	DerivationType          = model.DerivationType
	SymbolType              = model.SymbolType   
)

// Constants for symbol types.
const (
	Terminal    SymbolType  = model.Terminal
	NonTerminal SymbolType  = model.NonTerminal
)

const (
	CODONS_SIZE             = config.CODONS_SIZE
	EXPLORATION_PROBABILITY = config.EXPLORATION_PROBABILITY
	MAX_WRAPS               = config.MAX_WRAPS
)

const (
	RightDerivation 	    = model.RightDerivation
	LeftDerivation  	    = model.LeftDerivation
	HomogeneousDerivation   = model.HomogeneousDerivation
)