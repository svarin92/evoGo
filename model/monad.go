// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"fmt"
	"evoGo/utils"
)

/* Individual */

// Individual represents an individual in an evolving algorithm.
type Individual struct {
	fitness              float64  // Fitness value
	genome               []int    // Genome of the individual, represented as an array of integers
	phenotype            any      // Generic phenotype 
	template             any
	organism             IOrganism
	usedCodons           int      // Number of used codons by the individual
	usedWraps            int
	compiledPhenotype    any      // Compiled phenotype if applicable
	productionHistory    [][]IRuleModel
	oldProductionFitness map[string]float64
}

// Create initializes a new individual with a given genome.
func (ind *Individual) Create(genome []int) *Individual {

	if genome == nil {
		ind.genome = []int{}
	} else {
		ind.genome = genome
	}

	ind.fitness = 0
	ind.phenotype = ""
	ind.template = ""
	ind.usedCodons = 0
	ind.compiledPhenotype = nil
	ind.productionHistory = [][]IRuleModel{}
	ind.oldProductionFitness = make(map[string]float64)
	return ind
}

// Copy creates a deep copy of an individual.
func (ind *Individual) Copy() IIndividual { 
	newInd := NewIndividual(nil)

    // Genome copy.
    genome := make([]int, len(ind.GetGenome()))
    copy(genome, ind.GetGenome())
    newInd.SetGenome(genome)

    newInd.SetFitness(ind.GetFitness())
    newInd.SetPhenotype(ind.GetPhenotype())
    newInd.SetUsedCodons(ind.GetUsedCodons())
    newInd.SetTemplate(ind.GetTemplate())
    newInd.SetProductionHistory(utils.DeepCopyProductionHistory(ind.GetProductionHistory()))
	return newInd
}

// Evaluate assesses the individual's phenotype and updates their fitness.
func (ind *Individual) Evaluate(fitness FitnessFunc) error {
	ind.fitness = fitness(ind)
	return nil
}

func (ind *Individual) GetFitness() float64 {
	return ind.fitness
}

func (ind *Individual) GetGenome() []int {
    return ind.genome
}

func (ind *Individual) GetOldProductionFitness(key string) (float64, bool) {
    oldFitness, exists := ind.oldProductionFitness[key]
    return oldFitness, exists
}

// GetOrganism returns the organism associated with the individual.
func (ind *Individual) GetOrganism() IOrganism {
	return ind.organism
}

func (ind *Individual) GetPhenotype() any {
    return ind.phenotype
}

// GetProductionHistory returns a deep copy of the production history to 
// ensure data isolation between individuals.
func (ind *Individual) GetProductionHistory() [][]IRuleModel {
	return utils.DeepCopyProductionHistory(ind.productionHistory)
}

// Getter pour template.
func (ind *Individual) GetTemplate() any {
    return ind.template
}

func (ind *Individual) GetUsedCodons() int {
    return ind.usedCodons
}

// GeneratePhenotype generates the individual's phenotype using a Genomizer.
func (ind *Individual) GeneratePhenotype(genomizer IGenomizer) error {

    if err := genomizer.Genomize(ind.genome); err != nil {
        return fmt.Errorf("failed to genomize: %w", err)
    }

	ind.phenotype = genomizer.GetPhenotype()
    ind.productionHistory = utils.DeepCopyProductionHistory(genomizer.GetProductionHistory())
	ind.usedCodons = genomizer.GetUsedCodons()
	return nil
}

func (ind *Individual) MutateCodon(index int, value int) {

    if index >= 0 && index < len(ind.genome) {
        ind.genome[index] = value
    }

}

// Repair uses the immune system to repair the individual.
func (ind *Individual) Repair(immuneSys IImmune) error {
    return immuneSys.RepairIndividual(ind)
}

func (ind *Individual) SetFitness(value float64) {
    ind.fitness = value
}

func (ind *Individual) SetGenome(value []int) {
    ind.genome = value
}

func (ind *Individual) SetOldProductionFitness(key string, value float64) {
    
	if ind.oldProductionFitness == nil {
        ind.oldProductionFitness = make(map[string]float64)
    }
    
	ind.oldProductionFitness[key] = value
}

// SetOrganism associates an organism with the individual.
func (ind *Individual) SetOrganism(value IOrganism) {
	ind.organism = value
}

// Setter for phenotype.
func (ind *Individual) SetPhenotype(value any) {
    ind.phenotype = value
}

// Setter for productionHistory.
func (ind *Individual) SetProductionHistory(value [][]IRuleModel) {
    ind.productionHistory = value
}

// SetProductionStep replaces a specific step in productionHistory.
func (ind *Individual) SetProductionStep(index int, step []IRuleModel) {

    if index < 0 || index >= len(ind.productionHistory) {
        panic("index out of range")
    }
    
	// Deep copy of the new step (to avoid shared references).
    newStep := make([]IRuleModel, len(step))
    for j, rule := range step {
        newStep[j] = rule.Clone() // Use Clone() for deep copying
    }
    
	ind.productionHistory[index] = newStep
}

// Setter for template.
func (ind *Individual) SetTemplate(value any) {
    ind.template = value
}

// Setter for usedCodons.
func (ind *Individual) SetUsedCodons(value int) {
    ind.usedCodons = value
}

// String returns a textual representation of the individual.
func (ind *Individual) String() string {

	// Returns a formatted string representation of the Individual, including
	// its phenotype, fitness and used codons.
	return fmt.Sprintf(
		"phenotype = %s, fitness = %.2f, Used Codons: %d", 
		ind.phenotype, 
		ind.fitness,
		ind.usedCodons,
	)
	
}

func (ind *Individual) UpdateGenomeSegment(segment []int, offset int) {
    copy(ind.genome[offset:offset+len(segment)], segment)
}