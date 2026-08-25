// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Main package - Immune system for the Genomizer project. This file defines 
// the IImmune interface and its temporary implementation by the Genomizer, 
// pending the transition to cellular automata.
package controller

import (
	"fmt"
)

// --- Temporary implementation by Genomizer (to be removed later) ---

// GenomizerImmuneAdapter allows the Genomizer to implement IImmune. For now, 
// it is simply delegating calls to existing Genomizer methods. Until this be 
// replaced by CellularAutomatonImmuneSystem.
type GenomizerImmuneAdapter struct {
    genomizer IGenomizer
}

// AddToFailedProductions delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) AddToFailedProductions(production []IRuleModel, fitness float64) {
    adapter.genomizer.AddToFailedProductions(production, fitness)
}

// CorrectByGenome delegates to the Genomizer (temporary implementation).
func (adapter *GenomizerImmuneAdapter) CorrectByGenome(
    ind IIndividual,
    population []IIndividual,
    fitnessThreshold float64,
    averageFitness float64,
	fitnessFunction FitnessFunc,
) (bool, error) {

    // Call to the existing Genomizer method.
    return adapter.genomizer.CorrectByGenome(ind, population, fitnessThreshold, averageFitness, fitnessFunction)
}

// CorrectByGrammaticalPaths delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) CorrectByGrammaticalPaths(
    ind IIndividual,
    fitnessThreshold float64,
	fitnessFunction FitnessFunc,
) (bool, error) {

    // Call to the existing Genomizer method.
	return adapter.genomizer.CorrectByGrammaticalPaths(ind, fitnessThreshold, fitnessFunction)
}

// CorrectByTemplate delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) CorrectByTemplate(
    ind IIndividual,
    templateFunction TemplateFunc,
	fitnessFunction FitnessFunc,
) (bool, error) {

    // Call to the existing Genomizer method.
    return adapter.genomizer.CorrectByTemplate(ind, templateFunction, fitnessFunction)
}

// RepairIndividual delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) RepairIndividual(ind IIndividual) error {
	concreteInd, ok := ind.(*Individual)
    
	if !ok {
        return fmt.Errorf("individual is not *Individual")
    }

    // Call to the existing Genomizer method.
    return adapter.genomizer.RepairIndividual(concreteInd)
}

// UpdatePatternLibrary delegates to the Genomizer..
func (adapter *GenomizerImmuneAdapter) UpdatePatternLibrary(individuals []IIndividual) {

    // Call to the existing Genomizer method.
    adapter.genomizer.UpdatePatternLibrary(individuals)
}

// UpdateSuccessfulProductions delegates to the Genomizer.
func (adapter *GenomizerImmuneAdapter) UpdateSuccessfulProductions(individuals []IIndividual) {
    
    // Call to the existing Genomizer method.
    adapter.genomizer.UpdateSuccessfulProductions(individuals)
}