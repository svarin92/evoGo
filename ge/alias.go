// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

import (
	"evoGo/config"
	"evoGo/interfaces"
)

// Aliases for functional types.
type (
	FitnessFunc     = interfaces.FitnessFunc
	ReplacementFunc = interfaces.ReplacementFunc
	SelectionFunc   = interfaces.SelectionFunc
	TemplateFunc    = interfaces.TemplateFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IIndividual    = interfaces.IIndividual
	IRuleModel     = interfaces.IRuleModel

	IGenomizer     = interfaces.IGenomizer
	IGrammar       = interfaces.IGrammar
	IImmune        = interfaces.IImmune
)

const (
	CODONS_SIZE          = config.CODONS_SIZE
	GENERATION_SIZE      = config.GENERATION_SIZE
	MUTATION_PROBABILITY = config.MUTATION_PROBABILITY
)