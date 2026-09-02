// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package cell

import (
	"evoGo/controller"
	"evoGo/interfaces"
	"evoGo/model"
)

// NewDigitalCellFromGenome creates a new DigitalCell with a random genome.
func NewDigitalCellFromGenome(genome []int, grammar interfaces.IGrammar) (*DigitalCell, error) {
	// Create genomizer and immune system
	genomizer := controller.NewGenomizer(grammar)
	immuneSys := controller.NewGenomizerImmuneAdapter(genomizer)

	// Create individual
	individual := model.NewIndividual(genome)
	
	// Generate phenotype
	if err := individual.GeneratePhenotype(genomizer); err != nil {
		return nil, err
	}

	// Create membrane with immune system
	membrane := NewMembraneWithParams(0.5, 0.7, immuneSys)

	// Create and return the digital cell
	return NewDigitalCell(individual, membrane), nil
}

// NewDigitalCellPopulation creates a population of DigitalCells.
func NewDigitalCellPopulation(size int, grammar interfaces.IGrammar) ([]*DigitalCell, error) {
	population := make([]*DigitalCell, size)
	
	for i := 0; i < size; i++ {
		genome := generateRandomGenome(127) // Use same size as config
		cell, err := NewDigitalCellFromGenome(genome, grammar)
		if err != nil {
			return nil, err
		}
		population[i] = cell
	}
	
	return population, nil
}

// generateRandomGenome generates a random genome of the given size.
func generateRandomGenome(size int) []int {
	import "math/rand"
	genome := make([]int, size)
	for i := range genome {
		genome[i] = rand.Intn(256) // Random codon value
	}
	return genome
}

// NewCellSystemFromPopulation creates a CellSystem from a population of individuals.
func NewCellSystemFromPopulation(
	individuals []interfaces.IIndividual,
	grammar interfaces.IGrammar,
) (*CellSystem, error) {
	// Create genomizer and immune system
	genomizer := controller.NewGenomizer(grammar)
	immuneSys := controller.NewGenomizerImmuneAdapter(genomizer)

	// Create cell system
	system := NewCellSystem()

	// Create cells from individuals
	for _, ind := range individuals {
		// Create membrane with shared immune system
		membrane := NewMembraneWithParams(0.5, 0.7, immuneSys)
		
		// Create cell actor
		cell, err := system.CreateCell(ind, membrane)
		if err != nil {
			return nil, err
		}
		
		// Keep reference to the cell
		_ = cell
	}

	return system, nil
}

// NewCellSystemWithGrid creates a CellSystem with a grid topology.
func NewCellSystemWithGrid(
	rows, cols int,
	grammar interfaces.IGrammar,
) (*CellSystem, error) {
	// Create cell system
	system := NewCellSystem()

	// Create genomizer and immune system (shared for now)
	genomizer := controller.NewGenomizer(grammar)
	immuneSys := controller.NewGenomizerImmuneAdapter(genomizer)

	// Create cells in a grid
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Create genome
			genome := generateRandomGenome(127)
			
			// Create individual
			individual := model.NewIndividual(genome)
			if err := individual.GeneratePhenotype(genomizer); err != nil {
				return nil, err
			}
			
			// Create membrane
			membrane := NewMembraneWithParams(0.5, 0.7, immuneSys)
			
			// Create cell
			_, err := system.CreateCell(individual, membrane)
			if err != nil {
				return nil, err
			}
		}
	}

	// Connect cells in a grid
	err := system.ConnectNeighbors(rows * cols)
	if err != nil {
		return nil, err
	}

	return system, nil
}
