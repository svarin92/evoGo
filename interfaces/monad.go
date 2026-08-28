// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

// IIndividual defines the contract for an individual in an evolving 
// algorithm.
type IIndividual interface {
	
	// Getters to access the individual's properties.
	GetDynamicRules() map[string]IRuleModel
	GetDynamicRuleStack() []string
	GetFitness() float64
	GetGenome() []int
	GetLastValidPhenotype() any
	GetOldProductionFitness(key string) (float64, bool)	
	GetOrganism() IOrganism
	GetPhenotype() any
    GetProductionHistory() [][]IRuleModel
    GetUsedCodons() int

	// Setters to modify properties (if necessary).
	SetDynamicRules(map[string]IRuleModel)
	SetDynamicRuleStack([]string)
    SetFitness(float64)
	SetLastValidPhenotype(phenotype any)
    SetOrganism(IOrganism)
	SetPhenotype(any)	
	
	ClearDynamicRuleStack()
	Copy() IIndividual  // Copy creates a deep copy of the individual
	Evaluate(fitness FitnessFunc) error
    GeneratePhenotype(genomizer IGenomizer) error
	MutateCodon(index int, newValue int)
    Repair(immuneSys IImmune) error
	UpdateGenomeSegment(segment []int, offset int)
}