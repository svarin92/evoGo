// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package config

const (
	CODONS_SIZE             int     = 127
	CROSSOVER_PROBABILITY   float64 = 0.7
	ELITE_SIZE              int     = 1
	GENERATIONS             int     = 30
	GENERATION_SIZE         int     = 50

	// MAX_WRAPS defines the maximum number of times the genome can be reused.
	// A value of 100 allows for good diversity without the risk of infinite loops.
	// Note: A value of 0 disables bouncing (simple initial behavior).
	MAX_WRAPS 			    int 	= 0

	// Adjust the penalty factor. Too high a factor penalizes partial solutions too
	// heavily. A factor that is too low does not sufficiently reduce wraps.
	WRAP_PENALTY_FACTOR     float64 = 0.01

	MUTATION_PROBABILITY    float64 = 0.01
	POPULATION_SIZE         int     = 50

	EXPLORATION_PROBABILITY float64 = 0.2
)
