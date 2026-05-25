// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

type IOrganism interface {
	INotifiedModel
	GetFitness() float64
	GetPhenotype() any  // interface{}
	GetTemplate() any
	GetUsedWraps() int
	SetFitness(value float64)
	SetPhenotype(value any)
	SetTemplate(value any)
}

// Function to build an organism adapted to a phenotype.
type OrganismFactory func(phenotype any) IOrganism