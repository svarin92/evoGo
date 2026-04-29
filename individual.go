// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package main

import (
	"fmt"
)

/* Individual */

type (
	
	// Individual defines the contract for an individual in an evolving 
	// algorithm.
	IIndividual interface {
    	Evaluate(fitness FitnessFunc) error
    	GeneratePhenotype(genomizer IGenomizer) error
    	Repair(genomizer IGenomizer) error
    	GetProductionHistory() [][]IRuleModel
    	GetFitness() float64
	}

	// Individual represents an individual in an evolving algorithm.
	Individual struct {
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

)

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

// Evaluate assesses the individual's phenotype and updates their fitness.
func (ind *Individual) Evaluate(fitness FitnessFunc) error {
	ind.fitness = fitness(ind)
	return nil
}

// GetOrganism returns the organism associated with the individual.
func (ind *Individual) GetOrganism() IOrganism {
	return ind.organism
}

// GetProductionHistory returns a deep copy of the production history to 
// ensure data isolation between individuals.
func (ind *Individual) GetProductionHistory() [][]IRuleModel {
	return DeepCopyProductionHistory(ind.productionHistory)
}

// GeneratePhenotype generates the individual's phenotype using a Genomizer.
func (ind *Individual) GeneratePhenotype(genomizer IGenomizer) error {

    if err := genomizer.Genomize(ind.genome); err != nil {
        return fmt.Errorf("failed to genomize: %w", err)
    }

	ind.phenotype = genomizer.GetPhenotype()
    ind.productionHistory = DeepCopyProductionHistory(genomizer.GetProductionHistory())
	ind.usedCodons = genomizer.GetUsedCodons()
	return nil
}

// Repair repairs the individual using a Genomizer.
func (ind *Individual) Repair(genomizer IGenomizer) error {
    return genomizer.RepairIndividual(ind)
}

// SetOrganism associates an organism with the individual.
func (ind *Individual) SetOrganism(value IOrganism) {
	ind.organism = value
}

// String returns a textual representation of the individual.
func (ind *Individual) String() string {

	// Returns a formatted string representation of the Individual, including
	// its phenotype and fitness.
	return fmt.Sprintf("phenotype = %s; fitness = %.2f", ind.phenotype, ind.fitness)
}

/* Exports */

// NewIndividual is a factory function to create a new individual.
func NewIndividual(genome []int) *Individual {
	return new(Individual).Create(genome)
}