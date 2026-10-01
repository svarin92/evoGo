// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

import (
	"github.com/svarin92/evoGo/config"
	"github.com/svarin92/evoGo/interfaces"
	"github.com/svarin92/evoGo/model"
)

// Aliases for functional types.
type (
	FitnessFunc        = interfaces.FitnessFunc
	ReplacementFunc    = interfaces.ReplacementFunc
	SelectionFunc      = interfaces.SelectionFunc
	TemplateFunc       = interfaces.TemplateFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IIndividual        = interfaces.IIndividual
	IHybridizationHook = interfaces.IHybridizationHook
	IRuleModel         = interfaces.IRuleModel

	IGenomizer         = interfaces.IGenomizer
	IGrammar           = interfaces.IGrammar
	IImmune            = interfaces.IImmune
)

// Alias ​​for concrete types.
type (
	Individual         = model.Individual
)

const (
	CODONS_SIZE          = config.CODONS_SIZE
	GENERATION_SIZE      = config.GENERATION_SIZE
	MUTATION_PROBABILITY = config.MUTATION_PROBABILITY
)