// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package algo

import (
	"evoGo/interfaces"
	"evoGo/patterns/visitor"
)

// Aliases for functional types.
type (
	VisitorFunc = interfaces.VisitorFunc

	AlgoFactory = interfaces.AlgoFactory
)

// Interfaces imported to ensure architectural consistency.
type (
	IParseEBNF = interfaces.IParseEBNF
	IParseExpression = interfaces.IParseExpression
	IParseRule = interfaces.IParseRule
	IParseSequence = interfaces.IParseSequence
	IParseSubExpression = interfaces.IParseSubExpression
	IParseTerm = interfaces.IParseTerm

	IAlgo = interfaces.IAlgo
)

// Aliases for concrete types.
type (
	ModelVisitor = visitor.ModelVisitor
)