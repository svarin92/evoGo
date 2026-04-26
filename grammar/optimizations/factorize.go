// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package optimizations

import (
	"slices"

	"evoGo/model"
)

// DirectLeftRecurse eliminates direct left recursion in a rule.
// Parameters:
// - rule: the rule to process (e.g., A → Aa | b).
// - rules: the set of rules in the grammar.
// Returns a new rule without direct left recursion 
//   (e.g., A → bA_tail, A_tail → aA_tail | ε).
// Note: If no recursion is detected, the rule is returned unchanged.
func DirectLeftRecurse(
	rule IRuleModel,
	rules map[string]IRuleModel,	
) IRuleModel {
	var nonRecursiveProductions [][]IRuleModel
	var recursiveProductions [][]IRuleModel

	// Separate recursive and non-recursive productions.
	for _, production := range rule.GetSymbols() {

		if len(production) > 0 && production[0].GetText() == rule.GetText() {

			// Recursive case: we add the production without the first symbol 
			// (ex: Aa → a).
			recursiveProductions = append(recursiveProductions, production[1:])
		} else {
			nonRecursiveProductions = append(nonRecursiveProductions, production)
		}

	}

	// If no direct left recursion is found, return the rule unchanged.
	if len(recursiveProductions) == 0 {
		return rule
	}

	// Creation of a new non-terminal for the recursive part (e.g.,: A_tail).
	newNonTerminal := model.NewRuleModel(
		rule.GetText() + "_tail",
		NonTerminal,
		[][]IRuleModel{},
	)

	// Add recursive productions to the new non-terminal.
	for _, production := range recursiveProductions {

		// Create a new production by adding the new non-terminal to the end.
        newProduction := make([]IRuleModel, len(production))
        copy(newProduction, production)
		newProduction = append(newProduction, model.NewRuleModel(newNonTerminal.GetText(), NonTerminal, nil))
        currentRHS := newNonTerminal.GetSymbols()
		newNonTerminal.SetSymbols(append(currentRHS, newProduction))
	}

	// Adding an epsilon production to terminate the recursion.
	epsilonProduction := []IRuleModel{model.NewRuleModel("ε", Terminal, nil)}
	currentRHS := newNonTerminal.GetSymbols()
	newNonTerminal.SetSymbols(append(currentRHS, epsilonProduction))

	// Update the original symbol productions to use the new non-terminal.
	var newProductions [][]IRuleModel

	for _, production := range nonRecursiveProductions {
        newProduction := make([]IRuleModel, len(production))
        copy(newProduction, production)		
		newProduction = append(newProduction, model.NewRuleModel(newNonTerminal.GetText(), NonTerminal, nil))
		newProductions = append(newProductions, newProduction)
	}

	// Update the productions of the original rule.
	rule.SetSymbols(newProductions)

	// Add the new non-terminal to the grammar.
	rules[newNonTerminal.GetText()] = newNonTerminal

	return rule
}

// FindLongestCommonPrefix finds the longest common prefix among a set of 
// productions.
// Parameters:
// - productions: list of productions to analyze (e.g., [[a, b], [a, c]]).
// Returns the common prefix (e.g., [a]).
// Note: Returns nil if no production is provided or if no prefix is ​​found.
func FindLongestCommonPrefix(productions [][]IRuleModel) []IRuleModel {

	if len(productions) <= 1 {
		return nil
	}

	prefix := productions[0]

	for _, production := range productions[1:] {
		newLength := 0

		for newLength < len(prefix) && newLength < len(production) {

			if prefix[newLength].GetText() != production[newLength].GetText() {
				break
			}

			newLength++
		}

		// Truncate `prefix`to the new length.
		prefix = prefix[:newLength]

		if len(prefix) == 0 {
			return nil
		}

	}

	return prefix
}

// HasCircularDependency checks if a circular dependency exists between two
//  symbols.
// Parameters:
// - current: current symbol.
// - target: target symbol.
// - visited: map of symbols already visited.
// - dependencies: graph of dependencies between symbols.
// Returns true if a circular dependency is detected.
func HasCircularDependency(
	current, target string, 
	visited map[string]bool, 
	dependencies map[string][]string,
) bool {

	if visited[current] {
		return false
	}

	visited[current] = true

	for _, dependency := range dependencies[current] {

		if dependency == target {
			return true
		}

		if HasCircularDependency(dependency, target, visited, dependencies) {
			return true
		}

	}

	return false
}

// LeftFactorize applies left factorization to a rule.
// Parameters:
// - rule: the rule to factor (e.g., A → ab | ac | ad).
// - rules: the set of rules in the grammar.
// Returns a new factored rule (e.g., A → aA_tail, A_tail → b | c | d).
// Note: Uses DirectLeftRecurse to handle direct left recursion.
func LeftFactorize(
    rule IRuleModel,
    rules map[string]IRuleModel,
) IRuleModel {
    
	// 1. Removal of duplicates in productions.
    uniqueProductions := RemoveDuplicates(rule.GetSymbols())
    rule.SetSymbols(uniqueProductions)

    // 2. Solving the direct left recursion.
    rule = DirectLeftRecurse(rule, rules)

    // 3. Global left factorization (all productions combined).
    allProductions := rule.GetSymbols()
    commonPrefix := FindLongestCommonPrefix(allProductions)

    newProductions := [][]IRuleModel{}

    if len(commonPrefix) > 0 {
        
		// Creation of a single non-terminal for the common prefix.
        newNonTerminal := model.NewRuleModel(
            rule.GetText()+"_tail",  // Generic name: A_tail
            NonTerminal,
            [][]IRuleModel{},
        )

        // Adding suffixes to the new non-terminal.
        for _, production := range allProductions {
            suffix := production[len(commonPrefix):]
            
			if len(suffix) == 0 {
                suffix = []IRuleModel{model.NewRuleModel("ε", Terminal, nil)}
            }
            
			currentSuffixes := newNonTerminal.GetSymbols()
            newNonTerminal.SetSymbols(append(currentSuffixes, suffix))
        }

        // Update of the original rule productions.
        newProduction := make([]IRuleModel, len(commonPrefix))
        copy(newProduction, commonPrefix)
        newProduction = append(newProduction, newNonTerminal)
        newProductions = append(newProductions, newProduction)

        // Added the new non-terminal to the `rules` map.
        rules[newNonTerminal.GetText()] = newNonTerminal
    } else {

        // If no common prefix, keep the original productions.
        newProductions = append(newProductions, allProductions...)
    }

    rule.SetSymbols(newProductions)
    return rule
}

// IndirectLeftRecurse detects and resolves indirect left recursion.
// Parameters:
// - rules: the set of rules for the grammar.
// Note: Uses SubstituteDependentNT to substitute circular dependencies.
func IndirectLeftRecurse(rules map[string]IRuleModel) {

	// Create a map to track dependencies.
	dependencies := map[string][]string{}

	// Complete the dependency map.
	for symbol, rule := range rules {

		for _, production := range rule.GetSymbols() {

			for _, term := range production {

				if _, exists := rules[term.GetText()]; exists {
					dependencies[symbol] = append(dependencies[symbol], term.GetText())
				}

			}

		}

	}

	// Detect and resolve circular dependencies.
	for symbol := range dependencies {
		visited := make(map[string]bool)

		if HasCircularDependency(symbol, symbol, visited, dependencies) {
			substitutedRule := SubstituteDependentNT(symbol, dependencies, rules)
			factorizedRule := LeftFactorize(substitutedRule, rules)
			rules[symbol] = factorizedRule
		}

	}

}

// SubstituteDependentNT substitutes dependent non-terminals in a rule.
// Parameters:
// - symbol: symbol to substitute.
// - dependencies: dependency graph.
// - rules: set of rules.
// Returns a new rule with the substitutions applied.
func SubstituteDependentNT(
	symbol string, 
	dependencies map[string][]string,
	rules map[string]IRuleModel,
) IRuleModel {
	currentRule := rules[symbol]
	newProductions := make([][]IRuleModel, 0)

	for _, production := range currentRule.GetSymbols() {
		substitutionDone := false

		for i, term := range production {

			if _, isNonTerminal := rules[term.GetText()]; isNonTerminal {

				// Check if it is a circular dependency.
				if slices.Contains(dependencies[symbol], term.GetText()) {
					substitutionDone = true

					for _, dependentProduction := range rules[term.GetText()].GetSymbols() {

						// Add all the part present before the dependent non-terminal.
						newProduction := slices.Clone(production[:i])

						// Add the production of the dependent non-terminal.
						newProduction = append(newProduction, dependentProduction...)

						// Add all the part present after the dependent non-terminal.
						newProduction = append(newProduction, production[i+1:]...)

						// Add to our new productions.
						newProductions = append(newProductions, newProduction)
					}

					break
				}

			}

		}

		if !substitutionDone {
			newProductions = append(newProductions, production)
		}

	}

	// Returns a new modified rule.
	return model.NewRuleModel(
		currentRule.GetText(),        // Symbol
		currentRule.GetSymbolType(),  // SymbolType
		newProductions,               // rhs
	)
}