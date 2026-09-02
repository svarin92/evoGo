// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package examples

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"evoGo/cell"
	"evoGo/model"
)

// CellularEvolutionExample demonstrates the use of DigitalCells with go-actor
// to create a cellular evolutionary system with membranes.
func CellularEvolutionExample() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// For now, we'll create a simple population of cells without the full grammar
	// This is a simplified example to demonstrate the cellular architecture
	
	populationSize := 10
	
	// Create cell system
	system := cell.NewCellSystem()
	
	// Create cells
	for i := 0; i < populationSize; i++ {
		// Create genome
		genome := make([]int, 50)
		for j := range genome {
			genome[j] = rand.Intn(256)
		}
		
		// Create individual
		individual := model.NewIndividual(genome)
		
		// Create membrane
		membrane := cell.NewMembraneWithParams(0.5, 0.7, nil)
		
		// Create cell
		_, err := system.CreateCell(individual, membrane)
		if err != nil {
			log.Printf("Failed to create cell %d: %v", i, err)
			continue
		}
	}

	// Connect all cells
	err := system.ConnectAll()
	if err != nil {
		log.Printf("Failed to connect cells: %v", err)
		return
	}

	// Run simulation for a few generations
	generations := 5
	for gen := 0; gen < generations; gen++ {
		fmt.Printf("\n=== Generation %d ===\n", gen)
		
		// Send tick message to all cells
		msg := cell.NewMsgTick(gen, populationSize)
		err := system.BroadcastToAll(msg)
		if err != nil {
			log.Printf("Failed to broadcast tick: %v", err)
			continue
		}
		
		// Wait a bit for messages to propagate
		time.Sleep(100 * time.Millisecond)
		
		// Print status
		cells := system.GetAllCells()
		for i, cellActor := range cells {
			cell := cellActor.GetCell()
			fmt.Printf("Cell %d: Fitness=%.2f, PID=%v\n", i, cell.GetFitness(), cell.GetPID())
		}
	}

	// Shutdown
	system.Shutdown()
	fmt.Println("Simulation complete")
}

// CellularEvolutionWithTarget demonstrates evolution towards a specific target.
func CellularEvolutionWithTarget(target string) {
	// This is a more complete example that shows how to evolve cells towards a target
	// using the membrane for communication and the immune system for correction.
	
	// For now, this is a placeholder
	// A full implementation would require integrating with the existing evoGo
	// evolutionary loop and replacing the population with cells.
	
	fmt.Printf("Cellular evolution towards target: %s\n", target)
	fmt.Println("(Implementation placeholder - see CellularEvolutionExample for working code)")
}
