// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
    "evoGo/utils"
)

/* IndividualState helpers */

// RestoreIndividualState restores the individual's state from a backup.
func RestoreIndividualState(ind *Individual, state IndividualState) {
    ind.genome = state.genome
    ind.phenotype = state.phenotype
    ind.fitness = state.fitness
    ind.productionHistory = state.productionHistory
    ind.oldProductionFitness = state.oldProductionFitness
    ind.lastValidPhenotype = state.lastValidPhenotype
    ind.isExhausted = state.exhausted
    ind.dynamicRules = state.dynamicRules
    ind.dynamicRuleStack = state.dynamicRuleStack
}

// SaveIndividualState saves the current state of the individual.
func SaveIndividualState(ind *Individual) IndividualState {
    oldProductionFitness := make(map[string]float64, len(ind.oldProductionFitness))

    for k, v := range ind.oldProductionFitness {
        oldProductionFitness[k] = v
    }

    return IndividualState{
        genome:             ind.genome,
        phenotype:          ind.phenotype,
        fitness:            ind.fitness,
        productionHistory:  utils.DeepCopyProductionHistory(ind.productionHistory),
        oldProductionFitness: oldProductionFitness,
        lastValidPhenotype: ind.lastValidPhenotype,
        exhausted:          ind.isExhausted,
        dynamicRules:       utils.DeepCopyDynamicRules(ind.dynamicRules),
        dynamicRuleStack:   ind.dynamicRuleStack,  // A shallow copy is sufficient for []string
    }
}