// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package operators

import (
	"evoGo/config"
	"evoGo/interfaces"
)

// Interfaces imported to ensure architectural consistency.
type (
	IIndividual = interfaces.IIndividual
)

const (
	CODONS_SIZE           = config.CODONS_SIZE
	CROSSOVER_PROBABILITY = config.CROSSOVER_PROBABILITY
)
