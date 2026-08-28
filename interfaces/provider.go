// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

// ISuccessfulProduction defines the contract for a successful production.
//
// A successful production includes:
//   - The production itself (sequence of rules).
//   - Its fitness score.
//   - Its frequency of occurrence.
type ISuccessfulProduction interface {
    GetProduction() []IRuleModel  // Returns the output (e.g., []IRuleModel)
    GetFitness()    float64       // Return the associated fitness
    GetFrequency()  int           // Returns the frequency of occurrence
}

// IGrammarProvider provides the grammar data needed for rendering or other 
// operations. Implemented by: Genomizer.
type IGrammarProvider interface {
    FindSimilarProductions(production []IRuleModel, averageFitness float64) [][]IRuleModel
    GetSymbols() map[string]IRuleModel
    GetAverageFitness(production []IRuleModel) float64
    GetSuccessfulProductions() []ISuccessfulProduction
    ProductionSimilarity(p1, p2 []IRuleModel) float64
}

// IProductionHistoryProvider provides an individual's data for rendering or 
// analysis. Implemented by: Individual.
type IProductionHistoryProvider interface {
    GetProductionHistory() [][]IRuleModel
    GetPhenotype() any
    GetFitness() float64
    GetGenome() []int
    GetUsedCodons() int
    GetOldProductionFitness(key string) (float64, bool)
}

// To access population statistics.
type IPopulationStatsProvider interface {
    GetIndividuals() []IIndividual
    // GetGeneration() int
}

// GrammarProviderFactory is a function that provides an IGrammarProvider.
type GrammarProviderFactory func() IGrammarProvider