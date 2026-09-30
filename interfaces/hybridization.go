// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
package interfaces

// IHybridizationHook is the attachment point for an external module
// (e.g. evoCell). evoGo knows nothing about the module: it only calls 
// these two methods, if a hook is present.
//
// Contract (v1.0.0):
// - OnGenerationStart is called BEFORE genetic operators, with the
//   current population (post-selection), allowing the module to
//   seed/sync itself and, later, to enrich individuals via Deposit.
// - OnGenerationEnd is called AFTER evaluation, with the best
//   individual so far, allowing the module to harvest results.
//
// Nil-safety: implementers may receive an empty population at gen 0.
type IHybridizationHook interface {
    OnGenerationStart(generation int, population []IIndividual)
    OnGenerationEnd(generation int, bestEver IIndividual)
}