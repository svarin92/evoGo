// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package renderer

import (
	"evoGo/interfaces"
)

// Aliases for functional types.
type (
	GrammarProviderFactory     = interfaces.GrammarProviderFactory
)

// Interfaces imported to ensure architectural consistency.
type (
	IGrammarProvider           = interfaces.IGrammarProvider
	IProductionHistoryProvider = interfaces.IProductionHistoryProvider
    IPopulationStatsProvider   = interfaces.IPopulationStatsProvider
 
	IRenderer                  = interfaces.IRenderer

	IRuleModel                 = interfaces.IRuleModel
)