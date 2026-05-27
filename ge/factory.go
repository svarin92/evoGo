// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

import (
	"fmt"
	
	"evoGo/controller"
	"evoGo/model"
)

/* Exports */

// NewPopulation creates a new population with a shared Genomizer.
func NewPopulation(size int, grammar IGrammar) (*Population, error) {
	genomizer := controller.NewGenomizer(grammar)

	// Create the temporary adapter.
	immuneSys := controller.NewGenomizerImmuneAdapter(genomizer)

	// New array of Individuals with capacity for size elements.
	population := make([]IIndividual, size)

	for i := range population {
		genome := GenerateRandomGenome()

		// Creates an individual and sets its genome to default values.
		population[i] = model.NewIndividual(genome)

		if err := population[i].GeneratePhenotype(genomizer); err != nil {
			return nil, fmt.Errorf(
				"failed to generate phenotype for individual %d: %w", i, err,
			)
		}

	}

	return new(Population).Create(population, genomizer, immuneSys)
}