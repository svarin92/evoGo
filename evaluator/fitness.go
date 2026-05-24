// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package evaluator

/* NumericMatch */

type NumericMatch struct {
	ModelVisitor
}

/* StringMatch */

type StringMatch struct {
		ModelVisitor
}

/* FitnessFactory */

type FitnessMaker struct{

}

func (fm *FitnessMaker) Create() *FitnessMaker {
	return fm
}

func (fm *FitnessMaker) MakeNumericMatchCase(vf VisitorFunc) INumericMatch {
	return new(NumericMatch).Create(vf)
}

func (fm *FitnessMaker) MakeStringMatchCase(vf VisitorFunc) IStringMatch {
	return new(StringMatch).Create(vf)
}