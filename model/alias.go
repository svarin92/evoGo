// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"evoGo/interfaces"
	"evoGo/patterns/notifier"
)

// Aliases for functional types.
type (
	FitnessFunc         = interfaces.FitnessFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IVisitor            = interfaces.IVisitor

	IParseEBNF          = interfaces.IParseEBNF
	IParseExpression    = interfaces.IParseExpression
	IParseRule          = interfaces.IParseRule
	IParseSequence      = interfaces.IParseSequence
	IParseSubExpression = interfaces.IParseSubExpression
	IParseTerm          = interfaces.IParseTerm
	INumericMatch       = interfaces.INumericMatch
	IStringMatch        = interfaces.IStringMatch
	INumericTemplate    = interfaces.INumericTemplate
	IStringTemplate     = interfaces.IStringTemplate

	ITextProvider       = interfaces.ITextProvider
	
	IEBNFModel          = interfaces.IEBNFModel
	IExpressionModel    = interfaces.IExpressionModel
	IIdentifierModel    = interfaces.IIdentifierModel
	ILiteralModel       = interfaces.ILiteralModel
	IRuleModel          = interfaces.IRuleModel
	ISequenceModel      = interfaces.ISequenceModel
	ISubExpressionModel = interfaces.ISubExpressionModel
	ITermModel          = interfaces.ITermModel
	IIndividual         = interfaces.IIndividual
	IOrganism           = interfaces.IOrganism

	IGenomizer          = interfaces.IGenomizer
	IImmune             = interfaces.IImmune
	
)

// Aliases for concrete types.
type (
	NotifiedModel       = notifier.NotifiedModel
	SymbolType          = interfaces.SymbolType
)