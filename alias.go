// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package main

import (
	"evoGo/config"
	"evoGo/controller"
	"evoGo/interfaces"
	"evoGo/model"
)

// Aliases for functional types.
type (
	FitnessFunc           = interfaces.FitnessFunc
	TemplateFunc          = interfaces.TemplateFunc
	ReplacementFunc       = interfaces.ReplacementFunc
	SelectionFunc         = interfaces.SelectionFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IGenomizer            = interfaces.IGenomizer
	IGrammar              = interfaces.IGrammar
	IImmune               = interfaces.IImmune
	ISuccessfulProduction = interfaces.ISuccessfulProduction

	IIndividual           = interfaces.IIndividual
	IRuleModel            = interfaces.IRuleModel
)

// Alias ​​for concrete types.
type (
	SymbolType            = model.SymbolType
	Individual            = model.Individual

	SuccessfulProduction  = controller.SuccessfulProduction
)

// Constants for symbol types.
const (
	Terminal    SymbolType = model.Terminal
	NonTerminal SymbolType = model.NonTerminal
)

const (
	CODONS_SIZE             = config.CODONS_SIZE
	EXPLORATION_PROBABILITY = config.EXPLORATION_PROBABILITY
	GENERATIONS             = config.GENERATIONS
	GENERATION_SIZE         = config.GENERATION_SIZE
	MAX_WRAPS               = config.MAX_WRAPS
	MUTATION_PROBABILITY    = config.MUTATION_PROBABILITY
	POPULATION_SIZE         = config.POPULATION_SIZE
)