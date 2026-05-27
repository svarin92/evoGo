// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

import (
	"fmt"
	"log"
	"math/rand"

	"evoGo/evaluator"
	"evoGo/model"
	"evoGo/operators"
	"evoGo/renderer"
	"evoGo/utils"
)

// Evaluate initial population.
func EvaluateFitness(
	population *Population,
	fitnessFunction FitnessFunc,
) error {

	// Scan each individual in the population.
	for _, ind := range population.individuals {

		if ind.GetPhenotype() != nil && ind.GetPhenotype() != "" {

			// Assess the individual's fitness.
			if err := ind.Evaluate(fitnessFunction); err != nil {
				return fmt.Errorf("error during Evaluate: %w", err)
			}

		}

		// If the fitness level is below a threshold, mark the productions as 
		// failed.
		if ind.GetFitness() < 0.2 {

    		for _, production := range ind.GetProductionHistory() {
        		population.AddToFailedProductions(production, ind.GetFitness())
    		}
		
		}

	}

	// Update successful productions for the current population.
    population.UpdateSuccessfulProductions(population.GetIndividuals())

	return nil
}

// SearchLoop executes the grammatical evolution (GE) loop to find an optimal 
// individual that matches the target.
//
// Parameters:
// - maxGenerations: Maximum number of generations to simulate.
// - populationSize: Population size at each generation.
// - grammar: Grammar used to generate phenotypes.
// - target: Target to reach (e.g., a string like "golden").
// - replacementFunc: Function to replace individuals between generations 
//   (e.g., GenerationalReplacement).
// - selectionFunc: Function to select parents (e.g., TournamentSelection).
// - fitnessFunction: Function to evaluate the fitness of individuals.
//
// Returns:
// - IIndividual: The best individual found after evolution.
// - error: An error if evolution fails (e.g., empty population, invalid 
//   grammar).
func SearchLoop(
	maxGenerations int,
	populationSize int,
	grammar IGrammar,
	target any,
	replacementFunc ReplacementFunc,
	  // func([]IIndividual, []IIndividual, int) []IIndividual,
	selectionFunc  SelectionFunc,
	  // func([]IIndividual, int) []IIndividual,
	fitnessFunction FitnessFunc,
) (IIndividual, error) {

	// Create a template function.
    templateFunc := evaluator.NewTemplate(target)

	// Create the initial population.
    population, err := NewPopulation(populationSize, grammar)

    if err != nil {
        return nil, fmt.Errorf("failed to create population: %w", err)
    }

	// Create a Renderer for the rendering functions.
    renderer := renderer.NewRenderer(population.genomizer)	

	// Define the size of the elite and the tournament.
    eliteSize := max(int(float64(populationSize) * 0.1), 1)
    
	tournamentSize := 3

	// Evaluate initial population.
	if err := EvaluateFitness(population, fitnessFunction); err != nil {
		return nil, fmt.Errorf("failed to evaluate initial fitness: %w", err)
	}

	// Finding the best initial individual.
	bestEver := FindBestIndividual(population.individuals)
	bestGeneration := 1  // Initialize with the first generation

	// Sort a list of individuals in descending order (i.e., from highest 
	// to lowest).
	// individuals = sort.Sort(sort.Reverse(individuals))
	population.individuals = SortDescending(population.individuals)

	// Print statistics at the fisrt generation.
	renderer.PrintStats(population, 1, GENERATION_SIZE)

	// Iterate through generations using the step function to evolve 
	// populations, keeping track of the best individual found so far 
	// and printing statistics at each generation.
	var generation int = 2

	for ; generation <= maxGenerations+1; generation++ {
		var err error
		var currentBestEver IIndividual
		population, currentBestEver, err = Step(
			population, 
			grammar, 
			templateFunc, 
			replacementFunc, 
			selectionFunc, 
			fitnessFunction, 
			eliteSize, 
			tournamentSize, 
			bestEver,
		)

		if err != nil {
			return nil, fmt.Errorf("failed at generation %d: %w", generation, err)
		}

		// Update bestEver and bestGeneration if a new best individual is 
		// found.
        if currentBestEver.GetFitness() > bestEver.GetFitness() {
            bestEver = currentBestEver
            bestGeneration = generation  // Update the generation of the best individual
        }

		renderer.PrintStats(population, generation, GENERATION_SIZE)
	}

    // Display the best individual details.
    renderer.DisplayIndividualDetails(bestEver, bestGeneration)

    // Display the grammatical derivation.
    renderer.PrintGrammaticalDerivation(bestEver)

    // Export the derivation to DOT format.
    if err := renderer.ExportToDOT(bestEver, "best_ever.dot"); err != nil {
        log.Printf("Failed to export to DOT: %v", err)
    } else {
		log.Printf("Successfully exported derivation to best_individual.dot")
	}

	return bestEver, nil
}

// Perform some evolutionary computation steps (e.g., selection, Random
// matching, crossover, mutation). Step() first exploits good solutions 
// via crossover, then explores around these solutions via mutation.
//
// The crossover → mutation chain is a standard and relevant practice 
// in evolutionary algorithms (such as genetic algorithms or genetic 
// programming). Crossbreeding combines the characteristics of two parents 
// to produce children that inherit their "good" properties. But, it does 
// not create new genetic information, it simply rearranges what already 
// exists. Without mutation, the population could quickly stagnate (lack of
// diversity). 
// Mutation introduces new random variations into the genome. In nature, 
// mutations are random errors during DNA replication. While often neutral or 
// harmful, they are sometimes beneficial and enable evolution. Mutation allows 
// exploring new areas of the search space (avoiding local optima), maintaining 
// genetic diversity in the population, essential to avoid premature 
// convergence, and "repairing" solutions that might be stuck in a suboptimal
// state.
func Step(
	population *Population,
	grammar IGrammar,
	templateFunction TemplateFunc,
	replacementFunc ReplacementFunc,
	  // func([]IIndividual, []IIndividual, int) []IIndividual,
	selectionFunc SelectionFunc,
	  // func([]IIndividual, int) []IIndividual,
	fitnessFunction FitnessFunc,
    eliteSize int,
    tournamentSize int,	
	bestEver IIndividual,
) (*Population, IIndividual, error) {

	// 1. Selection of parents.
	//
	// Selection creates a pool of potential parents from the current 
	// population (individuals). But, it does not yet randomly choose two 
	// specific parents from the parent pool to perform a cross. By randomly 
	// choosing pairs (matching) from this pool, the algorithm is allowed to 
	// explore different combinations of genomes, which is crucial for 
	// diversity.
	//
	// For each parent, we organize a tournament between 3 random individuals,
	// and we keep the best one. This creates moderate selection pressure,
	// favoring the best individuals without completely eliminating diversity.
	iParents := selectionFunc(population.individuals, tournamentSize)  // tournamentSize = 3

	// Verification of the uniqueness of individuals in parents.
    if AreAllIndividualsIdentical(iParents) {
        
        // Return a copy of the current population.
        return population, bestEver, nil
    }

	// 2. Crossover parents and add to the new population.
	//
	// GENERATION_SIZE must be even (because each crossover produces two 
	// children).

	// Creation of the new population.
	newIndividuals := make([]IIndividual, 0, len(population.individuals))

	for len(newIndividuals) < GENERATION_SIZE {

		// Random matching.
		p0, p1 := iParents[rand.Intn(len(iParents))], iParents[rand.Intn(len(iParents))]

		// Avoid p0 == p1: we avoid crossing a parent with himself.
		for p0 == p1 {
			p1 = iParents[rand.Intn(len(iParents))]
		}

		// Crossover.
		child0, child1 := operators.OnePointCrossover(p0, p1, true)

		newIndividuals = append(newIndividuals, child0, child1)
	}

	// Truncation if necessary (if GENERATION_SIZE is odd).
	if len(newIndividuals) > GENERATION_SIZE {
		newIndividuals = newIndividuals[:GENERATION_SIZE]
	}

	// 3. Population mutation by codon.
	//
	// Each codon is mutated independently of the others. This allows for 
	// detailed exploration of the genome without disrupting the individual too
	// drastically. A typical probability is 0.01 to 0.1 (1% to 10%). Too high, 
	// it could destroy good solutions. Conversely, too low, it could slow down 
	// exploration.
	newIndividuals = operators.Mutate(newIndividuals, MUTATION_PROBABILITY)  // mutationProbability = 0.01
	
	// Repair after mutation (Context 1): the mutation can corrupt the genome 
	// Early repair prevents errors from being propagated to subsequent 
	// generations.
    for i := range newIndividuals {

        if err := newIndividuals[i].Repair(population.genomizer); err != nil {
            log.Printf("Step: Repair impossible for individual %d after mutation: %v", i, err)
            
			// Option: replace with a new random individual.
            newIndividuals[i] = model.NewIndividual(GenerateRandomGenome())
        }

    }

    // 4. Generation of phenotypes of new individuals.
    for _, ind := range newIndividuals {

        if err := ind.GeneratePhenotype(population.genomizer); err != nil {
            return nil, nil, fmt.Errorf("failed to generate phenotype: %w", err)
        }

    }	

    // Repair after generation of the phenotype to correct inconsistencies in 
	// the production history.
    for i := range newIndividuals {

        if err := newIndividuals[i].Repair(population.genomizer); err != nil {
            log.Printf("Step: Repair impossible for individual %d after generation of the phenotype: %v", i, err)
            newIndividuals[i] = model.NewIndividual(GenerateRandomGenome())
        }

    }

	// 5. Evaluate the fitness of the new population.
	evalPop := &Population{
		individuals: newIndividuals,
		genomizer: population.genomizer,
		immuneSys:   population.immuneSys,
	}

	if err := EvaluateFitness(evalPop, fitnessFunction); err != nil {
		return nil, nil, fmt.Errorf("failed to evaluate fitness: %w", err)
	}

	// 6. Correction of suboptimal genomes using linguistic patterns.

	// Calculate the average fitness of the population.
	var fitnessVals []float64
	
	for _, ind := range newIndividuals {
	
		if ind.GetPhenotype() != nil {
        	fitnessVals = append(fitnessVals, ind.GetFitness())
    	}
	
	}

	averageFitness := utils.Average(fitnessVals)

    for i, ind := range newIndividuals {

		if ind.GetFitness() < 1.0 {
   
			// Apply the corrections in order.
            if ind.GetFitness() > 0.5 {

                if ok, err := population.CorrectByTemplate(ind, templateFunction, fitnessFunction); err != nil {
                    log.Printf("Template fix failure:: %v", err)

					// Repair after failure (Context 2): if a correction (e.g., 
					// CorrectByTemplate, CorrectGenome) fails or degrades 
					// fitness, the individual may be in an inconsistent state. 
					// A post-correction repair restores a valid state.
                	if repairErr := newIndividuals[i].Repair(population.genomizer); repairErr != nil {
                    	log.Printf("Failed to repair: %v", repairErr)
                	}

                } else if ok {
                    newIndividuals[i] = ind
                
					if ind.GetFitness() >= 1.0 { continue }  // Move to the next individual if the target is reached
                
				}

            }
			
            if ind.GetFitness() < 0.9 {

                if ok, err := population.CorrectByGenome(ind, newIndividuals, 0.7, averageFitness, fitnessFunction); err != nil {
                    log.Printf("Genome error: %v", err)
                } else if ok {
                    newIndividuals[i] = ind
                
					if ind.GetFitness() >= 1.0 { continue }  // Move to the next individual if the target is reached
                
				}

            }
    		
 			if ind.GetFitness() < 0.95 {

				if ok, err := population.CorrectByGrammaticalPaths(ind, 0.8, fitnessFunction); err != nil {
                    log.Printf("Grammatical error: %v", err)
                } else if ok {
                    newIndividuals[i] = ind
                }
            
			}			
						
		}
    
	}

	// 7. Updating meta-knowledge (before correction).
	population.UpdateSuccessfulProductions(newIndividuals)
    population.UpdatePatternLibrary(newIndividuals)

	// 8. Replace the sorted individuals with the new populations.
	population.individuals = replacementFunc(newIndividuals, population.individuals, eliteSize)  // eliteSize = 1

	// 9. Update best ever sample if needed.
	currentBest := FindBestIndividual(population.individuals)
    
	if currentBest.GetFitness() > bestEver.GetFitness() {
        bestEver = currentBest
    }

	return population, bestEver, nil
}