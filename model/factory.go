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

// NewRuleModel creates a new rule with an optional DerivationType. If 
// DerivationType is not provided, HomogeneousDerivation is used by default.
func NewRuleModel(
    symbol string, 
    symbolType SymbolType, 
    rhs [][]IRuleModel,
    derivationType ...DerivationType,  // Variadic parameter to make it optional
) IRuleModel {
    ruleModel := &RuleModel{
        Symbol:     symbol,
        SymbolType: symbolType,
        rhs:        rhs,  // the private rhs field is exported
        count:      1,    // Default value
    }

    if len(derivationType) > 0 {
        ruleModel.derivationType = derivationType[0]
    } else {
        ruleModel.derivationType = HomogeneousDerivation  // By default
    }

    return ruleModel
}

func NewRuleModelWithCount(symbol string, symbolType SymbolType, count int) IRuleModel {
    return &RuleModel{
        Symbol:  symbol,
        SymbolType: symbolType,
        count:   count,
    }
}