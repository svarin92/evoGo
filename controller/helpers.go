// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"evoGo/model"
	"fmt"
	"log"
	"strconv"
	"strings"
)

/* Generic helpers */

// Check if a symbol is present in a production. Used to detect recursive 
// rules.
func ContainsSymbol(prod []IRuleModel, symbol string) bool {
    
    for _, s := range prod {
    
        if s.GetText() == symbol {
            return true
        }
    
    }
    
    return false
}

// CloneSymbols creates a deep copy of a rule map (codons).
func CloneSymbols(original map[string]IRuleModel) map[string]IRuleModel {
	cloned := make(map[string]IRuleModel)

	for k, v := range original {

		// Clone each sublist of rhs ([][]*RuleModel).
		rhsCopy := make([][]IRuleModel, len(v.GetSymbols()))

		for i, subList := range v.GetSymbols() {
			rhsCopy[i] = append([]IRuleModel(nil), subList...)  // Clone the sublist
		}

		cloned[k] = model.NewRuleModel(
            v.GetText(), 
            v.GetSymbolType(),
			rhsCopy, // Uses rhs deep copy
        )
	}

	return cloned
}

// Helper to convert the phenotype to a string.
func ConvertPhenotypeToString(phenotype any) (string, error) {

    switch p := phenotype.(type) {
    case string:
        return p, nil
    case int, int32, int64, float32, float64:
        return fmt.Sprintf("%v", p), nil
    default:
        return "", fmt.Errorf("unsupported type: %T", phenotype)
    }

}

// Finds the maximum fitness in a list of individuals.
func FindMaxFitness(individuals []*Individual) float64 {
	
    if len(individuals) == 0 {

        // -- Warning --
        log.Printf("FindMaxFitness: no individuals")
        
        return 0.0
    }

    maxFitness := individuals[0].GetFitness()

	for _, ind := range individuals[1:] {

		if ind.GetFitness() > maxFitness {
			maxFitness = ind.GetFitness()
		}

	}

    // -- Debug --
    // log.Printf("FindMaxFitness: maxFitness = %v", maxFitness)

	return maxFitness
}

// Format the contents of a string reduction cache (map[string][][]*RuleModel) 
// for logs.
// Example: 
// - map[string][][]*RuleModel{"a": [["a" (Terminal) → "vowel" (NonTerminal)], 
//   ["a" (Terminal) → "letter" (NonTerminal)]], ...}.
func FormatCacheContent(cache *ReductionChainCache) string {
    cache.mutex.Lock()
    defer cache.mutex.Unlock()

    formatted := make(map[string]string)

    for symbol, chains := range cache.cache {
        var chainStrings []string
    
        for _, chain := range chains {
            var symbols []string
    
            for _, rule := range chain {
                symbols = append(symbols, fmt.Sprintf("%s (%v)", rule.GetText(), rule.GetSymbolType()))
            }
    
            chainStrings = append(chainStrings, strings.Join(symbols, " → "))
        }
    
        formatted[symbol] = strings.Join(chainStrings, " | ")
    }
    
    return fmt.Sprintf("%v", formatted)
}

// Format a reduced string into a readable string.
// Example: 
// - []string{"space", "letter", "string"} → "space → letter → string".
func FormatReductionChain(chain []string) string {

    if len(chain) == 0 {
        return "Chaîne de réduction vide"
    }
    
    return strings.Join(chain, " → ")
}

// If the string is a list of *RuleModel or a similar structure.
func FormatRuleModelChain(chain []IRuleModel) string {
    var symbols []string

    for _, rule := range chain {
        symbols = append(symbols, rule.GetText())
    }

    return strings.Join(symbols, " → ")
}

// GetBaseSymbols extracts the symbol names from a production or alternative 
// (ignoring _tail and indices).
func GetBaseSymbols(symbols []IRuleModel) []string {
    baseSymbols := make([]string, len(symbols))
    
    for i, symbol := range symbols {
        baseSymbols[i] = strings.TrimSuffix(
            strings.TrimSpace(strings.Split(symbol.GetText(), " ")[0]),
            "_tail",
        )
    
    }
    
    return baseSymbols
}

// GenerateDynamicRuleName generates a unique name for a dynamic rule (ARNnc)
// from a sequence of symbols ([]IRuleModel) or productions ([][]IRuleModel).
func GenerateDynamicRuleName(sequence interface{}) string {
    var parts []string

    // Determine the sequence type and extract the symbol names.
    switch seq := sequence.(type) {
    case []IRuleModel:
        
        // Case 1: sequence is a list of symbols 
        //         (e.g., [letter 1, letter 1, letter 1]).
        for _, symbol := range seq {
            symbolName := strings.TrimSpace(strings.Split(symbol.GetText(), " ")[0])
            parts = append(parts, symbolName)
        }
    
    case [][]IRuleModel:
    
        // Case 2: sequence is a list of productions 
        //         (e.g., [[consonant 1], [vowel 1], [consonant 1]]).
        for _, production := range seq {
    
            if len(production) > 0 {
                symbolName := strings.TrimSpace(strings.Split(production[0].GetText(), " ")[0])
                parts = append(parts, symbolName)
            }
    
        }
    
    default:
        return "unknown_exp"  // Default case (should not happen)
    }

    return strings.Join(parts, "_") + "_exp"
}

// HasTailSymbol checks if a production contains a symbol with the suffix 
// "_tail".
func HasTailSymbol(production []IRuleModel) bool {
    
    for _, symbol := range production {
    
        if strings.HasSuffix(symbol.GetText(), "_tail") {
            return true
        }
    
    }
    
    return false
}

// ParseSymbol analyze a symbol string to extract the symbol and its type.
func IsSameSymbol(a, b string) bool {
    
    // Extraire le nom du symbole (premier mot).
    aName := strings.Split(strings.TrimSpace(a), " ")[0]
    bName := strings.Split(strings.TrimSpace(b), " ")[0]
    
    // Comparer uniquement les noms (ignorer les indices de type et les espaces supplémentaires).
    return aName == bName
}

// ParseSymbol analyze a symbol string to extract the symbol and its type.
func ParseSymbol(symbolStr string) (string, SymbolType, error) {
    parts := strings.Fields(symbolStr)
    
    if len(parts) != 2 {
        return "", 0, fmt.Errorf("incorrect symbol format: %s", symbolStr)
    }

    symbol := parts[0]
    symbolType, err := strconv.Atoi(parts[1])
    
    if err != nil {
        return "", 0, fmt.Errorf("Incorrect symbol type format: %s", parts[1])
    }

    return symbol, SymbolType(symbolType), nil
}

/* Helpers for the Genomizer.                                                  */
/* These functions de^pend on the Genomizer's grammar rules and Symbol table.  */
/* They are used to validate reductions and expansions in the context of       */ 
/* recursive grammars.                                                         */

// CanReduceTo checks if a symbol (sourceSymbol) can be reduced to another 
// symbol (targetSymbol) by tracing back up the hierarchy of production rules 
// (bottom-up logic).
// Example: space → letter if a rule exists of the type "letter → space".
func CanReduceTo(sourceSymbol, targetSymbol string, g *Genomizer) bool {

    // -- Debug --
    // log.Printf("CanReduceTo: Direct reduction verification '%s' → '%s'", sourceSymbol, targetSymbol)

    // Trivial case: the symbols are already equal.
    if sourceSymbol == targetSymbol {
        return true
    }

    // Iterate through all the rules to find if targetSymbol produces 
    // sourceSymbol, that is, if targetSymbol is a parent of sourceSymbol.
    if rule, exists := g.GetSymbols()[targetSymbol]; exists {
    
        for _, prod := range rule.GetSymbols() {
    
            // Check if the direct output of targetSymbol is sourceSymbol.
            // Example: if targetSymbol = "letter" and prod = ["space"], 
            //          then "space" can be reduced to "letter".
            if len(prod) == 1 && prod[0].GetText() == sourceSymbol {

                // -- Debug --
                // log.Printf("CanReduceTo: Direct reduction found : '%s' → '%s'", sourceSymbol, targetSymbol)

                return true
            }
    
            // Check indirect (recursive) productions.
            if len(prod) == 1 {
    
                if CanReduceTo(sourceSymbol, prod[0].GetText(), g) {
                    return true
                }
    
            }
    
        }
    
    }

    // -- Debug --
    // log.Printf("CanReduceTo: No direct reduction found for '%s' → '%s'", sourceSymbol, targetSymbol)

    // If no rule allows the reduction, return false.
    return false
}

// CanReduceToEpsilon checks if a symbol can be reduced to ε according to the 
// grammar. Uses the Genomizer to access the rules.
func CanReduceToEpsilon(symbol string, g *Genomizer) bool {
    
    // Extract the symbol name (e.g., "string_tail" from "string_tail 1").
    symbolName := strings.Split(symbol, " ")[0]

    // Retrieve the rule associated with this symbol.
    rule, exists := g.GetSymbols()[symbolName]
    
    if !exists {
        return false
    }

    // Check if ε is a possible production.
    for _, prod := range rule.GetSymbols() {
    
        if len(prod) == 1 && prod[0].GetText() == "ε" {
            return true
        }
    
    }
    
    return false
}

// CanReduceToWithEpsilon checks if `prevProduction` can be reduced to 
// `production` using the grammar's reduction rules (e.g., string_tail → ε).
func CanReduceToWithEpsilon(prevProduction, production []IRuleModel, g *Genomizer) bool {

    // Case 1: If production is empty, check that all prevProduction symbols 
    // can be reduced to ε.
    if len(production) == 0 {
        
        for _, symbol := range prevProduction {
        
            if !CanReduceToEpsilon(symbol.GetText(), g) {
                return false
            }
        
        }
        
        return true
    }

    // Case 2: Symbol-by-symbol comparison using reduction rules.
    i, j := 0, 0
    
    for i < len(prevProduction) && j < len(production) {
        prevSymbol := prevProduction[i].GetText()
        prodSymbol := production[j].GetText()

        // If the current prevProduction symbol can be reduced to ε, 
        // we ignore it.
        if CanReduceToEpsilon(prevSymbol, g) {
            i++
            continue
        }

        // If the current prevProduction symbol can be reduced to prodSymbol.
        if CanReduceTo(prevSymbol, prodSymbol, g) {
            i++
            j++
            continue
        }

        // If the symbols are identical, we pass.
        if IsSameSymbol(prevSymbol, prodSymbol) {
            i++
            j++
            continue
        }

        // No cases match.
        return false
    }

    // Check that the remaining symbols of prevProduction can be reduced to ε.
    for i < len(prevProduction) {
    
        if !CanReduceToEpsilon(prevProduction[i].GetText(), g) {
            return false
        }

        i++
    }

    // Verify that all production symbols have been matched.
    return j == len(production)
}

// IsRecursiveExpansion checks if a production is a recursive expansion
// (that is, if it can be generated by decoding a recursive production).
func IsRecursiveExpansion(
    production []IRuleModel, 
    history [][]IRuleModel, 
    index int,
    g *Genomizer) bool {

    // A recursive expansion is a production that:
    // 1. Does not contain a _tail symbol;
    // 2. Can be generated by decoding a previous recursive production.

    // Search for a previous recursive production in the history.
    for i := index - 1; i >= 0; i-- {
        prevProduction := history[i]

        // We are only interested in recursive productions (containing _tail).
        if !HasTailSymbol(prevProduction) {
            continue
        }

        // Check if `production` is a possible reduction of `prevProduction`.
        if IsReductionOf(prevProduction, production, g) {
            return true
        }
    
    }
    
    return false
}

func IsRecursiveRightExpansion(prevProduction, production []IRuleModel, g *Genomizer) bool {

    // -- Debug --
    // log.Printf(
    //     "IsRecursiveRightExpansion: Checking prevProduction=%v, production=%v",
    //     FormatRuleModelChain(prevProduction),
    //     FormatRuleModelChain(production),
    // )
    
    // Condition 1: `prevProduction` must be in the form [A string_tail] 
    // (where A is a symbol).
    if len(prevProduction) != 2 {

        // -- Debug --
        // log.Printf(
        //     "IsRecursiveRightExpansion: FAIL - prevProduction length != 2 (len=%d)",
        //     len(prevProduction),
        // )

        return false
    }

    firstSymbolPrev := prevProduction[0].GetText()  // Ex: "letter 1"
    tailSymbol := prevProduction[1].GetText()       // Ex: "string_tail 1"

    // -- Debug -- Vérification du symbole _tail.
    // log.Printf(
    //     "IsRecursiveRightExpansion: firstSymbolPrev=%s, tailSymbol=%s",
    //     firstSymbolPrev,
    //     tailSymbol,
    // )

    // Verify that the second symbol is indeed a _tail.
    if !strings.HasSuffix(strings.Split(tailSymbol, " ")[0], "_tail") {

        // -- Debug --
        // log.Printf(
        //     "IsRecursiveRightExpansion: FAIL - tailSymbol '%s' does not end with '_tail'",
        //     tailSymbol,
        // )

        return false
    }

    // Condition 2: `production` must start with the same symbol as 
    // `prevProduction[0]`.
    if len(production) == 0 {
        return false
    }
 
    firstSymbolProd := production[0].GetText()

    // -- Debug --
    // log.Printf(
    //     "IsRecursiveRightExpansion: firstSymbolProd=%s, firstSymbolPrev=%s",
    //     firstSymbolProd,
    //     firstSymbolPrev,
    // )
 
    if !IsSameSymbol(firstSymbolPrev, firstSymbolProd) {

        // -- Debug --
        // log.Printf(
        //     "IsRecursiveRightExpansion: FAIL - firstSymbolProd and firstSymbolPrev are not the same",
        // )

        return false
    }

    // Condition 3: All `production` symbols must be identical to the first 
    // symbol.
    for _, symbol := range production {
    
        if !IsSameSymbol(firstSymbolProd, symbol.GetText()) {

            // -- Debug --
            // log.Printf(
            //     "IsRecursiveRightExpansion: FAIL - production[%d]='%s' != firstSymbolProd='%s'",
            //     i,
            //     symbol.GetText(),
            //     firstSymbolProd,
            // )

            return false
        }
    
    }

    // Condition 4: Verify that `tailSymbol` can generate at least one symbol 
    // compatible with `firstSymbolProd` 
    // (e.g., string_tail → letter string_tail or string_tail → ε).
    tailName := strings.Split(tailSymbol, " ")[0]
    tailRule, exists := g.GetSymbols()[tailName]
    
    if !exists {

        // -- Debug --
        // log.Printf(
        //     "IsRecursiveRightExpansion: FAIL - no rule found for tailSymbol '%s'",
        //     tailName,
        // )

        return false
    }

    // -- Debug --
    // log.Printf(
    //     "IsRecursiveRightExpansion: tailRule for '%s' = %v",
    //     tailName,
    //     tailRule.GetSymbols(),
    // )

    // Check that at least one production of `tailRule` contains a symbol 
    // compatible with `firstSymbolProd`.
    for _, prod := range tailRule.GetSymbols() {
        
        for _, symbol := range prod {
    
            if IsSameSymbol(symbol.GetText(), firstSymbolProd) {

                // log.Printf(
                //     "IsRecursiveRightExpansion: PASS - tailRule[%d] contains compatible symbol '%s'",
                //     i,
                //     symbol.GetText(),
                // )

                return true
            }
    
        }
    
    }

    return true
}

// IsReductionOfchecks whether `production` can be obtained by reducing 
// `prevProduction` according to the rules of your grammar.
func IsReductionOf(prevProduction, production []IRuleModel, g *Genomizer) bool {
    
    // Case 1: Right recursive expansion 
    // (e.g., [letter 1 string_tail 1] → [letter 1 letter 1] 
    //        or [letter 1 letter 1 letter 1 letter 1])
    // We include cases where len(production) >= len(prevProduction), because 
    // a reduction can maintain the same length
    // (e.g., [letter 1 string_tail 1] → [letter 1 letter 1] 
    //        if string_tail → letter).
    if len(production) >= len(prevProduction) {
        return IsRecursiveRightExpansion(prevProduction, production, g)
    }

    // Case 2: Reduction with deletion (ε) 
    // (e.g., [letter 1 string_tail 1] → [letter 1] with string_tail → ε).
    if len(prevProduction) > len(production) {
        return CanReduceToWithEpsilon(prevProduction, production, g)
    }

    return false
}
