// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package optimizations

import "evoGo/model"

type(
	IRuleModel = model.IRuleModel
	RuleModel = model.RuleModel
	SymbolType = model.SymbolType
)

// Constants for symbol types.
const (
	Terminal    SymbolType = model.Terminal
	NonTerminal SymbolType = model.NonTerminal
)

// ExpandAndSimplifyRepetition applies ExpandRepetition and SimplifyRepetition.
func ExpandAndSimplifyRepetition(
	repetition string, 
	rhs [][]IRuleModel, 
	rules map[string]IRuleModel,
) [][]IRuleModel {
    expanded := ExpandRepetition(repetition, rhs, rules)
    return SimplifyRepetition(expanded, rules)
}

// FactorizeAndSimplify applies LeftFactorize and SimplifyRepetition.
func FactorizeAndSimplify(
	rule IRuleModel, 
	rules map[string]IRuleModel,
) [][]IRuleModel {

	// Applies left-hand factorization to the rule.
    factorizedRule := LeftFactorize(rule, rules)

	// Applies the repetition simplification to the RHS of the factored rule.
    return SimplifyRepetition(factorizedRule.GetSymbols(), rules)
}