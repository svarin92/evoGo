// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

// IRenderer defines the contract for rendering.
type IRenderer = interface {
	DisplayIndividualDetails(provider IProductionHistoryProvider, generation int)
	ExportToDOT(provider IProductionHistoryProvider, filename string) error
	PrintGenomeCorrection(
	provider IProductionHistoryProvider,
	fitnessThreshold float64, 
	afterCorrection bool,
    averageFitness float64,
	)
	PrintGrammaticalDerivation(provider IProductionHistoryProvider)
	PrintOriginalRules()
	PrintStats(provider IPopulationStatsProvider, generation int, generationSize int)
	PrintSuccessfulProductions()
}

// RendererFactory is a function that provides an IRenderer.
type RendererFactory func() IRenderer