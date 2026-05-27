// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

import (
    "math/rand"
	"sort"
)

// Helper to check if all individuals in a range are identical.
func AreAllIndividualsIdentical(individuals []IIndividual) bool {

    if len(individuals) == 0 {
        return true
    }

    first := individuals[0]

    for _, ind := range individuals[1:] {
    
		if ind.GetFitness() != first.GetFitness() || ind.GetPhenotype() != first.GetPhenotype() {
            return false
        }
    
	}
    
	return true
}

// Helper to find the best individual.
func FindBestIndividual(individuals []IIndividual) IIndividual {

	if len(individuals) == 0 {
        return nil
    }

    best := individuals[0]
    
	for _, ind := range individuals[1:] {
    
		if ind.GetFitness() > best.GetFitness() {
            best = ind
        }
    
	}
    
	return best
}

// GenerateRandomGenome generates a random genome of size CODONS_SIZE.
func GenerateRandomGenome() []int {
	genome := make([]int, CODONS_SIZE)

	for i := range genome {
		genome[i] = rand.Intn(CODONS_SIZE)
	}

	return genome
}

// SortDescending sorts a list of individuals by decreasing fitness.
func SortDescending(individuals []IIndividual) []IIndividual {
	sort.Slice(individuals, func(i, j int) bool {
		return individuals[i].GetFitness() > individuals[j].GetFitness()
	})
	return individuals
}
