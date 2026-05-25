// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

// IIndividual defines the contract for an individual in an evolving 
// algorithm.
type IIndividual interface {
	
	// Getters to access the individual's properties.
	GetFitness() float64
	GetGenome() []int
	GetOldProductionFitness(key string) (float64, bool)	
	GetOrganism() IOrganism
	GetPhenotype() any
    GetProductionHistory() [][]IRuleModel
    GetUsedCodons() int

	// Setters to modify properties (if necessary).
    SetFitness(float64)
    SetOrganism(IOrganism)
	SetPhenotype(any)	
	
	// Copy creates a deep copy of the individual.
	Copy() IIndividual
	Evaluate(fitness FitnessFunc) error
    GeneratePhenotype(genomizer IGenomizer) error
	MutateCodon(index int, newValue int)
    Repair(immuneSys IImmune) error
	UpdateGenomeSegment(segment []int, offset int)
}