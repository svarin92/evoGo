// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package evaluator

import (
	"evoGo/interfaces"
	"evoGo/model"
	"evoGo/patterns/visitor"
)

// Alias ​​for functional types.
type (
	FitnessFunc      = interfaces.FitnessFunc
	TemplateFunc     = interfaces.TemplateFunc
	VisitorFunc      = interfaces.VisitorFunc

	EvaluatorFactory = interfaces.EvaluatorFactory
	FitnessFactory   = interfaces.FitnessFactory
	OrganismFactory  = interfaces.OrganismFactory
	TemplateFactory  = interfaces.TemplateFactory
)

// Interfaces imported to ensure architectural consistency.
type (
	INumericMatch    = interfaces.INumericMatch
	IStringMatch     = interfaces.IStringMatch

	IFitnessVisitor  = interfaces.IFitnessVisitor
	ITemplateVisitor = interfaces.ITemplateVisitor
	
	IFitnessMaker    = interfaces.IFitnessMaker
	ITemplateMaker   = interfaces.ITemplateMaker

	INumericTemplate = interfaces.INumericTemplate
	IStringTemplate  = interfaces.IStringTemplate	

	IEvaluable       = interfaces.IEvaluable
	IIndividual      = interfaces.IIndividual
	IOrganism        = interfaces.IOrganism
)

// Alias ​​for specialized types.
type (
	IVisited         = visitor.IVisited
)

// Alias ​​for concrete types.
type (
	ModelVisitor     = visitor.ModelVisitor
	
	Numeric          = model.Numeric
	Ontology         = model.Ontology
)