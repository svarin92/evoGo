// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package evaluator

/* Exports */

func NewFitnessMaker() IFitnessMaker {
	return new(FitnessMaker).Create()
}

func NewFitnessEvaluator(target any) IEvaluable {
	return new(Evaluator).CreateFitness(target)
}

func NewFitness(target any) FitnessFunc {
	fitness := func() IEvaluable { return NewFitnessEvaluator(target) }
	return func(individual IIndividual) float64 {
		evaluator := fitness()
		return evaluator.Evaluate(individual)
	}
}

func NewTemplateMaker() ITemplateMaker {
	return new(TemplateMaker).Create()
}

func NewTemplateEvaluator(target any) IEvaluable {
	return new(Evaluator).CreateTemplate(target)
}

func NewTemplate(target any) TemplateFunc {
	template := func() IEvaluable { return NewTemplateEvaluator(target) }
	return func(individual IIndividual) bool {

		// -- Debug --
		// log.Printf("templateFunction: BEFORE Format: individual phenotype=%q\n", individual.GetPhenotype())

		evaluator := template()
		result := evaluator.Format(individual)

		// -- Debug --
		// log.Printf("templateFunction: AFTER Format: individual phenotype=%q, result=%v\n", 
		// 	individual.GetPhenotype(), 
		// 	result,
		// )

		return result
	}
}