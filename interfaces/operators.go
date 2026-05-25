// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

type (

    // ReplacementFunc is a function that replaces a population with a new one,
	// while retaining a certain number of elites.
    ReplacementFunc func(newIndividuals, oldIndividuals []IIndividual, eliteSize int) []IIndividual

    // SelectionFunc is a function that selects parents for breeding.
    SelectionFunc func(individuals []IIndividual, tournamentSize int) []IIndividual
    
)