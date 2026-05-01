// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The Individuals package implements a population of genetic individuals.
package main

import (
	"fmt"
	"math/rand"
	"sort"
)

/* Population */

// Population represents a group of individuals sharing a common Genomizer. 
// It manages genetic operations (correction, production updates) at the 
// population level.
type Population struct {

	// genomizer is the shared genetic context for all individuals in the 
	// population.
	genomizer   IGenomizer
	
	immuneSys   IImmuneSystem
    
	// individuals is the list of individuals in the population, sorted by 
	// fitness.
	individuals []IIndividual
}

// Create initializes a Population with a list of individuals and a Genomizer.
func (pop *Population) Create(
	individuals []IIndividual, 
	genomizer IGenomizer,
	immuneSys IImmuneSystem,
) (*Population, error) {

	if genomizer == nil {
        return nil, fmt.Errorf("genomizer cannot be nil")
    }

    if len(individuals) == 0 {
        return nil, fmt.Errorf("individuals cannot be empty")
    }

	pop.genomizer = genomizer
	pop.individuals = individuals
	pop.immuneSys = immuneSys
	return pop, nil
}

// Add a failed production to the list of failed productions.
func (pop *Population) AddToFailedProductions(production []IRuleModel, fitness float64) {
    pop.immuneSys.AddToFailedProductions(production, fitness)
}

// Correct an individual using their genome.
func (pop *Population) CorrectByGenome(
    ind *Individual, 
    population []*Individual, 
    fitnessThreshold float64, 
    averageFitness float64,
    fitnessFunction FitnessFunc,
) (bool, error)	{
	return pop.immuneSys.CorrectByGenome(ind, population, fitnessThreshold, averageFitness, fitnessFunction)
}

// Correct an individual using grammatical pathways.
func (pop *Population) CorrectByGrammaticalPaths(
    ind *Individual,
    fitnessThreshold float64,
    fitnessFunction FitnessFunc,
) (bool, error) {
	return pop.immuneSys.CorrectByGrammaticalPaths(ind, fitnessThreshold, fitnessFunction)
}

// Correct an individual using a template.
func (pop *Population) CorrectByTemplate(
    ind *Individual,
    templateFunction TemplateFunc,
    fitnessFunction FitnessFunc,
) (bool, error) {
    return pop.immuneSys.CorrectByTemplate(ind, templateFunction, fitnessFunction)
}       	

// GetIndividuals returns a copy of the individuals slice to avoid external 
// modifications.
func (pop *Population) GetIndividuals() []IIndividual {
    individuals := make([]IIndividual, len(pop.individuals))
    copy(individuals, pop.individuals)
    return individuals
}

// Size returns the number of individuals in the population.
func (pop *Population) Size() int {
    return len(pop.individuals)
}

// Update successful productions in the Genomizer for a given list of 
// individuals.
func (pop *Population) UpdateSuccessfulProductions(individuals []*Individual) {
    pop.immuneSys.UpdateSuccessfulProductions(individuals)
}

// Update the pattern library in the Genomizer for a given list of 
// individuals.
func (pop *Population) UpdatePatternLibrary(individuals []*Individual) {
    pop.immuneSys.UpdatePatternLibrary(individuals)
}

/* Helpers */

// Generate a random genome of size CODONS_SIZE.
func GenerateRandomGenome() []int {
	genome := make([]int, CODONS_SIZE)

	for i := range genome {
		genome[i] = rand.Intn(CODONS_SIZE)
	}

	return genome
}

func SortDescending(individuals []IIndividual) []IIndividual {
	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].GetFitness() > individuals[j].GetFitness()
	})
	return individuals
}

/* Exports */

// Create a new population with a shared Genomizer.
func NewPopulation(size int, grammar IGrammar) (*Population, error) {
	genomizer := NewGenomizer(grammar)

	// Create the temporary adapter.
	immuneSys := NewGenomizerImmuneAdapter(genomizer)

	// New array of Individuals with capacity for size elements.
	population := make([]IIndividual, size)

	for i := range population {
		genome := GenerateRandomGenome()

		// Creates an individual and sets its genome to default values.
		population[i] = NewIndividual(genome)

		if err := population[i].GeneratePhenotype(genomizer); err != nil {
			return nil, fmt.Errorf(
				"failed to generate phenotype for individual %d: %w", i, err,
			)
		}

	}

	return new(Population).Create(population, genomizer, immuneSys)
}
