// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

// SuccessfulProduction represents a successful production with its fitness 
// and frequency.
type SuccessfulProduction struct {
	Production []IRuleModel  // Successful production
	Fitness    float64       // Fitness associated
	Frequency  int           // Number of occurrences
}

func (sp SuccessfulProduction) GetProduction() []IRuleModel {
    return sp.Production
}

func (sp SuccessfulProduction) GetFitness() float64 {
    return sp.Fitness
}

func (sp SuccessfulProduction) GetFrequency() int {
    return sp.Frequency
}
