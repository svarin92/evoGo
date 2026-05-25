// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import "log"

/* Numeric */

type Numeric struct {
	NotifiedModel
	*Individual
}

func (num *Numeric) Create(individual *Individual) *Numeric {
	num.Individual = individual
	return num
}

func (num *Numeric) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case INumericMatch, INumericTemplate:
		v.Visit(num)
	default:
		num.VisitedModel.DoAccept(v)
	}

}

func (num *Numeric) GetFitness() float64 {
	return num.fitness
}

func (num *Numeric) GetPhenotype() any {
	return num.phenotype
}

func (num *Numeric) GetTemplate() any {
	return num.template
}

func (num *Numeric) GetUsedWraps() int {
	return num.usedWraps
}

func (num *Numeric) SetFitness(value float64) {
	num.fitness = value
}

func (num *Numeric) SetPhenotype(value any) {
	num.phenotype = value
}

func (num *Numeric) SetTemplate(value any) {
	num.template = value
}

/* Ontology */

type Ontology struct {
	NotifiedModel
	*Individual
}

func (onto *Ontology) Create(individual *Individual) *Ontology {
	onto.Individual = individual
	return onto
}

func (onto *Ontology) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case IStringMatch, IStringTemplate:
		v.Visit(onto)
	default:
		onto.VisitedModel.DoAccept(v)
	}

}

func (onto *Ontology) GetFitness() float64 {
	return onto.fitness
}

func (onto *Ontology) GetPhenotype() any {
	return onto.phenotype
}

func (onto *Ontology) GetTemplate() any {
	return onto.template
}

func (onto *Ontology) GetUsedWraps() int {
	return onto.usedWraps
}

func (onto *Ontology) SetFitness(value float64) {
	onto.fitness = value
}

func (onto *Ontology) SetPhenotype(value any) {
	
	if s, ok := value.(string); !ok {
		log.Printf("Warning: expected string phenotype, got %T", s)
	}

	onto.phenotype = value
}

func (onto *Ontology) SetTemplate(value any) {
	onto.template = value
}