// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"evoGo/utils"
	"fmt"
	"log"
	"maps"
	"strings"
)

/* IndividualState */

// IndividualState represents a complete snapshot of an Individual's state.
// This type is private to the model package to encapsulate internal state.
type IndividualState struct {
    genome               []int
    phenotype            any
    fitness              float64
    productionHistory    [][]IRuleModel
    oldProductionFitness map[string]float64
    lastValidPhenotype   any
    exhausted            bool
    dynamicRules         map[string]IRuleModel
    dynamicRuleStack     []string
}

func (state *IndividualState) GetFitness() float64 {
	return state.fitness
}

/* Individual */

// Individual represents a cellular agent with its own genome and phenotype.
// Like a cell, it can mutate, cross with other individuals, and activate 
// repair mechanisms.
type Individual struct {
	fitness              float64  // Fitness value
	
	// dynamicRules represents the individual's library of non-coding 
	// RNAs (ncRNAs). Each ncRNA (a key in the map) is a dynamic rule 
	// that regulates phenotype generation, just as miRNAs regulate 
	// gene expression by targeting specific sequences.
	dynamicRules         map[string]IRuleModel

	// dynamicRuleStack represents the stack of active non-coding RNAs 
	// (transient ncRNAs). Like snoRNAs or siRNAs, these ncRNAs are 
	// transient and used to regulate phenotype generation at a given 
	// moment. They follow a LIFO logic (last added, first used).
    dynamicRuleStack     []string

	// genome represents the individual's DNA sequence, like a cell's 
	// genome.
	genome               []int    // Genome of the individual, represented as an array of integers
	
	// phenotype represents the phenotype generated from the genome, 
	// such as a protein synthesized from DNA.
	phenotype            any      // Generic phenotype 
	
	template             any
	organism             IOrganism
	usedCodons           int      // Number of used codons by the individual
	usedWraps            int
	compiledPhenotype    any      // Compiled phenotype if applicable
	
	// productionHistory represents the history of productions used to 
	// generate the phenotype.
	productionHistory    [][]IRuleModel
	
	oldProductionFitness map[string]float64

	// LastValidPhenotype represents the individual's "immune memory." 
	// Like a memory B cell, it stores the last valid phenotype
	// to enable rapid reactivation in the event of a reduction failure.
	lastValidPhenotype   any     // immune memory

	// isExhausted marks the individual as "exhausted" (like a T-cell in 
	// chronic infection). Once exhausted, the individual skips grammatical 
	// reduction attempts (CorrectByGrammaticalPaths) but can still be 
	// corrected by other mechanisms (CorrectByTemplate, CorrectByGenome).
	isExhausted bool
}

// Create initializes a new individual with a given genome.
func (ind *Individual) Create(genome []int) *Individual {

	if genome == nil {
		ind.genome = []int{}
	} else {
		ind.genome = genome
	}

	ind.fitness = 0
	ind.dynamicRules = make(map[string]IRuleModel)
    ind.dynamicRuleStack = []string{}
	ind.phenotype = ""
	ind.template = ""
	ind.usedCodons = 0
	ind.compiledPhenotype = nil
	ind.productionHistory = [][]IRuleModel{}
	ind.oldProductionFitness = make(map[string]float64)
	return ind
}

// ClearDynamicRuleStack clears the ncRNA stack (simulates biological 
// degradation).
func (ind *Individual) ClearDynamicRuleStack() {

	// ncRNAs are degraded in the cell cytoplasm (Individual) after regulating 
	// gene expression: ncRNAs are stored within the Individual (specifically, 
	// in the cytoplasm). They are degraded (cleared) after use by the 
	// Individual itself.
    ind.dynamicRuleStack = []string{}  // Degradation of ncRNAs
}

// Copy creates a deep copy of an individual.
func (ind *Individual) Copy() IIndividual { 

	// Create a new individual with the same parameters as the original.
	newInd := NewIndividual(nil)

    // Copy the genome.
    genome := make([]int, len(ind.GetGenome()))
    copy(genome, ind.GetGenome())
    newInd.SetGenome(genome)

	// Copy simple fields (primitive or immutable).
    newInd.SetFitness(ind.GetFitness())
    newInd.SetPhenotype(ind.GetPhenotype())
    newInd.SetUsedCodons(ind.GetUsedCodons())
    newInd.SetTemplate(ind.GetTemplate())

    // Copy the production history (deep copy).
	newInd.SetProductionHistory(utils.DeepCopyProductionHistory(ind.GetProductionHistory()))

	// Copy dynamicRules (persistent non-coding RNAs).
    dynamicRules := make(map[string]IRuleModel)
	maps.Copy(dynamicRules, ind.GetDynamicRules())
    newInd.SetDynamicRules(dynamicRules)

	// Copier dynamicRuleStack (ARNnc transitoires).
    dynamicRuleStack := make([]string, len(ind.GetDynamicRuleStack()))
    copy(dynamicRuleStack, ind.GetDynamicRuleStack())
    newInd.SetDynamicRuleStack(dynamicRuleStack)
	
	return newInd
}

// Evaluate assesses the individual's phenotype and updates their fitness.
func (ind *Individual) Evaluate(fitness FitnessFunc) error {
	ind.fitness = fitness(ind)
	return nil
}

// GetDynamicRules returns the individual's dynamic rules (ncRNAs).
func (ind *Individual) GetDynamicRules() map[string]IRuleModel {

    if ind.dynamicRules == nil {
        return make(map[string]IRuleModel)  // Avoid returning nil
    }
    
	// Returns a copy to prevent external modifications.
    dynamicRulesCopy := make(map[string]IRuleModel, len(ind.dynamicRules))
    
	// for k, v := range ind.dynamicRules {
    //     dynamicRulesCopy[k] = v
    // }
	maps.Copy(dynamicRulesCopy, ind.dynamicRules)
    
	return dynamicRulesCopy
}

// GetDynamicRuleStack returns the stack of the individual's active 
// non-coding RNAs.
func (ind *Individual) GetDynamicRuleStack() []string {

	if ind.dynamicRuleStack == nil {
        return []string{}  // Avoid returning nil
    }

	// Returns a copy to avoid external modifications.
    stackCopy := make([]string, len(ind.dynamicRuleStack))
    copy(stackCopy, ind.dynamicRuleStack)
    return stackCopy
}

func (ind *Individual) GetFitness() float64 {
	return ind.fitness
}

func (ind *Individual) GetGenome() []int {
    return ind.genome
}

func (ind *Individual) GetExhausted() bool {
    return ind.isExhausted
}

// Getter for LastValidPhenotype.
func (ind *Individual) GetLastValidPhenotype() any {
    return ind.lastValidPhenotype
}

func (ind *Individual) GetOldProductionFitness(key string) (float64, bool) {
    oldFitness, exists := ind.oldProductionFitness[key]
    return oldFitness, exists
}

// GetOrganism returns the organism associated with the individual.
func (ind *Individual) GetOrganism() IOrganism {
	return ind.organism
}

func (ind *Individual) GetPhenotype() any {
    return ind.phenotype
}

// GetProductionHistory returns a deep copy of the production history to 
// ensure data isolation between individuals.
func (ind *Individual) GetProductionHistory() [][]IRuleModel {
	return utils.DeepCopyProductionHistory(ind.productionHistory)
}

// Getter pour template.
func (ind *Individual) GetTemplate() any {
    return ind.template
}

func (ind *Individual) GetUsedCodons() int {
    return ind.usedCodons
}

// GeneratePhenotype generates the individual's phenotype using a Genomizer.
func (ind *Individual) GeneratePhenotype(genomizer IGenomizer) error {

	// -- Debug --
	// log.Printf("GeneratePhenotype: Starting with genome: %v (length: %d)", ind.genome, len(ind.genome))
    // log.Printf("GeneratePhenotype: Current phenotype: %v", ind.phenotype)
    // log.Printf("GeneratePhenotype: Dynamic rules: %v", genomizer.GetDynamicRules())
    // log.Printf("GeneratePhenotype: Dynamic rule stack: %v", genomizer.GetDynamicRuleStack())

    if err := genomizer.Genomize(ind.genome, ind); err != nil {

		// -- Warning --
		log.Printf("GeneratePhenotype: Failed to genomize: %v", err)
        
		return fmt.Errorf("failed to genomize: %w", err)
    }

	ind.phenotype = genomizer.GetPhenotype()
    ind.productionHistory = utils.DeepCopyProductionHistory(genomizer.GetProductionHistory())
	ind.usedCodons = genomizer.GetUsedCodons()

	// -- Debug --
	// log.Printf("GeneratePhenotype: Final phenotype: %v", ind.phenotype)

	return nil
}

// IsStateValid verifies that the individual's state is consistent.
func (ind *Individual) IsStateValid() bool {
    
    // Check that the phenotype is not nil or empty.
    if ind.phenotype == nil || ind.phenotype == "" {

        log.Printf("[DEBUG] IsStateValid: FAILED - Phenotype is nil or empty")

        return false
    }

    // Verify that the production history is not empty.
    if len(ind.productionHistory) == 0 {
        
        log.Printf("[DEBUG] IsStateValid: FAILED - ProductionHistory is empty")

        return false
    }

    // Check that the genome is not empty.
    if len(ind.genome) == 0 {
        
        log.Printf("[DEBUG] IsStateValid: FAILED - Genome is empty")

        return false
    }

    // Check that LastValidPhenotype is not nil.
    if ind.isExhausted && ind.lastValidPhenotype == nil {
        
        log.Printf("[DEBUG] IsStateValid: FAILED - LastValidPhenotype is nil and exhausted = true")
        
        return false
    }

    // --- Check ONLY dynamic rules in productionHistory ---
    // Verify that the ncRNAs are consistent with the history: if the history 
    // contains dynamic rules, they must be in dynamicRules.
    for _, production := range ind.productionHistory {

        for _, rule := range production {
        
            // Check only the rules that are supposed to be dynamic: those 
            // containing "_exp" or another pattern.
            if rule.GetSymbolType() == NonTerminal {
                ruleText := rule.GetText()
        
                // If the rule has a dynamic rule name (e.g., contains "_exp").
                if strings.Contains(ruleText, "_exp") {
        
                    if _, exists := ind.dynamicRules[ruleText]; !exists {
        
                        // -- Debug --
                        log.Printf(
                            "[DEBUG] IsStateValid: FAILED - Dynamic rule %q not found in dynamicRules",
                            ruleText,
                        )
        
                        return false
                    }
        
                }
        
                // Sinon, c'est une règle statique : on ne vérifie pas
            }
        
        }
        
    }

    log.Printf("[DEBUG] IsStateValid: PASSED")

    return true
}

func (ind *Individual) MutateCodon(index int, value int) {

    if index >= 0 && index < len(ind.genome) {
        ind.genome[index] = value
    }

}

// Repair uses the immune system to repair the individual.
func (ind *Individual) Repair(immuneSys IImmune) error {
    return immuneSys.RepairIndividual(ind)
}

// SetDynamicRules sets the individual's dynamic rules (non-coding RNA).
func (ind *Individual) SetDynamicRules(value map[string]IRuleModel) {

    // Creates a deep copy to avoid shared references.
    ind.dynamicRules = make(map[string]IRuleModel, len(value))
    
	// for k, v := range value {
    //     ind.dynamicRules[k] = v
    // }
	maps.Copy(ind.dynamicRules, value)
}

// SetDynamicRuleStack sets the stack of active ncRNAs for the individual.
func (ind *Individual) SetDynamicRuleStack(value []string) {

    // Creates a copy to avoid shared references.
    ind.dynamicRuleStack = make([]string, len(value))
    copy(ind.dynamicRuleStack, value)
}

func (ind *Individual) SetFitness(value float64) {
    ind.fitness = value
}

func (ind *Individual) SetGenome(value []int) {
    ind.genome = value
}

func (ind *Individual) SetExhausted(exhausted bool) {
    ind.isExhausted = exhausted
}

// Setter for LastValidPhenotype.
func (ind *Individual) SetLastValidPhenotype(phenotype any) {
    ind.lastValidPhenotype = phenotype
}

func (ind *Individual) SetOldProductionFitness(key string, value float64) {
    
	if ind.oldProductionFitness == nil {
        ind.oldProductionFitness = make(map[string]float64)
    }
    
	ind.oldProductionFitness[key] = value
}

// SetOrganism associates an organism with the individual.
func (ind *Individual) SetOrganism(value IOrganism) {
	ind.organism = value
}

// Setter for phenotype.
func (ind *Individual) SetPhenotype(value any) {
    ind.phenotype = value
}

// Setter for productionHistory.
func (ind *Individual) SetProductionHistory(value [][]IRuleModel) {
    ind.productionHistory = utils.DeepCopyProductionHistory(value)
}

// SetProductionStep replaces a specific step in productionHistory.
func (ind *Individual) SetProductionStep(index int, step []IRuleModel) {

    if index < 0 || index >= len(ind.productionHistory) {
        panic("index out of range")
    }
    
	// Deep copy of the new step (to avoid shared references).
    newStep := make([]IRuleModel, len(step))
    for j, rule := range step {
        newStep[j] = rule.Clone() // Use Clone() for deep copying
    }
    
	ind.productionHistory[index] = newStep
}

// Setter for template.
func (ind *Individual) SetTemplate(value any) {
    ind.template = value
}

// Setter for usedCodons.
func (ind *Individual) SetUsedCodons(value int) {
    ind.usedCodons = value
}

// String returns a textual representation of the individual.
func (ind *Individual) String() string {

	// Returns a formatted string representation of the Individual, including
	// its phenotype, fitness and used codons.
	return fmt.Sprintf(
		"phenotype = %s, fitness = %.2f, Used Codons: %d", 
		ind.phenotype, 
		ind.fitness,
		ind.usedCodons,
	)
	
}

func (ind *Individual) UpdateGenomeSegment(segment []int, offset int) {
    copy(ind.genome[offset:offset+len(segment)], segment)
}