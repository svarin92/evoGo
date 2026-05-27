// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package main

import (
	"evoGo/config"
	"evoGo/interfaces"
)

// Interfaces imported to ensure architectural consistency.
type (
	IIndividual           = interfaces.IIndividual
)

const (
	GENERATIONS             = config.GENERATIONS
	POPULATION_SIZE         = config.POPULATION_SIZE
)