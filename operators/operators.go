// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package operators

import (
	"log"
	"math/rand"
	"sort"

	"evoGo/model"
)

// GenerationalReplacement replaces a population with a new one, while 
// retaining the elites of the old population.
func GenerationalReplacement(newPop, oldPop []IIndividual, eliteSize int) []IIndividual {

    // Sort populations by decreasing fitness.
    sort.Slice(oldPop, func(i, j int) bool {
        return oldPop[i].GetFitness() > oldPop[j].GetFitness()
    })
    sort.Slice(newPop, func(i, j int) bool {
        return newPop[i].GetFitness() > newPop[j].GetFitness()
    })

    // Retains the elites of the old population.
    for i := 0; i < eliteSize && i < len(oldPop); i++ {
        newPop = append(newPop, oldPop[i].Copy())
    }

    // Sort and return the new population.
    sort.Slice(newPop, func(i, j int) bool {
        return newPop[i].GetFitness() > newPop[j].GetFitness()
    })

    if len(newPop) > len(oldPop) {
        newPop = newPop[:len(oldPop)]
    }

    return newPop
}

// Mutation is a key operator in genetic or evolutionary algorithms. It 
// introduces genetic diversity into the population by randomly modifying 
// part of an individual's genome. Without mutation, the population could 
// converge too quickly toward a local
//
// optimum, without sufficiently exploring the solution space. It allows 
// the discovery of new areas of the search space, to prevent all individuals
// from becoming too similar and to complete the crossing by providing random 
// variations.
//
// Variations are possible: In adaptive mutation, the probability of mutation 
// evolves according to the diversity of the population or the number of 
// generations. While the probability of non-uniform mutation decreases over 
// time to favor convergence. It is still possible to operate by exchanging 
// codon sequences between individuals.
func Mutate(individuals []IIndividual, mutationProbability float64) []IIndividual {

	for _, ind := range individuals {

    	if len(ind.GetGenome()) != CODONS_SIZE {
        	log.Printf("WARNING: Genome size is %d, expected %d\n", len(ind.GetGenome()), CODONS_SIZE)
    	}

		for i := range ind.GetGenome() {
		// for i := 0; i < len(ind.genome); i++ {

			// For each codon in the genome, we draw a random number.
			if rand.Float64() < mutationProbability {

				// Limit the mutation to valid codons.
				ind.MutateCodon(i, rand.Intn(CODONS_SIZE))
			}

		}

	}

	// Returns the list of individuals after mutation.
	//
	// Let's assume an individual with a genome of 5 codons and 
	// mutationProbability = 0.1 (10% chance per codon):
	//
	// ind.genome = [10, 20, 30, 40, 50]
	//
	// Possible outcome after mutation:
	//
	// The codon at index 1 (value 20) is mutated to 85.
	// The codon at index 3 (value 40) is mutated to 5.
	// New genome: [10, 85, 30, 5, 50].
	return individuals
}

// Given two individuals, create two children using one-point crossover and 
// return them.
//
// One-point crossover is a classic method used in genetic algorithms to 
// combine the genomes of two parents to produce two children. The idea is 
// to choose a random cutoff point in the genomes of both parents and swap 
// the segments after this point between the two parents to create two new 
// individuals (children). This allows the genetic characteristics of the 
// parents to be mixed, promoting the exploration of new solutions while 
// retaining part of the genetic information of the parents.
//
// This method can present potential disruptions if the crossover point cuts 
// in the middle of an important genetic sequence and produces less suitable 
// offspring. It can also introduce  positional bias: genes near the beginning 
// or end of the genome are less likely to be mixed.
//
// Variations are possible: two-point crossover cut points for added diversity,
// with uniform crossover, each gene has an independent probability of being 
// exchanged, and the arithmetic crossing which proceeds by linear combination 
// of the values ​​of the genes.
func OnePointCrossover(parent0, parent1 IIndividual, withinUsed bool) (IIndividual, IIndividual) {

	// The crossing is not systematic. We draw a random number between 0 and 1:
	// If this number is less than or equal to 0.7, the crossover is performed. 
	// Otherwise, we
	// simply return copies of the parents (without crossover). This helps 
	// maintain a certain diversity and avoids overly rapid convergence.
	// crossoverProbability := 0.7: probability of making a crossover (70%)

	if rand.Float64() > CROSSOVER_PROBABILITY {

		// If the probability is not respected, we return copies of the 
		// parents.
		return parent0.Copy(), parent1.Copy()
	}

	// Determines the random crossover point.
	//
	// The maximum length between the two parental genomes is determined to 
	// avoid index errors.
	maxLen := len(parent0.GetGenome())

	if withinUsed {
		usedCodons0 := parent0.GetUsedCodons()
        usedCodons1 := parent1.GetUsedCodons()

		// The notion of "used part" of the genome (usedCodons) is explicitly 
		// used to limit crossover.
		maxLen = min(usedCodons0, usedCodons1)
	} else if len(parent1.GetGenome()) < maxLen {

		// If parent0.genome is longer than parent1.genome, the length of 
		// parent1.genome is used (and vice versa).
		maxLen = len(parent1.GetGenome())
	}

	crossoverPoint := rand.Intn(maxLen)  // Cutoff point between 0 and maxLen-1

	// Create children.
	//
	// We initialize two new individuals (child0 and child1) with genomes of 
	// the same size as their respective parents.
	child0 := model.NewIndividual(make([]int, len(parent0.GetGenome())))
	child1 := model.NewIndividual(make([]int, len(parent1.GetGenome())))

	// Copy segments before/after the crossing point.
	//
	// child0 inherits the first part from parent0 and the second part from 
	// parent1.child1 inherits the first part from parent1 and the second part 
	// from parent0.
	child0.UpdateGenomeSegment(parent0.GetGenome()[:crossoverPoint], 0)
	child0.UpdateGenomeSegment(parent1.GetGenome()[crossoverPoint:], crossoverPoint)
	child1.UpdateGenomeSegment(parent1.GetGenome()[:crossoverPoint], 0)
	child1.UpdateGenomeSegment(parent0.GetGenome()[crossoverPoint:], crossoverPoint)

	// Let's assume:
	//
	// parent0.genome = [1, 2, 3, 4, 5]
	// parent1.genome = [6, 7, 8, 9, 10]
	// crossoverPoint = 2
	//
	// Result after crossing:
	//
	// child0.genome = [1, 2, 8, 9, 10] (first part of parent0, second part of 
	// parent1)
	// child1.genome = [6, 7, 3, 4, 5] (first part of parent1, second part of 
	// parent0)
	return child0, child1
}

/* Selection */

// Two selection methods: tournament and truncation.

// Given an entire population, draw tournamentSize competitors randomly 
// and return the best.
//
// Tournament selection is a common method for choosing which individuals 
// (or "parents") will reproduce in an evolutionary algorithm. It allows 
// for controlling selection pressure (i.e., the extent to which the best 
// individuals are favored). We organize "tournaments" between a small 
// group of individuals drawn randomly from the population. The winner of 
// the tournament (the individual with the best fitness) is selected for 
// breeding. This process is repeated until there are enough parents to 
// form the next generation.
//
// Tournament selection helps maintain a balance between exploration 
// (diversity) and exploitation (quality). How to choose tournamentSize? 
// Small (2-3): Less pressure, more diversity, slower convergence. Large 
// (5+): More pressure, faster convergence, but risk of losing diversity.
// Typical: 3 or 4 is a good compromise for most problems. We can modify 
// the pressure differently using variants such as. In tournament Selection 
// with probability, the winner is not always the best, but has a higher 
// probability of being selected. While in the stochastic version, even the 
// worst player has a small chance of being selected.
//
// Returns a slice of references to the original individuals (no copies).
//  Warning: The returned individuals must not be modified directly, as this 
// would also affect the original population.
func TournamentSelection(population []IIndividual, tournamentSize int) []IIndividual {

	// List of winners (selected parents).
	selected := make([]IIndividual, 0, len(population))

	for len(selected) < len(population) {  // We fill until we have a complete population

		// Group of competitors for a tournament. Tournament Size: the number 
		// of individuals competing in each tournament (e.g., 3). By adjusting 
		// tournamentSize, one can influence the selection pressure. 2 favors 
		// a more random selection (less pressure). 5 or more favors the best 
		// individuals (more pressure). If tournamentSize = population size, 
		// this is equivalent to selecting the best individual (very high 
		// pressure).
		tournament := make([]IIndividual, tournamentSize)

		for i := range tournament {

			// We randomly draw an individual from the population.
			tournament[i] = population[rand.Intn(len(population))]
		}

		// Competitors are sorted by decreasing fitness (best first).
		sort.Slice(tournament, func(i, j int) bool {
			return tournament[i].GetFitness() > tournament[j].GetFitness()
		})

		// The tournament winner is added to the parents list.
		selected = append(selected, tournament[0])
	}

	// The function returns a list of parents (winners), which will be used 
	// for reproduction (crossing and mutation).
	return selected
}
