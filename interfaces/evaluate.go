// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

import (
	"reflect"
)

type IEvaluable interface {
	Evaluate(individual IIndividual) float64
	Format(individual IIndividual) bool
	RegisterFitnessVisitor(targetType reflect.Type, fitnessFunc IFitnessVisitor)
	RegisterTemplateVisitor(targetType reflect.Type, templateFunc ITemplateVisitor)
	RegisterOrganismFactory(phenotypeType reflect.Type, factory OrganismFactory)
}

type EvaluatorFactory func() IEvaluable