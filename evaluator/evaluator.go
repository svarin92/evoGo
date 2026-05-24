// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package evaluator

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"

	"evoGo/config"
	"evoGo/model"
	"evoGo/utils"
)

/* Evaluator */

// Smart bridge and router between phenotypes and organisms.
type Evaluator struct {

	// Routing table: phenotype type → organism factory.	
	organismFactoryTable map[reflect.Type]OrganismFactory

	target 				 any
	templateCache 		 map[any]TemplateFunc

	// Routing table: target type → visitor.
	fitnessVisitorTable  map[reflect.Type]IFitnessVisitor
	templateVisitorTable map[reflect.Type]ITemplateVisitor

	fitnessFac           FitnessFactory
	templateFac          TemplateFactory
	numericMatchCase     INumericMatch
	numericTemplateCase  INumericTemplate
	stringMatchCase      IStringMatch
	stringTemplateCase   IStringTemplate
}

func (e *Evaluator) Create(target any) *Evaluator {
	e.target = target
	return e
}

func (e *Evaluator) CreateFitness(target any) *Evaluator {
	e.Create(target)
	e.fitnessFac = func() IFitnessMaker { return NewFitnessMaker() }
	e.fitnessVisitorTable = map[reflect.Type]IFitnessVisitor{}
	e.organismFactoryTable = map[reflect.Type]OrganismFactory{}
	return e
}

func (e *Evaluator) CreateTemplate(target any) *Evaluator {
	e.Create(target)
	e.templateCache = make(map[any]TemplateFunc)
	e.templateFac = func() ITemplateMaker { return NewTemplateMaker() }
	e.templateVisitorTable = map[reflect.Type]ITemplateVisitor{}
	return e
}

func (e *Evaluator) MakeFitness(individual IIndividual) {
	fitness := e.fitnessFac()
	e.numericMatchCase = fitness.MakeNumericMatchCase(e.HandleNumericMatchFunc())
	e.stringMatchCase = fitness.MakeStringMatchCase(e.HandleStringMatchFunc())

	// Registration of visitors.
	e.RegisterFitnessVisitor(reflect.TypeFor[float64](), e.numericMatchCase)
	e.RegisterFitnessVisitor(reflect.TypeFor[string](), e.stringMatchCase)

	// Registration of organism factories.
	e.RegisterOrganismFactory(reflect.TypeFor[float64](), func(phenotype any) IOrganism {

		// Update the individual's phenotype.
		individual.SetPhenotype(phenotype.(float64))

		// Create the organism of the existing individual.
		return model.NewNumeric(individual)
	})
	e.RegisterOrganismFactory(reflect.TypeFor[string](), func(phenotype any) IOrganism {

		// Update the individual's phenotype.
		individual.SetPhenotype(phenotype.(string))

		// Create the organism of the existing individual.
		return model.NewOntology(individual)
	})
}

func (e *Evaluator) MakeTemplate() {
	template := e.templateFac()
	e.numericTemplateCase = template.MakeNumericTemplateCase(e.HandleNumericTemplateFunc())
	e.stringTemplateCase = template.MakeStringTemplateCase(e.HandleStringTemplateFunc())

	// Registration of visitors.
	e.RegisterTemplateVisitor(reflect.TypeFor[float64](), e.numericTemplateCase)	
	e.RegisterTemplateVisitor(reflect.TypeFor[string](), e.stringTemplateCase)
}

func (e *Evaluator) ApplyNumericTemplate(phenotype float64, length int) float64 {
	phenotypeStr := fmt.Sprintf("%.0f", phenotype)

    // Adjust the length.
    if len(phenotypeStr) < length {

    	// Add zeros at the beginning.
        phenotypeStr = strings.Repeat("0", length-len(phenotypeStr)) + phenotypeStr
    } else if len(phenotypeStr) > length {
        
		// Truncate on the left.
        phenotypeStr = phenotypeStr[len(phenotypeStr)-length:]
    }

    // Convert to float64.
    corrected, err := strconv.ParseFloat(phenotypeStr, 64)
    
	if err != nil {
        return phenotype  // Return unchanged in case of error
    }

    return corrected
}

// ApplyStringTemplate corrects a string phenotype according to a template 
// "CVCVCV".
func (e *Evaluator) ApplyStringTemplate(phenotype, template, target string) string {
    corrected := []rune(phenotype)
    vowels := map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}

    for i, c := range template {

        if i >= len(corrected) || i >= len(target) {
            break
        }

        currentChar := corrected[i]
		targetChar := rune(target[i])

        if currentChar == targetChar {
            continue
        }

        isCurrentCharVowel := vowels[currentChar]
        
		switch c {
		case 'V':

            // If the template expects a vowel and the current letter is not a 
			// vowel, we replace it with the target letter.
            if !isCurrentCharVowel {
                corrected[i] = targetChar
            }

        case 'C':

            // If the template expects a consonant and the current letter is a 
			// vowel, we replace it with the target letter.
            if isCurrentCharVowel {
                corrected[i] = targetChar
            }

        }	

	}

    return string(corrected)
}

// DoApplyNumericTemplate applies a template to a numerical phenotype. It is 
// called after the template has been generated (via DoHandleNumericTemplate).
func (e *Evaluator) DoApplyNumericTemplate(model IOrganism) {

	// 1. Retrieve the phenotype and the template.
    phenotype, ok1 := utils.ToFloat64(model.GetPhenotype())   
	template, ok2 := model.GetTemplate().(int)

    if !ok1 || !ok2 {
        log.Printf(
			"Error: incompatible types (phenotype=%T, template=%T)\n",
            model.GetPhenotype(), model.GetTemplate(),
		)
        return
    }	
    
	// 2. Apply the template (adjust the length).
	corrected := e.ApplyNumericTemplate(phenotype, template)
	
	// 3. Update the phenotype.
    model.SetPhenotype(corrected)
}

// DoApplyStringTemplate applies a template to a string phenotype. It is 
// called after the template has been generated (via DoHandleStringTemplate).
func (e *Evaluator) DoApplyStringTemplate(model IOrganism) {

    // 1. Verify that the phenotype and the template are strings.
    phenotype, ok1 := model.GetPhenotype().(string)
    template, ok2 := model.GetTemplate().(string)

    if !ok1 || !ok2 {
        log.Printf(
			"Error: incompatible types (phenotype=%T, template=%T)\n",
			model.GetPhenotype(), model.GetTemplate(),
		)
        return
    }

	// 2. Apply the template.
    corrected := e.ApplyStringTemplate(phenotype, template, e.target.(string))

	// 3. Update the phenotype.
    model.SetPhenotype(corrected)
}

// DoHandleNumericTemplate generates a template for a numerical phenotype.
// It is called by HandleNumericTemplateFunc and uses the template visitor.
func (e *Evaluator) DoHandleNumericTemplate(model IOrganism) {

	// 1. Cast the target (e.target) to a numeric.
	target, ok1 := utils.ToFloat64(e.target)

	// 2. Retrieves the organism's phenotype (first typed as a string by the 
	//    constructor).
	candidate, ok2 := utils.ToFloat64(model.GetPhenotype)

	if !ok1 || !ok2 {
		log.Printf(
			"Error: incompatible types for numeric template (target=%T, phenotype=%T)\n",
			target, candidate,
		)
		return
	}

	// 3. Genrate template.
	template := e.GenerateNumericTemplate(candidate, target)

	// 4. Updates the organism template before correcting the phenotype.
	model.SetTemplate(template)
}

func (e *Evaluator) DoHandleStringTemplate(model IOrganism) {

	// 1. Cast the target (e.target) to a string.
	target, ok1 := e.target.(string)

	// 2. Retrieves the organism's template (already typed as a string by the 
	// constructor).
	candidate, ok2 := model.GetPhenotype().(string)

	if !ok1 || !ok2 {
        log.Printf(
			"Error: incompatible types for template string (target=%T, phenotype=%T)\n",
            e.target, model.GetPhenotype(),
		)
        return
    }

	// 3. Generate template (a string "CVCVCV").
	template := e.GenerateStringTemplate(candidate, target)

	// 4. Updates the organism template before correcting the phenotype.
	model.SetTemplate(template)
}

// Evaluates the fitness of an organism's phenotype based on its similarity to
// a target number.
func (e *Evaluator) DoHandleNumericMatch(model IOrganism) {

	// 1. Cast the target (e.target) to a numeric.
	target, ok := e.target.(float64)

	if !ok {
		model.SetFitness(0)  // Invalid target type
		return
	}

	// 2. Retrieves the organism's phenotype (first typed as a string by the 
	//    constructor).
	candidate := model.GetPhenotype().(float64)
	usedWraps := model.GetUsedWraps()

	// 3. Calculate fitness.
	fitness := e.DoNumericMatch(candidate, target, usedWraps)

	// 4. Updates the organism's fitness to the calculated value.
	model.SetFitness(fitness)
}

// Evaluates the fitness of an organism's phenotype based on its similarity 
// to a target string.
func (e *Evaluator) DoHandleStringMatch(model IOrganism) {

	// 1. Cast the target (e.target) to a string
	target, ok := e.target.(string)

	if !ok {
		model.SetFitness(0) // Invalid target type
		return
	}

	// 2. Retrieves the organism's phenotype (already typed as a string by 
	//    the constructor).
	candidate := model.GetPhenotype().(string)
	usedWraps := model.GetUsedWraps()

	// 3. Calculate fitness.
	fitness := e.DoStringMatch(candidate, target, usedWraps)

	// 4. Updates the organism's fitness to the calculated value.
	model.SetFitness(fitness)
}

// Routing logic and dynamic construction of the organism.
func (e *Evaluator) DoEvaluate(ind IIndividual) float64 {

	// 1. Retrieve the phenotype from the individual.
	phenotype := ind.GetPhenotype()

	// 2. Determines the type of the phenotype.
	phenotypeType := reflect.TypeOf(phenotype)

	// 3. Retrieve the appropriate organism constructor.
	organismFactory, ok := e.organismFactoryTable[phenotypeType]

	if !ok {
		return 0  // Unsupported phenotype type
	}

	// 4. Builds the organism.
	organism := organismFactory(phenotype)

	// 5. Store the organism within the individual.
    ind.SetOrganism(organism)

	// 6. Determine the target type (e.target).
	targetType := reflect.TypeOf(e.target)

	// 7. Retrieves the appropriate visitor for the target.
	visitor, ok := e.fitnessVisitorTable[targetType]

	if !ok {
		return 0  // Unsupported target type
	}

	// 8. Call organism.Accept(visitor) to trigger the visit. The generic
	//    visitor is called with a callback that provides the organism at
	//    the time of the call. This avoids static typing issues and ensures
	//    that the organism is of the expected type (IOrganism) at runtime.
	organism.Accept(visitor, func() IVisited {
		return organism
	})

	// 9. Update the individual's fitness
	ind.SetFitness(organism.GetFitness())
	
	// 10. Returns the fitness calculated by the visitor.
	return ind.GetFitness()
}

func (e *Evaluator) DoFormat(ind IIndividual) bool {

    // 1. Retrieve the organism from the individual (already created in DoEvaluate).
    organism := ind.GetOrganism()

    if organism == nil {
        log.Printf("no organism associated with the individual\n")
        return false
    }

    // 2. Retrieve the template visitor.
    phenotypeType := reflect.TypeOf(ind.GetPhenotype())

    visitor, ok := e.templateVisitorTable[phenotypeType]

	if !ok {
        log.Printf("no template visitor for type %v\n", phenotypeType)
        return false
    }

    // 3. Generate and apply the template via Accept.
    organism.Accept(visitor, func() IVisited { return organism })

	// 4. Update the individual's phenotype (via the organism).
	ind.SetPhenotype(organism.GetPhenotype())

    // 5. Return true to indicate that the correction has been applied (but not yet evaluated).
    return true
}

func (e *Evaluator) DoNumericMatch(candidate, target float64, usedWraps int) float64 {

	// Special case: if the candidate is equal to the target, return 1.0 immediately.
	if candidate == target {
		return 1.0
	}

	// 1. Calculate the absolute distance between the target and the candidate
	//    (equivalent to the Levenshtein distance for strings).
	distance := math.Abs(target - candidate)

	// 2. Normalize the distance to obtain a fitness score. The maximum 
	//    possible distance depends on the scale of the numbers. Here, 
	//    we use the sum of the absolute values ​​as a reference.
	maxDistance := math.Abs(target) + math.Abs(candidate)

	if maxDistance == 0 {
		return 0.0  // Avoid division by zero
	}

	// Normalize the distance to [0, 1].
	normalizedDistance := distance / maxDistance
	fitness := 1.0 - normalizedDistance

	// 3. Bonus for correctly placed numbers (equivalent to correctly placed 
	//    letters). For numbers, a bonus can be added if the first digits 
	//    match.
	candidateStr := fmt.Sprintf("%.6f", candidate)  // 6 decimal places to avoid precision artifacts
	targetStr := fmt.Sprintf("%.6f", target)

	positionBonus := 0.0
	minLength := min(len(candidateStr), len(targetStr))

	for i := 0; i < minLength; i++ {

		if candidateStr[i] == targetStr[i] {
			positionBonus += 0.05  // Bonus for each matching number
		}

	}

	// Apply the bonus.
	fitness += positionBonus

	// Limit fitness to 0.99 if it's not a perfect match.
	if fitness >= 1.0 && candidate != target {
		fitness = 0.99
	}

	// 4. Linear penalty for wraps (as for strings).
	wrapPenalty := config.WRAP_PENALTY_FACTOR * float64(usedWraps)
	fitness = math.Max(0, fitness-wrapPenalty)  // Fitness normalized between 0 and 1

	return fitness
}

func (e *Evaluator) DoStringMatch(candidate, target string, usedWraps int) float64 {

	// Special case: if the candidate is equal to the target, return 1.0 
	// immediately.
	if candidate == target {
		return 1.0
	}

	// 1. Calculate the Levenshtein distance. Favors phenotypes close to 
	//    "golden" even if they are not perfect.
	distance := utils.LevenshteinDistance(candidate, target)

	targetLength := len(target)
	candidateLength := len(candidate)

	// 2. Penalize length deviations heavily. Ensures that phenotypes are the 
	//    correct length.
	lengthDifference := int(math.Abs(float64(targetLength - candidateLength)))
	distance += lengthDifference * 2  // Additional weight for length

	// 3. Normalize the distance to obtain a fitness score. The shorter the 
	//    distance, the higher the fitness.
	maxLength := max(candidateLength, targetLength)

	if maxLength == 0 {
		return 0.0
	}

	// Normalize the distance to [0, 1].
	fitness := 1.0 - (float64(distance) / float64(maxLength))

	// 4. Bonus for well-placed letters.
	positionBonus := 0.0

	for i := 0; i < min(len(candidate), len(target)); i++ {

		if candidate[i] == target[i] {
			positionBonus += 0.05  // Increased bonus for well-placed letters
		}

	}

	// Apply the bonus.
	fitness += positionBonus

	// Limit fitness to 0.99 if not an exact match.
	if fitness >= 1.0 && candidate != target {
		fitness = 0.99
	}

	// 5. Linear penalty for wraps.
	wrapPenalty := config.WRAP_PENALTY_FACTOR * float64(usedWraps)

	// Fitness normalized between 0 and 1 (1 = perfect match).
	fitness = max(0, fitness-wrapPenalty)

	return fitness
}

func (e *Evaluator) Evaluate(individual IIndividual) float64 {

	// Implementation of certain internal data methods or algorithms used to
	// evaluate fitness.
	e.MakeFitness(individual)

	return e.DoEvaluate(individual)
}

func (e *Evaluator) Format(individual IIndividual) bool {

	// Implementation of certain internal data methods or algorithms used to
	// evaluate templating.
	e.MakeTemplate()

	return e.DoFormat(individual)
}

// GenerateNumericTemplate generates a template (int) for a numerical phenotype.
// Returns an int (maximum length between the target and the phenotype).
func (e *Evaluator) GenerateNumericTemplate(phenotype float64, target float64) int {

	// Generating the numeric template.
	targetStr := fmt.Sprintf("%.0f", target)
	phenotypeStr := fmt.Sprintf("%.0f", phenotype)
	return max(len(targetStr), len(phenotypeStr))  // example: 5 for "12345"

}

func (e *Evaluator) GenerateStringTemplate(phenotype string, target string) string {
	var template strings.Builder
    vowels := map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}

    // Use the target to generate the template (e.g., "golden" → "CVCCVC").
    for _, c := range target {

		if c == '_' {
            template.WriteString("_")
        } else if vowels[c] {
            template.WriteString("V")
        } else {
            template.WriteString("C")
        }

    }
	
    return template.String()  // Example: "CVCCVC" for "golden"
}

func (e *Evaluator) HandleNumericMatchFunc() func(any) {
	return func(data any) {

		model, ok := data.(IOrganism)

		if !ok {
			fmt.Println("Error: model is not an Organism")
			return
		}

		switch model.GetPhenotype().(type) {
		case int, int32, int64, uint, uint32, uint64, float32, float64:
			e.DoHandleNumericMatch(model)
		default:
			fmt.Printf("Error: phenotype is not a numeric (type: %T)\n", model.GetPhenotype())
			return
		}

	}
}

func (e *Evaluator) HandleNumericTemplateFunc() func(any) {
	return func(data any) {

		model, ok := data.(IOrganism)

		if !ok {
			fmt.Println("Error: model is not an Organism")
			return
		}

		// Verify that the phenotype is numerical.
		switch model.GetPhenotype().(type) {
		case int, int32, int64, uint, uint32, uint64, float32, float64:

			// 1. Generate the template.
			e.DoHandleNumericTemplate(model)

			// 2. Apply the template.
			e.DoApplyNumericTemplate(model)
		default:
			fmt.Printf("Error: phenotype is not a numeric (type: %T)\n", model.GetPhenotype())
			return
		}

	}
}

func (e *Evaluator) HandleStringMatchFunc() func(any) {
	return func(data any) {

		model, ok := data.(IOrganism)

		if !ok {
			fmt.Println("Error: model is not an Organism")
			return
		}

		phenotype := model.GetPhenotype()
		_, ok = phenotype.(string)

		if !ok {
			fmt.Println("Error: phenotype is not an string")
			return
		}

		e.DoHandleStringMatch(model)
	}
}

func (e *Evaluator) HandleStringTemplateFunc() func(any) {
	return func(data any) {

		model, ok := data.(IOrganism)

		if !ok {
			fmt.Println("Error: model is not an Organism")
			return
		}

		phenotype := model.GetPhenotype()
		_, ok = phenotype.(string)

		if !ok {
			fmt.Println("Error: phenotype is not an string")
			return
		}

		// 1. Generate the template.
		e.DoHandleStringTemplate(model)

		// -- Debug --
		// log.Printf("Handle: After generate string template phenotype: %v", model.GetPhenotype())

		// 2. Apply the template.
		e.DoApplyStringTemplate(model)

		// -- Debug --
		// log.Printf("Handle: After apply string template phenotype: %v", model.GetPhenotype())
	}
}

// Register an organism constructor for a given phenotype type.
func (e *Evaluator) RegisterOrganismFactory(
	phenotypeType reflect.Type, 
	factory OrganismFactory,
) {
	e.organismFactoryTable[phenotypeType] = factory
}

// Registers a fitness function (visitor) for a given target type.
func (e *Evaluator) RegisterFitnessVisitor(
	targetType reflect.Type, 
	fitnessFunc IFitnessVisitor,
) {
	e.fitnessVisitorTable[targetType] = fitnessFunc
}

// Registers a template function (visitor) for a given phenotype type.
func (e *Evaluator) RegisterTemplateVisitor(
	phenotypeType reflect.Type, 
	templateFunc ITemplateVisitor,
) {
	e.templateVisitorTable[phenotypeType] = templateFunc
}