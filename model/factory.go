// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

/* Exports */

// NewIndividual is a factory function to create a new individual.
func NewIndividual(genome []int) *Individual {
	return new(Individual).Create(genome)
}

// NewNumeric is a factory that checks the type and initializes a Numeric.
// It accepts an Individual (for the public API) but checks that it is a 
// *model.Individual.
func NewNumeric(individual IIndividual) *Numeric {
    concreteInd, ok := individual.(*Individual)

    if !ok {
        return nil
    }

    return new(Numeric).Create(concreteInd)
}

// NewOntology is a factory that checks the type and initializes an Ontology.
// It accepts an Individual (for the public API) but checks that it is a 
// *model.Individual.
func NewOntology(individual IIndividual) *Ontology {
    concreteInd, ok := individual.(*Individual)

    if !ok {
        return nil
    }

    return new(Ontology).Create(concreteInd)
}

// NewRuleModel creates a new RuleModel instance. Function exported to allow 
// creation from other packages.
func NewRuleModel(symbol string, symbolType SymbolType, rhs [][]IRuleModel) IRuleModel {
    return &RuleModel{
        Symbol:     symbol,
        SymbolType: symbolType,
        rhs:        rhs,  // the private rhs field is exported
    }
}

func NewRuleModelWithCount(symbol string, symbolType SymbolType, count int) IRuleModel {
    return &RuleModel{
        Symbol:  symbol,
        SymbolType: symbolType,
        count:   count,
    }
}