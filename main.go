package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// main.go - Implementation of a grammatical evolution (GE) algorithm to 
// generate optimal individuals based on a target (e.g., the word "golden"). 
// The program uses a grammar defined in an external file and a selection/
// tournaments loop to evolve towards the solution.

var CLI struct {
	File string `arg:"" type:"existingfile" help:"File to parse."`
}

func main() {
	ctx := kong.Parse(&CLI)

	// Open the grammar file provided as an argument.
	r, err := os.Open(CLI.File)
	ctx.FatalIfErrorf(err, "failed to open file")
	defer r.Close()

	// Create a new grammar based on the file.
	grammar, err := NewSpeciesGrammar(ctx, r)
	ctx.FatalIfErrorf(err, "failed to create species grammar")

	// Target for evolution (e.g., the word "golden").
	target := "golden"

	// Create an evaluation function that measures the proximity of individuals
	// to the target.
	fitnessFunc := NewFitness(target)

	// Grammatical Evolution (GE) - Execute the evolution algorithm over a 
	// given number of generations.
	//
	// 	 - GenerationalReplacement: Replaces older individuals with newer 
	// 	   ones, preserving elites.
	// 	 - TournamentSelection: Selects individuals for breeding via a 
	// 	   tournament.
	//
	bestEver, err := SearchLoop(
		GENERATIONS, 		// Total number of generations
		POPULATION_SIZE,	// Population size
		grammar,			// Grammar used to generate individuals
		target,				// Target to be reached
		func(newIndividuals, oldIndividuals []*Individual, eliteSize int) []*Individual { 
			return GenerationalReplacement(newIndividuals, oldIndividuals, eliteSize) 
		},
		func(individuals []*Individual, tournamentSize int) []*Individual { 
			return TournamentSelection(individuals, tournamentSize) 
		},
		fitnessFunc,		// Evaluation function
	)
	ctx.FatalIfErrorf(err, "failed to find the bestever individual")

	// Display the best individual found after evolution.
	fmt.Printf("Best individual: %s\n", bestEver)
}
