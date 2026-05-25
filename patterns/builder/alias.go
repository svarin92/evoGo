// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package builder

import (
	"evoGo/interfaces"
	"evoGo/model"
)

// Aliases for functional types.
type (
	BuilderFactory     = interfaces.BuilderFactory
)

// Interfaces imported to ensure architectural consistency.
type (
	IBuilder           = interfaces.IBuilder
	IVisitor           = interfaces.IVisitor
	IVisited           = interfaces.IVisited[IVisitor]

	IRuleModel         = interfaces.IRuleModel
)

// Aliases for concrete types.
type (
	EBNFModel          = model.EBNFModel
	ExpressionModel    = model.ExpressionModel
	IdentifierModel    = model.IdentifierModel
	LiteralModel       = model.LiteralModel
	RuleModel          = model.RuleModel
	SequenceModel      = model.SequenceModel
	SubExpressionModel = model.SubExpressionModel
	TermModel          = model.TermModel
)