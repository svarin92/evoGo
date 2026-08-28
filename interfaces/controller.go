// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

import "github.com/alecthomas/participle/v2/ebnf"

/* Genomizer */

// Evolution Controller. It is a Controller specialized in evolution logic.
type IGenomizer interface {
	AddToFailedProductions(production []IRuleModel, fitness float64)
	CorrectByGenome(
		ind IIndividual,
		population []IIndividual,
		fitnessThreshold float64,
		averageFitness float64,
		fitnessFunction FitnessFunc,
	) (bool, error)
	CorrectByGrammaticalPaths(
		ind IIndividual,
		fitnessThreshold float64,
		fitnessFunction FitnessFunc,
	) (bool, error)
	CorrectByTemplate(
		ind IIndividual,
		templateFunction TemplateFunc,
		fitnessFunction FitnessFunc,
	) (bool, error)
	Genomize(genome []int, individual IIndividual) error
	GetAverageFitness(production []IRuleModel) float64
	GetDynamicRules() map[string]IRuleModel
	GetDynamicRuleStack() []string
	GetPhenotype() string
	GetProductionHistory() [][]IRuleModel
	GetSuccessfulProductions() []ISuccessfulProduction
	GetSymbols() map[string]IRuleModel
	GetUsedCodons() int
	GetUsedWraps() int
	FindSimilarProductions(production []IRuleModel, averageFitness float64) [][]IRuleModel
	ProductionSimilarity(p1, p2 []IRuleModel) float64
	RepairIndividual(ind IIndividual) error
	UpdatePatternLibrary(individuals []IIndividual)
	UpdateSuccessfulProductions(individuals []IIndividual)
}

type GenomizerFactory func() IGenomizer

/* Immunizer */

// Repair Controller. It is a Controller specialized in integrity maintenance.
//
// IImmune defines the contract for corrective operations. Eventually,
// this interface will be implemented by a cellular automaton-based system.
type IImmune interface {

	// CorrectByGenome corrects an individual based on their genome and the
	// population.
	CorrectByGenome(
		ind IIndividual,
		population []IIndividual,
		fitnessThreshold float64,
		averageFitness float64,
		fitnessFunction FitnessFunc,
	) (bool, error)

	// --- The "defense" layer: The correctors (CorrectBy*, RepairIndividual)
	// 	   repair defective individuals (such as T lymphocytes that destroy
	//     infected cells). ---

	// CorrectByGrammaticalPaths corrects an individual using grammatical
	// paths.
	CorrectByGrammaticalPaths(
		ind IIndividual,
		fitnessThreshold float64,
		fitnessFunction FitnessFunc,
	) (bool, error)

	// CorrectByTemplate corrects an individual using a custom template.
	CorrectByTemplate(
		ind IIndividual,
		templateFunction TemplateFunc,
		fitnessFunction FitnessFunc,
	) (bool, error)

	RepairIndividual(ind IIndividual) error

	// --- The "memory" layer: The AddToFailedProductions,
	//     UpdateSuccessfulProductions, and UpdatePatternLibrary
	//     methods learn from past errors and successes (such as
	//     B lymphocytes producing antibodies after an infection). ---

	// AddToFailedProductionsm marks a production as "toxic" if it leads to
	// low fitness (such as an antibody neutralizing a pathogen).
	AddToFailedProductions(production []IRuleModel, fitness float64)

	// UpdateSuccessfulProductions enhances cellular transitions that lead
	// to high-performing states (such as the proliferation of effective B
	// lymphocytes).
	UpdateSuccessfulProductions(individuals []IIndividual)

	// UpdatePatternLibrary identifies emerging patterns in the cellular grid
	// (e.g., stable configurations = successful patterns).
	UpdatePatternLibrary(individuals []IIndividual)
}

/* Serializer */

// Data Controller. It is a Controller specialized in data management.
type ISerializer interface {
	GetRules() map[string]IRuleModel
	GetStartRule() string
	GetTerminals() map[string]bool
	Serialize(rules map[string]IRuleModel, node ebnf.Node) error
	ToString() (out string)
}

// SerializerFactory is a function that creates an instance of ISerializer.
type SerializerFactory func() ISerializer
