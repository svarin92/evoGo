// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

type IFitnessVisitor interface {
	IVisitor
}

type IFitnessMaker interface {
	MakeNumericMatchCase(vf VisitorFunc) INumericMatch
	MakeStringMatchCase(vf VisitorFunc) IStringMatch
}

// Generic fitness function (public interface).
type FitnessFunc func(IIndividual) float64

type FitnessFactory func() IFitnessMaker