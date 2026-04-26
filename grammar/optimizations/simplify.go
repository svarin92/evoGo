// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package optimizations

import (
	"fmt"
	"slices"
	"strings"
)

// RemoveDuplicates removes duplicate productions from a list of productions.
// Uses a signature key including the symbol and its type to avoid conflicts.
// Parameters:
// - productions: list of productions to filter (e.g., [[a, b], [a, b], [c]]).
// Returns a list of unique productions (e.g., [[a, b], [c]]).
func RemoveDuplicates(productions [][]IRuleModel) [][]IRuleModel {
    uniqueProductions := make([][]IRuleModel, 0)
    seen := make(map[string]bool)
    
	for _, production := range productions {
    
		var signature strings.Builder
    
		for _, rule := range production {
            
			// Inclure le symbole et son type dans la signature.
			fmt.Fprintf(&signature, "%s_%d_", rule.GetText(), rule.GetSymbolType())
        }

		signatureStr := signature.String()
    
		if !seen[signatureStr] {
            seen[signatureStr] = true
            uniqueProductions = append(uniqueProductions, production)
        }
    
	}
    
	return uniqueProductions
}

// SimplifyRepetition replaces each non-terminal (ex: X) in rhs by its 
// productions (ex: a or b if X → a | b). and removes duplicates.
// Parameters:
// - rhs: list of productions to simplify (e.g., [[X], [c]]).
// - rules: set of grammar rules.
// Returns a list of simplified productions (e.g., [[a], [b], [c]]).
func SimplifyRepetition(
	rhs [][]IRuleModel,
	rules map[string]IRuleModel,
) [][]IRuleModel {
	simplified := make([][]IRuleModel, 0, len(rhs))

	// For each production in rhs (e.g., [X] or [c]), we initialize 
	// productionsAfterSubstitution with an empty production [].
	for _, production := range rhs {
		
		// Generate all possible combinations of productions after 
		// substitution.
		productionsAfterSubstitution := [][]IRuleModel{{}}

		// For each symbol in the current production (e.g., X in [X]), 
		// we  check whether it is a non-terminal or a terminal.
		for _, rule := range production {
			
			// If the symbol is a non-terminal, we check if rule (e.g., X) 
			// is defined in rules (e.g., X → a | b). In which case, we 
			// retrieve its outputs with SubstituteNT(definedRule.GetSymbols())
			// (e.g., [[a], [b]]).
			if definedRule, exists := rules[rule.GetText()]; exists {
			
				// Substitute the non-terminal with its productions.
				substitutedProductions := SubstituteNT(definedRule.GetSymbols())
				temp := make([][]IRuleModel, 0)

				// For each existing production in productionsAfterSubstitution
				// (e.g., [] at the beginning), we combine it with each 
				// substituted production (e.g., [a] and [b]):
				// - existing = [] + substituted = [a] → [a]
				// - existing = [] + substituted = [b] → [b]
				// Thus, productionsAfterSubstitution = [[a], [b]].
				for _, existing := range productionsAfterSubstitution {
					
					for _, substituted := range substitutedProductions {
						combined := append(slices.Clone(existing), substituted...)
						temp = append(temp, combined)
					}
				}

				// productionsAfterSubstitution is updated with the new 
				// combinations (temp).
				productionsAfterSubstitution = temp
			} else {
				
				// Add the terminal to all existing productions. If 
				// productionsAfterSubstitution = [[a], [b]] and 
				// rule = c, then:
				// - [a] becomes [a, c]
				// - [b] becomes [b, c]
				// Then, productionsAfterSubstitution = [[a, c], [b, c]].
				for i := range productionsAfterSubstitution {
					productionsAfterSubstitution[i] = append(productionsAfterSubstitution[i], rule)
				}

			}

		}

		// After processing all the symbols in a production, all the generated
		//  combinations are added to simplified.
		// Example, if rhs = [[X], [c]] and X → a | b, then:
		// - For [X]: productionsAfterSubstitution = [[a], [b]]
		// - For [c]: productionsAfterSubstitution = [[c]]
		// Then, simplified = [[a], [b], [c]].
		simplified = append(simplified, productionsAfterSubstitution...)
	}

	// Remove duplicates with the robust key. If simplified = [[a], [a], [b]], 
	// then RemoveDuplicates returns [[a], [b]].
	return RemoveDuplicates(simplified)
}

// SubstituteNT returns the productions of a non-terminal without modification.
// Parameters:
// - rhs: list of productions of the non-terminal (e.g., [[a], [b]]).
// Returns the same list of productions (e.g., [[a], [b]]).
func SubstituteNT(rhs [][]IRuleModel) [][]IRuleModel {
	return rhs
}