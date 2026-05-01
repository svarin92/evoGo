// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Main package - Immune system for the Genomizer project. This file 
// defines the IImmuneSystem interface and its temporary implementation 
// by the Genomizer, pending the transition to cellular automata.
package main

// IImmuneSystem defines the contract for corrective operations. Eventually, 
// this interface will be implemented by a cellular automaton-based system.
type IImmuneSystem interface {

    // CorrectByGenome corrects an individual based on their genome and the 
	// population.
    CorrectByGenome(
		ind *Individual, 
		population []*Individual, 
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
		ind *Individual, 
		fitnessThreshold float64,
		fitnessFunction FitnessFunc,
	) (bool, error)

    // CorrectByTemplate corrects an individual using a custom template.
    CorrectByTemplate(
		ind *Individual, 
		templateFunction TemplateFunc,
		fitnessFunction FitnessFunc,
	) (bool, error)

	RepairIndividual(ind *Individual) error

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
	UpdateSuccessfulProductions(individuals []*Individual)

	// UpdatePatternLibrary identifies emerging patterns in the cellular grid 
	// (e.g., stable configurations = successful patterns).
    UpdatePatternLibrary(individuals []*Individual)
}

// --- Temporary implementation by Genomizer (to be removed later) ---

// GenomizerImmuneAdapter allows the Genomizer to implement 
// IImmuneSystem. This implementation is temporary and will be 
// replaced by CellularAutomatonImmuneSystem. It simply delegates 
// calls to existing Genomizer methods.
type GenomizerImmuneAdapter struct {
    genomizer IGenomizer
}

// AddToFailedProductions delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) AddToFailedProductions(production []IRuleModel, fitness float64) {
    adapter.genomizer.AddToFailedProductions(production, fitness)
}

// CorrectByGenome delegates to the Genomizer (temporary implementation).
func (adapter *GenomizerImmuneAdapter) CorrectByGenome(
    ind *Individual,
    population []*Individual,
    fitnessThreshold float64,
    averageFitness float64,
	fitnessFunction FitnessFunc,
) (bool, error) {

    // Call to the existing Genomizer method.
    return adapter.genomizer.CorrectByGenome(ind, population, fitnessThreshold, averageFitness, fitnessFunction)
}

// CorrectByGrammaticalPaths delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) CorrectByGrammaticalPaths(
    ind *Individual,
    fitnessThreshold float64,
	fitnessFunction FitnessFunc,
) (bool, error) {
    return adapter.genomizer.CorrectByGrammaticalPaths(ind, fitnessThreshold, fitnessFunction)
}

// CorrectByTemplate delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) CorrectByTemplate(
    ind *Individual,
    templateFunction TemplateFunc,
	fitnessFunction FitnessFunc,
) (bool, error) {
    return adapter.genomizer.CorrectByTemplate(ind, templateFunction, fitnessFunction)
}

// RepairIndividual delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) RepairIndividual(ind *Individual) error {
    return adapter.genomizer.RepairIndividual(ind)
}

// UpdateSuccessfulProductions delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) UpdateSuccessfulProductions(individuals []*Individual) {
    adapter.genomizer.UpdateSuccessfulProductions(individuals)
}

// UpdatePatternLibrary delegates to the Genomizer..
func (adapter *GenomizerImmuneAdapter) UpdatePatternLibrary(individuals []*Individual) {
    adapter.genomizer.UpdatePatternLibrary(individuals)
}

/* Exports */

// NewGenomizerImmuneAdapter creates an adapter for the Genomizer.
func NewGenomizerImmuneAdapter(genomizer IGenomizer) *GenomizerImmuneAdapter {

	if genomizer == nil {
		panic("genomizer cannot be nil")
	}

    return &GenomizerImmuneAdapter{genomizer: genomizer}
}