package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"evoGo/evaluator"
	"evoGo/ge"
	"evoGo/grammar"
	"evoGo/operators"
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
	grammar, err := grammar.NewSpeciesGrammar(ctx, r)
	ctx.FatalIfErrorf(err, "failed to create species grammar")

	// Target for evolution (e.g., the word "golden").
	target := "golden"

	// Create an evaluation function that measures the proximity of individuals
	// to the target.
	fitnessFunc := evaluator.NewFitness(target)

	// replacementFunc defines the generational replacement strategy. It 
	// takes the following input:
	// - newIndividuals: the newly generated individuals (children after 
	//   crossover/mutation).
	// - oldIndividuals: the current population before replacement.
	// - eliteSize: the number of elite individuals to keep from the old 
	//   population.
	// It returns a new slice of individuals representing the updated 
	// population, by combining the elites from the old population and the 
	// new individuals. Here, we use operators.GenerationalReplacement to 
	// implement a classic generational replacement.
	// Note: This function must preserve the population size: 
	//   (len(newIndividuals) == len(oldIndividuals)).
	replacementFunc := func(newIndividuals, oldIndividuals []IIndividual, eliteSize int) []IIndividual { 
		return operators.GenerationalReplacement(newIndividuals, oldIndividuals, eliteSize) 
	}

	// selectionFunc defines the strategy for selecting parents for 
	// reproduction. It takes the following input:
	// - individuals: the current population from which to select parents.
	// - tournamentSize: the tournament size for selection (e.g., 3 for a 
	//   tournament between 3 individuals).
	// It returns a slice of individuals selected as parents for the next 
	// generation. Here, we use operators.TournamentSelection to implement 
	// tournament selection, which favors the fittest individuals while 
	// maintaining some diversity.
	// Note: The size of the returned slice depends on the implementation 
	// (generally equal to len(individuals)).
	selectionFunc := func(individuals []IIndividual, tournamentSize int) []IIndividual { 
		return operators.TournamentSelection(individuals, tournamentSize) 
	}

	// Grammatical Evolution (GE) - Execute the evolution algorithm over a 
	// given number of generations. Returns the best individual found after 
	// evolution, or an error.
	bestEver, err := ge.SearchLoop(
		GENERATIONS,        // Total number of generations
		POPULATION_SIZE,    // Population size
		grammar,            // Grammar used to generate individuals
		target,	            // Target to be reached
		replacementFunc,    // Generational replacement function (here, elitist replacement)
		selectionFunc,      // Parent selection function (here, tournament selection)
		fitnessFunc,		// Evaluation function
	)
	ctx.FatalIfErrorf(err, "failed to find the bestever individual")

	// Display the best individual found after evolution.
	fmt.Printf("Best individual: %s\n", bestEver)
}
