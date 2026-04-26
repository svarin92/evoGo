// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package optimizations

import (
	"fmt"

	"evoGo/model"
)

// ExpandRepetition extends repetitions within a right-hand side (RHS) of 
// a rule.
// Parameters:
// - repetition: the type of repetition ("*", "+", "?").
// - rhs: the right-hand side of the rule to extend (e.g., [[a, b]]).
// - rules: the map of the grammar rules.
// Returns a new RHS extended according to the repetition.
// Examples:
// - "*": X → a b → X_seq, X_star → X_seq X_star | ε.
// - "+": X → a b → X_seq, X_plus → X_seq X_plus | X_seq.
// - "?": X → a b → X_seq, X_seq → X_seq | ε.
func ExpandRepetition(
	repetition string, 
	rhs [][]IRuleModel, 
	rules map[string]IRuleModel,
) [][]IRuleModel {
 
	if len(rhs) == 0 || len(rhs[0]) == 0 {
        return rhs  // Limiting case: Empty RHS or empty production
    }

	// Create a new non-terminal for the entire sequence in the RHS.
	sequenceSymbol := fmt.Sprintf("%s_seq", rhs[0][0].GetText())
	sequenceRule := model.NewRuleModel(sequenceSymbol, NonTerminal, rhs)
	rules[sequenceSymbol] = sequenceRule

	// Processing according to the type of repetition.
	switch repetition {
	case "*":
		// "*" : 0 <= n < ∞ → X = sequence X | ε.
		newSymbol := fmt.Sprintf("%s_star", sequenceSymbol)
		repetitionRule := model.NewRuleModel(
            newSymbol,
            NonTerminal,
			[][]IRuleModel{
				{
					model.NewRuleModel(sequenceSymbol, NonTerminal, nil),
					model.NewRuleModel(newSymbol, NonTerminal, nil),
				},
				{
					model.NewRuleModel("ε", Terminal, nil),
				},
			},
		)
		rules[newSymbol] = repetitionRule

		// Return a production that uses the new rule.
		return [][]IRuleModel{
			{
				model.NewRuleModel(newSymbol, NonTerminal, nil),
			},
		}

	case "+":
		// "+" : n >= 1 → X = sequence X | sequence.
		newSymbol := fmt.Sprintf("%s_plus", sequenceSymbol)
		repetitionRule := model.NewRuleModel(
            newSymbol,
            NonTerminal,
			[][]IRuleModel{
				{
					model.NewRuleModel(sequenceSymbol, NonTerminal, nil),
					model.NewRuleModel(newSymbol, NonTerminal, nil),
				},
				{
					model.NewRuleModel(sequenceSymbol, NonTerminal, nil),
				},
			},
		)
		rules[newSymbol] = repetitionRule

		// Return a production that uses the new rule.
		return [][]IRuleModel{
			{
				model.NewRuleModel(newSymbol, NonTerminal, nil),
			},
		}

	case "?":
		// "?": n = 0 || n = 1 → X = sequence | ε.
		return [][]IRuleModel{
			{
				model.NewRuleModel(sequenceSymbol, NonTerminal, nil),
			},
			{
				model.NewRuleModel("ε", Terminal, nil),
			},
		}

	default:

		// Default case: no repetition, return the RHS unchanged.
		return rhs
	}

}