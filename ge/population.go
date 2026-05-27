// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The Individuals package implements a population of genetic individuals.
package ge

import (
	"fmt"

	// "runtime/debug"
)

/* Population */

// Population represents a group of individuals sharing a common Genomizer.
// It manages genetic operations (correction, production updates) at the
// population level.
type Population struct {

	// genomizer is the shared genetic context for all individuals in the 
	// population.
	genomizer   IGenomizer
	
	immuneSys   IImmune
    
	// individuals is the list of individuals in the population, sorted by 
	// fitness.
	individuals []IIndividual
}

// Create initializes a Population with a list of individuals, a Genomizer and
// an Immune System.
func (pop *Population) Create(
	individuals []IIndividual, 
	genomizer IGenomizer,
	immuneSys IImmune,
) (*Population, error) {

	// 1. Verify that genomizer is not nil.
	if genomizer == nil {
        return nil, fmt.Errorf("genomizer cannot be nil")
    }

	// 2. Check that individuals is not empty.
    if len(individuals) == 0 {
        return nil, fmt.Errorf("individuals cannot be empty")
    }

	// 3. Check that immune system is not nil.
	if immuneSys == nil {
        return nil, fmt.Errorf("immune system cannot be nil")
    }

	// 4. Initialize the population.
	pop.genomizer = genomizer
	pop.individuals = individuals
	pop.immuneSys = immuneSys

	return pop, nil
}

// AddToFailedProductions adds a failed production to the list of failed 
// productions.
func (pop *Population) AddToFailedProductions(production []IRuleModel, fitness float64) {

    // -- Debug --
	// defer func() {
    //    if r := recover(); r != nil {
    //        fmt.Printf("Panic in AddToFailedProductions: %v\n", r)
    //        debug.PrintStack()  // Affiche la stack trace
    //    }
    // }()

    pop.immuneSys.AddToFailedProductions(production, fitness)
}

// CorrectByGenome corrects an individual using their genome.
func (pop *Population) CorrectByGenome(
    ind IIndividual, 
    population []IIndividual, 
    fitnessThreshold float64, 
    averageFitness float64,
    fitnessFunction FitnessFunc,
) (bool, error)	{
	return pop.immuneSys.CorrectByGenome(ind, population, fitnessThreshold, averageFitness, fitnessFunction)
}

// CorrectByGrammaticalPaths corrects an individual using grammatical paths.
func (pop *Population) CorrectByGrammaticalPaths(
    ind IIndividual,
    fitnessThreshold float64,
    fitnessFunction FitnessFunc,
) (bool, error) {
	return pop.immuneSys.CorrectByGrammaticalPaths(ind, fitnessThreshold, fitnessFunction)
}

// CorrectByTemplate corrects an individual using a template.
func (pop *Population) CorrectByTemplate(
    ind IIndividual,
    templateFunction TemplateFunc,
    fitnessFunction FitnessFunc,
) (bool, error) {
    return pop.immuneSys.CorrectByTemplate(ind, templateFunction, fitnessFunction)
}       	

func (pop *Population) GetIndividuals() []IIndividual {
    individuals := make([]IIndividual, len(pop.individuals))
	copy(individuals, pop.individuals)
    return individuals
}

// Size returns the number of individuals in the population.
func (pop *Population) Size() int {
    return len(pop.individuals)
}

// UpdateSuccessfulProductions updates successful productions in the Genomizer 
// for a given list of individuals.
func (pop *Population) UpdateSuccessfulProductions(individuals []IIndividual) {
    pop.immuneSys.UpdateSuccessfulProductions(individuals)
}

// UpdatePatternLibrary updates the pattern library in the Genomizer for a 
// given list of individuals.
func (pop *Population) UpdatePatternLibrary(individuals []IIndividual) {
    pop.immuneSys.UpdatePatternLibrary(individuals)
}
