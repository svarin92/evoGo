// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"reflect"
	"slices"
	"strconv"

	// "runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"evoGo/model"
	"evoGo/utils"

	lru "github.com/hashicorp/golang-lru"
)

type FailedProduction struct {
	Production []IRuleModel
	Fitness    float64
	Timestamp  time.Time  // To avoid blacklisting a production indefinitely
}

// Structure for storing the current reduction chain.
type ReductionChainCache struct {
    cache map[string][][]IRuleModel  // List of reduction chains for each symbol
    mutex sync.Mutex
}

// Represent a successful genome with its associated fitness.
type SuccessfulGenome struct {
    Genome  []int
    Fitness float64
}

/* Genomizer */

// Genomizer (equivalent to cellular machinery): nucleus or ribosomes
type Genomizer struct {
	symbols           map[string]IRuleModel
    grammar           IGrammar
	phenotype         string
    mu                sync.Mutex
	startRule         string
	usedCodons        int
	usedWraps         int
    currentRecursiveProduction []IRuleModel  // Recursive production in progress (e.g., [letter 1 string_tail 1])

    // dynamicRules represents the cell's library of non-coding RNAs 
    // (ncRNAs). Each ncRNA (a key in the map) is a dynamic rule that 
    // regulates phenotype generation, just as miRNAs regulate gene 
    // expression by targeting specific sequences.    
    dynamicRules      map[string]IRuleModel  // Dynamic rules (non-coding RNAs or ncRNAs)
    
    // dynamicRuleStack represents the stack of active non-coding RNAs 
    // (transient ncRNAs). Like snoRNAs or siRNAs, these ncRNAs are 
    // transient and used to regulate phenotype generation at a given 
    // moment. They follow a LIFO logic.
    dynamicRuleStack  []string               // Stack for decoding
	
    failedProductions []FailedProduction
	historySize       int                    // History size
    PatternLibrary    *LinguisticPatternLibrary

	// productionHistory represents the history of productions used to 
    // generate the phenotype, like the history of transcriptions in a 
    // cell.
    // Is filled with the productions used to generate an individual's 
    // current phenotype, which gives its the grammatical derivation. This 
    // trace is used in CorrectGenome() to identify and replace suboptimal 
    // productions. It is also used in rebuildGenomeFromHistory() to 
    // recreate a genome after a correction. 
    // Example: "golden" → productionHistory = [prod1, prod2, ..., prod6].
	productionHistory [][]IRuleModel 

	// Same as SuccessfulProduction, but with limited size (e.g. 10 entries). 
    // Stores recent successful productions (e.g., last 10 generations). 
    // Allows you to favor productions that have performed  well recently, 
    // even if their average historical fitness is lower. A production that 
    // performed well 5 generations ago will have a high RecentSuccessScore, 
    // even if its average fitness across all generations is average. In 
    // RecentSuccessScore(), it is used to bias the selection towards recent
    // productions, which allows the system to quickly adapt to changes 
    // (for example, if the target changes).
	recentSuccessfulProductions []SuccessfulProduction

    reduceCache *lru.Cache  // LRU cache for CanReduceToInitialSymbol

	// Subset of "successful" productions that led to high-performing 
    // phenotypes (high fitness) and reusable to guide the future generation.
    // In CorrectGenome(), successfulProductions productions are used to 
    // replace suboptimal productions. If fitness is high, productions from 
    // productionHistory are added to successfulProductions and 
    // recentSuccessfulProductions. 
    // Example: If fitness = 0.9, then prod1, prod2, etc. are added to both 
    // lists.
	successfulProductions []SuccessfulProduction 

    successfulGenomes     []SuccessfulGenome
}

func (g *Genomizer) Create(grammar IGrammar) *Genomizer {
    
    // Initialize the LRU cache.
    cache, err := lru.New(1000)
    
    if err != nil {
        panic(fmt.Sprintf("Failed to create LRU cache in Create: %v", err))
    }

    g.symbols = CloneSymbols(grammar.GetSymbols())
    g.grammar = grammar
	g.phenotype = ""
	g.startRule = grammar.GetStartRule()
	g.historySize = 10
    g.PatternLibrary = NewLinguisticPatternLibrary()
	g.productionHistory = [][]IRuleModel{}
    g.reduceCache = cache
	g.recentSuccessfulProductions = make([]SuccessfulProduction, 0, g.historySize)
    g.successfulGenomes = make([]SuccessfulGenome, 0)

    // Initialization of fields for non-coding RNAs.
    g.dynamicRules = make(map[string]IRuleModel)  // Map initialization
    g.dynamicRuleStack = []string{}               // Stack initialization
	return g
}

// AddFailedProductions adds a production to the failure list.
func (g *Genomizer) AddToFailedProductions(production []IRuleModel, fitness float64) {
    g.failedProductions = append(g.failedProductions, FailedProduction{
        Production: production,
        Fitness:    fitness,
        Timestamp:  time.Now(),  // Optional: for future expiration
    })

	if len(g.failedProductions) > 100 {
        g.failedProductions = g.failedProductions[1:]
    }

}

func (g *Genomizer) AddToRecentSuccessfulProductions(production []IRuleModel, fitness float64) {

    // Add production to history.
    g.recentSuccessfulProductions = append(g.recentSuccessfulProductions, SuccessfulProduction{
        Production: production,
        Fitness:    fitness,
        Frequency:  1,
    })

    // Limit the size of the history.
    if len(g.recentSuccessfulProductions) > g.historySize {
        g.recentSuccessfulProductions = g.recentSuccessfulProductions[1:]
    }

}

// AddToSuccessfulGenomes adds a successful genome to the list, avoiding 
// duplicates.
func (g *Genomizer) AddToSuccessfulGenomes(genome []int, fitness float64) {
    genomeStr := fmt.Sprintf("%v", genome)

    for _, existingGenome := range g.successfulGenomes {
    
        if fmt.Sprintf("%v", existingGenome) == genomeStr {
            return  // The genome is already present
        }
    
    }

    g.successfulGenomes = append(g.successfulGenomes, SuccessfulGenome{Genome: genome, Fitness: fitness})
}

// AddToSuccessfulProductions adds a production to the subset of "successful" 
// productions.
func (g *Genomizer) AddToSuccessfulProductions(production []IRuleModel, fitness float64) {

    // Check if the production already exists.
    for i, sp := range g.successfulProductions {

        if reflect.DeepEqual(sp.Production, production) {
            g.successfulProductions[i].Frequency++
            g.successfulProductions[i].Fitness = (g.successfulProductions[i].Fitness + fitness) / 2
            return
        }
		
    }

    // Otherwise, add a new entry.
    g.successfulProductions = append(g.successfulProductions, SuccessfulProduction{
        Production: production,
        Fitness:    fitness,
        Frequency:  1,
    })
}

// ApplyAtomicReductions reduces first-rank non-terminals,  
//   (e.g., vowel → letter, consonant → letter).
func (g *Genomizer) ApplyAtomicReductions(
    currentSymbols *[]IRuleModel,
    productionSequence *[][]IRuleModel,
    depth *int,
    cache *ReductionChainCache,
) bool {
    reduced := false

    // -- Debug --
    // log.Printf("ApplyAtomicReductions: Starting with the current sequence: %v, depth: %d", *currentSymbols, *depth)    

    for i := 0; i < len(*currentSymbols); i++ {
        symbol := (*currentSymbols)[i].GetText()

        // -- Debug --
        // log.Printf("ApplyAtomicReductions: Symbol processing '%s' to the position %d", symbol, i)

        // Check if the symbol is a first-rank non-terminal, 
        //   (e.g., vowel, consonant, space).
        if _, isNonTerminal := g.GetSymbols()[symbol]; isNonTerminal {

            // -- Debug --
            // log.Printf("ApplyAtomicReductions: %s is a non-terminal, reduction search...", symbol)
            
            // Browse the rules to find a possible reduction, 
            //   (e.g., vowel → letter).
            for ntSymbol, rule := range g.GetSymbols() {

                // -- Debug --
                // log.Printf("ApplyAtomicReductions: Rule verification %s → %v", ntSymbol, rule.rhs)

                for _, prod := range rule.GetSymbols() {
            
                    // -- Debug --
                    // log.Printf("ApplyAtomicReductions: Production verification %v", prod)

                    // Check if the rule is of type "letter → vowel", 
                    //   (len(prod) == 1).
                    if len(prod) == 1 && prod[0].GetText() == symbol {

                        // -- Debug --
                        // log.Printf("ApplyAtomicReductions: Atomic rule found: %s → %v", ntSymbol, prod)
                        
                        // Replace the first-rank non-terminal with its parent.
                        oldSymbol := (*currentSymbols)[i]
                        (*currentSymbols)[i] = model.NewRuleModel(ntSymbol, NonTerminal, nil)

                        // Update productionSequence.
                        if len(*currentSymbols) > 1 {                 
                            *productionSequence = append(*productionSequence, []IRuleModel{prod[0]})
                        } else {
                            *productionSequence = append([][]IRuleModel{{prod[0]}}, *productionSequence...)
                        }

                        // Update the cache.
                        newChain  := []IRuleModel{
                            model.NewRuleModel(oldSymbol.GetText(), NonTerminal, nil),
                            model.NewRuleModel(ntSymbol, NonTerminal, nil),
                        }

                        g.UpdateReductionChain(oldSymbol.GetText(), newChain, cache, false)
                        g.UpdateReductionChain(ntSymbol, newChain , cache, false)

                        // -- Debug -- Display cache contents after update.
                        // log.Printf("ApplyAtomicReductions: Successful reduction for '%s' → '%s'. Cache after update: %s", 
                        //    oldSymbol.symbol, ntSymbol, FormatCacheContent(cache))

                        *depth++
                        reduced = true

                        // -- Debug --
                        // log.Printf("ApplyAtomicReductions: Reduction applied: %s → %s",
                        //    prod[0].symbol, ntSymbol)

                        break  // Proceed to the next iteration
                    }

                }

            }

        }

    }

    // -- Debug --
    // if !reduced {
    //   
    //    for ntSymbol, rule := range g.GetSymbols() {
    //    
    //        for _, prod := range rule.rhs {
    //
    //            log.Printf("  - %s → %s", ntSymbol, prod[0].symbol)
    //    
    //        }
    //    
    //    }
    //    
    // }
    
    return reduced
}

// Apply a direct recursive reduction to a subsequence of symbols wherever the 
// recursive symbol is in the production, 
//   (e.g., right-hand recursion: 'letter → X' if 'letter → letter X');
//   (e.g., left-hand recursion: 'letter → X' if 'letter → X letter').
func (g *Genomizer) ApplyDirectRecursiveMatches(
    currentSymbols *[]IRuleModel, 
    productionSequence *[][]IRuleModel, 
    depth *int,
    cache *ReductionChainCache,
) bool {
    reduced := false

    // -- Debug --
    // log.Printf("ApplyDirectRecursiveMatches: Starting with the current sequence: %v, depth: %d", *currentSymbols, *depth)

    // Special case for unique sequences.
    if len(*currentSymbols) == 1 {
        subSequence := (*currentSymbols)[0:1]
        ntSymbol, _, recursiveRule := g.FindRecursiveRule(subSequence)

        if ntSymbol != "" {
    
            // Apply the reduction.
            oldSymbol := (*currentSymbols)[0]
            (*currentSymbols)[0] = model.NewRuleModelWithCount(ntSymbol, NonTerminal, oldSymbol.GetCount())

            // Update productionSequence.
            if oldSymbol.GetCount() > 1 {
                *productionSequence = append([][]IRuleModel{recursiveRule}, *productionSequence...)
            } else {
                *productionSequence = append([][]IRuleModel{{oldSymbol}}, *productionSequence...)
            }

            // Build the new reduction chain.
            newChain := []IRuleModel{
                model.NewRuleModel(oldSymbol.GetText(), NonTerminal, nil),
                model.NewRuleModel(ntSymbol, NonTerminal, nil),
            }

            // Update the cache.
            g.UpdateReductionChain(oldSymbol.GetText(), newChain, cache, false)
            g.UpdateReductionChain(ntSymbol, newChain, cache, false)

            *depth++
            reduced = true

            // -- Debug --
            // log.Printf("ApplyDirectRecursiveMatches: Direct recursive reduction applied: %v → %s (règle: %v), new séquence: %v",
            //     []*RuleModel{oldSymbol}, ntSymbol, recursiveRule, *currentSymbols)
            // log.Printf("ApplyDirectRecursiveMatches: cache contents: %s", FormatCacheContent(cache))

            return reduced
        }

    }

    // -- Debug --
    log.Printf("ApplyDirectRecursiveMatches: No recursive reduction applied.")

    return reduced  // Exit as soon as a reduction is applied
}

// ApplyIndirectMatches reduces terminal/non-terminal sequences to 
// non-terminals via indirect recursive rules,  
//   (e.g., 'letter → X' if 'letter → "a" X').
func (g *Genomizer) ApplyIndirectMatches(
    currentSymbols *[]IRuleModel,
    productionSequence *[][]IRuleModel,
    depth *int,
    cache *ReductionChainCache,
) bool {
    reduced := false

    // -- Debug --
    // log.Printf("ApplyIndirectMatches: Début avec currentSymbols = %v", *currentSymbols)

    for startIndex := 0; startIndex < len(*currentSymbols); startIndex++ {
        symbol := (*currentSymbols)[startIndex].GetText()

        // Retrieve all reduction strings for the symbol.
        chains := g.GetReductionChain(symbol, cache)
        
        if len(chains) == 0 {
            
            // -- Debug --
            // log.Printf("ApplyIndirectMatches: No reduction chain found for '%s'", symbol)
            
            continue
        }

        // Traverse all reduction chains for the symbol.
        for _, chain := range chains {

            // The first element of the chain is the original terminal.
            if len(chain) == 0 {
                continue
            }
            
            originalTerminal := chain[0].GetText()

            // Construct the subsequence with the original terminal and the 
            // following symbol.
            subSequence := []IRuleModel{model.NewRuleModel(originalTerminal, Terminal, nil)}

            if startIndex+1 < len(*currentSymbols) {
                subSequence = append(subSequence, (*currentSymbols)[startIndex+1])
            }

            // -- Debug --
            // log.Printf("ApplyIndirectMatches: Constructed subsequence: %v", subSequence)

            // Find the production rule corresponding to the subsequence.
            // ntSymbol, prod, recursiveSymbol := g.FindRecursiveRule(subSequence)
            ntSymbol, prod, _ := g.FindRecursiveRule(subSequence)
            
            if ntSymbol == "" {
                
                // -- Debug --
                // log.Printf("ApplyIndirectMatches: No rule matches the subsequence %v", subSequence)
                
                continue
            }

            // -- Debug --
            // log.Printf("ApplyIndirectMatches: Matching rule found: %s → %v", ntSymbol, prod)

            // Apply indirect recursive reduction.
            newProduction := make([]IRuleModel, len(prod))
            for j, s := range prod {
                newProduction[j] = model.NewRuleModel(s.GetText(), NonTerminal, nil)
            }

            *productionSequence = append([][]IRuleModel{newProduction}, *productionSequence...)
            newSymbols := make([]IRuleModel, 0)
            newSymbols = append(newSymbols, (*currentSymbols)[:startIndex]...)
            newSymbols = append(newSymbols, model.NewRuleModel(ntSymbol, NonTerminal, nil))

            if startIndex+1 < len(*currentSymbols) {
                newSymbols = append(newSymbols, (*currentSymbols)[startIndex+1:]...)
            }

            // Save the original subsequence before modification.
            // originalSubSequence := []IRuleModel{(*currentSymbols)[startIndex]}

            *currentSymbols = newSymbols
            *depth++
            reduced = true

            // Update the cache for each symbol in the subsequence.
            newChain := append(chain, model.NewRuleModel(ntSymbol, NonTerminal, nil))
            g.UpdateReductionChain(originalTerminal, newChain, cache, false)

            // -- Debug --
            // log.Printf("ApplyIndirectMatches: Reduction applied: %v → %s (recursive symbol: %s), new sequence: %v",
            //     originalSubSequence, ntSymbol, recursiveSymbol, *currentSymbols)

            return reduced
        }

    }

    // -- Debug --
    // if !reduced {
    //     log.Printf("ApplyIndirectMatches: Aucune réduction indirecte appliquée.")
    // }
    
    return reduced
}

// Example 1: [vowel consonant] → syllable.
func (g *Genomizer) ApplyMixedSequenceReduction(
    currentSymbols *[]IRuleModel,
    productionSequence *[][]IRuleModel,
    depth *int,
    cache *ReductionChainCache,
) bool {
    reduced := false
    i := 0

    // -- Debug --
    // log.Printf("ApplyMixedSequenceReduction: Starting with the current sequence: %v, depth: %d", *currentSymbols, *depth)

    // Browse the sequence.
    for i < len(*currentSymbols) {

        // Test all possible subsequence lengths, starting with the longest.
        maxSubSeqLength := len(*currentSymbols) - i
        
        for subSeqLength := maxSubSeqLength; subSeqLength >= 2; subSeqLength-- {
            subSequence := (*currentSymbols)[i : i+subSeqLength]
        
            // log.Printf("ApplyMixedSequenceReduction: Test of the %v subsequence (indices %d:%d, depth %d)",
            //    subSequence, i, i+subSeqLength, subSeqLength)

            // Check that the subsequence contains at least 2 different symbols 
            // (to avoid homogeneous sequences).
            isMixed := false

            if len(subSequence) >= 2 {
                firstSymbol := subSequence[0].GetText()
            
                for _, sym := range subSequence[1:] {
            
                    if sym.GetText() != firstSymbol {
                        isMixed = true
                        break
                    }
            
                }
            
            }
            
            if !isMixed {
                continue  // Ignore homogeneous sequences
            }

            // Find a reduction rule for this subsequence.
            // ntSymbol, matchedRule := g.FindMatchingRule(subSequence, cache)
            ntSymbol, _ := g.FindMatchingRule(subSequence, cache)
        
            if ntSymbol != "" {
        
                // Apply reduction.
                originalSubSequence := make([]IRuleModel, subSeqLength)
                copy(originalSubSequence, subSequence)

                newSymbols := make([]IRuleModel, 0)
                newSymbols = append(newSymbols, (*currentSymbols)[:i]...)
                newSymbols = append(newSymbols, model.NewRuleModel(ntSymbol, model.NonTerminal, nil))
        
                if i+subSeqLength < len(*currentSymbols) {
                    newSymbols = append(newSymbols, (*currentSymbols)[i+subSeqLength:]...)
                }

                *currentSymbols = newSymbols
                *productionSequence = append([][]IRuleModel{originalSubSequence}, *productionSequence...)
                *depth++
                reduced = true

                // Update the cache for each symbol in the original subsequence.
                for _, sym := range originalSubSequence {
                    chains := g.GetReductionChain(sym.GetText(), cache)
                    newChain := append(chains[0], model.NewRuleModel(ntSymbol, model.NonTerminal, nil))
                    g.UpdateReductionChain(sym.GetText(), newChain, cache, false)
                }

                // Start again from the beginning of the sequence.
                i = 0
                break  // Exit the loop of subsequence length
            }

        }

        i++
    }

    return reduced
}

// ApplySequenceReduction reduces symbol sequences using non-recursive 
// production rules. It focuses solely on homogeneous sequences and ignores 
// mixed sequences during this phase. This ensures that [letter 1 letter 1 ...] 
// is fully reduced to [syllable 1 syllable 1 ...] before any other operation.
// Example 1: [letter, letter, letter] → letters.
// Example 2: ["a" letter] → [letter letter] → syllabe.
// Example 3: [letter "b"] → [letter letter] → syllabe.
func (g *Genomizer) ApplySequenceReduction(
    currentSymbols *[]IRuleModel,
    productionSequence *[][]IRuleModel,
    depth *int,
    cache *ReductionChainCache,
) bool {
    reduced := false
    i := 0

    // -- Debug --
    // log.Printf("ApplySequenceReduction: Starting with the current sequence: %v, depth: %d", *currentSymbols, *depth)

    for i < len(*currentSymbols) {
    
        // Retrieve the current symbol via the interface.
        symbol := (*currentSymbols)[i].GetText()
        maxLength := 1

        // Find the longest subsequence of identical consecutive symbols.
        for i+maxLength < len(*currentSymbols) && (*currentSymbols)[i+maxLength].GetText() == symbol {
            maxLength++
        }

        // If the subsequence has a length >= 2, try to reduce it.
        if maxLength >= 2 {
            subSequence := (*currentSymbols)[i : i+maxLength]
    
            // -- Debug --
            // log.Printf("ApplySequenceReduction: Checking the subsequence %v (indices %d:%d, length %d)",
            //    subSequence, i, i+maxLength, maxLength)

            // Find a reduction rule for this homogeneous subsequence.
            // ntSymbol, matchedRule := g.FindMatchingRule(subSequence, cache)
            ntSymbol, _ := g.FindMatchingRule(subSequence, cache)
    
            if ntSymbol != "" {  // Apply reduction
    
                // Create a copy of the original subsequence.
                originalSubSequence := make([]IRuleModel, maxLength)
                copy(originalSubSequence, subSequence)

                // Construct the new sequence with the non-terminal symbol.
                newSymbols := make([]IRuleModel, 0)
                newSymbols = append(newSymbols, (*currentSymbols)[:i]...)
                newSymbols = append(newSymbols, model.NewRuleModel(ntSymbol, NonTerminal, nil))
    
                if i+maxLength < len(*currentSymbols) {
                    newSymbols = append(newSymbols, (*currentSymbols)[i+maxLength:]...)
                }

                // Update currentSymbols and productionSequence.
                *currentSymbols = newSymbols
                *productionSequence = append([][]IRuleModel{originalSubSequence}, *productionSequence...)
                *depth++
                reduced = true

                // -- Debug --
                // log.Printf("ApplySequenceReduction: Reduction applied: %v → %s (rule: %s → %v)",
                //    subSequence, ntSymbol, ntSymbol, matchedRule.GetSymbols()[0])

                // Update the cache for each symbol in the original subsequence.
                for _, sym := range originalSubSequence {
    
                    // Retrieve existing reduction strings for the symbol.
                    chains := g.GetReductionChain(sym.GetText(), cache)

                    // Build the new reduction chain: [chaîne existante] + 
                    // [nouveau symbole non-terminal (ntSymbol)]
                    newChain := append(chains[0], model.NewRuleModel(ntSymbol, model.NonTerminal, nil))

                    // Update the cache with the new string.
                    g.UpdateReductionChain(sym.GetText(), newChain, cache, false)

                    // -- Debug --
                    // log.Printf("ApplySequenceReduction: Cache updated for %s: %v", sym.GetText(), cache.cache[sym.GetText()])
                    // log.Printf("ApplySequenceReduction: Cache updated : %s", FormatCacheContent(cache))
                }

                // Start again from the beginning of the sequence.
                i = 0
                continue
            }
    
        }

        i++
    }
    
    // -- Debug --
    // log.Printf("ApplySequenceReduction: New sequence : %v, depth : %d",
    //    *currentSymbols, *depth)

    return reduced  // Exit after the first reduction has been applied
}

// ApplySequenceSimplification simplifies a sequence using reduction rules,
//   (e.g., [syllable syllable] → [syllable_2]).
func (g *Genomizer) ApplySequenceSimplification(
    currentSymbols *[]IRuleModel,
    productionSequence *[][]IRuleModel,
    depth *int,
    cache *ReductionChainCache,
) bool {

    // Check if the sequence contains fewer than 2 symbols.
    if len(*currentSymbols) < 2 {
        return false
    }

    simplified := false

    // -- Debug --
    // log.Printf("ApplySequenceSimplification: Starting with the current sequence: %v, depth: %d", *currentSymbols, *depth)

    // First, try to simplify the repetitive sequences.
    simplifiedSequence, reduced := g.CanReduceSequence(*currentSymbols)
    
    if reduced {
        originalSequence := make([]IRuleModel, len(*currentSymbols))
        copy(originalSequence, *currentSymbols)

        *currentSymbols = simplifiedSequence

        // Update productionSequence.
        *productionSequence = append([][]IRuleModel{originalSequence}, *productionSequence...)

        *depth++
        simplified = true

        // -- Debug --
        // log.Printf("ApplySequenceSimplification: Simplification applied: %v → %v", originalSequence, *currentSymbols)
    } else {
        
        // If simplifying the repetitions fails, try building a suffix with a 
        // recursive tail.
        suffix := ""

        for i, symbol := range *currentSymbols {
        
            if i > 0 {
                suffix += "_"
                }
        
            suffix += symbol.GetText() + "_tail"
        }

        // -- Debug --
        // log.Printf("ApplySequenceSimplification: Construction of the suffix: %s", suffix)

        // Search for a non-terminal symbol that ends with this suffix.
        for ntSymbol := range g.GetSymbols() {
        
            if strings.HasSuffix(ntSymbol, suffix) {
                baseSymbolWithUnderscore := strings.TrimSuffix(ntSymbol, suffix)
                
                // Clean up the base symbol to remove the residual underscore.
                baseSymbol := strings.TrimSuffix(baseSymbolWithUnderscore, "_")
            
                // -- Debug --
                // log.Printf("ApplySequenceSimplification: Non-terminal symbol found with suffix: %s, base symbol: %s", ntSymbol, baseSymbol)

                // Check if the basic symbol exists in the grammar.
                if _, exists := g.GetSymbols()[baseSymbol]; exists {
            
                    // Replace the current sequence with the base symbol.
                    originalSubSequence := make([]IRuleModel, len(*currentSymbols))
                    copy(originalSubSequence, *currentSymbols)

                    *currentSymbols = []IRuleModel{model.NewRuleModel(baseSymbol, NonTerminal, nil)}

                    // Update productionSequence.
                    *productionSequence = append([][]IRuleModel{originalSubSequence}, *productionSequence...)

                    *depth++
                    simplified = true

                    // -- Debug --
                    // log.Printf("ApplySequenceSimplification: Reduction applied: %v → %s", originalSubSequence, baseSymbol)
                    
                    break
                }
                
            }
        
        }

    }

    if !simplified && len(*currentSymbols) == 1 {
        
        // -- Debug --
        // log.Printf("ApplySequenceSimplification: Simplification complete with a single symbol: %v", *currentSymbols)
        
        simplified = true
    } else if !simplified {
        
        // -- Debug --
        // log.Printf("ApplySequenceSimplification: No simplification applied.")
    }

    return simplified
}

// BuildRecursiveExpansion constructs an explicit recursive expansion from a 
// factorable sequence. Example: For a sequence of 6 "letter"s, it returns 
// [letter 1, letter 1, letter 1, letter 1, letter 1, letter 1] ([]IRuleModel).
// This expansion will be inserted as a new production into the history.
func (g *Genomizer) BuildRecursiveExpansion(seq struct {
    Sequence       [][]IRuleModel
    StartIndex     int
    EndIndex       int
    FactorizedRule string
    SequenceLength int
}) []IRuleModel {

    // 1. Construct the explicit recursive expansion.
    expansion := make([]IRuleModel, seq.SequenceLength)
    
    for j := 0; j < seq.SequenceLength; j++ {
    
        // Create a rule template for each occurrence. 
        // Example: "letter 1" (assuming the rule is "letter").
        ruleText := fmt.Sprintf("%s", seq.FactorizedRule)
        expansion[j] = model.NewRuleModel(
            ruleText,
            NonTerminal,  // Symbol type (adjust according to your model)
            nil,          // No sub-symbols for simple expansion
        )
    
    }

    // -- Debug --
    // log.Printf(
    //     "BuildRecursiveExpansion: Built expansion for rule %q (length %d): %v",
    //     seq.FactorizedRule,
    //     seq.SequenceLength,
    //     expansion,
    // )

    // Return the expansion as []IRuleModel.
    return expansion
}

func (g *Genomizer) CalculateAverageFitness(individuals []*Individual) float64 {

    if len(individuals) == 0 {
        return 0.0
    }

    // Use a utility function to calculate the average.
    fitnessValues := make([]float64, len(individuals))
    
    for i, ind := range individuals {
        fitnessValues[i] = ind.GetFitness()
    }

    return utils.Average(fitnessValues)
}

// GetAverageFitness calculates the average fitness as the average of the 
// fitness values ​​associated with a particular production, calculated over 
// its entire usage history, and stored in g.successfulProductions. Fitness 
// reflects the overall quality of a production across generations. It helps 
// promote productions that have historically led to high-quality phenotypes.
func (g *Genomizer) GetAverageFitness(candidate []IRuleModel) float64 {

    if len(g.successfulProductions) == 0 {
        
        // -- Debug --
        // log.Printf("GetAverageFitness: no successful productions available")

        return 0.0
    }

    totalFitness := 0.0
    count := 0

    // Browse the "successful" productions to find those that match the candidate.
    for _, sp := range g.successfulProductions {

        if reflect.DeepEqual(sp.Production, candidate) {
            totalFitness += sp.Fitness
            count++

            // -- Debug --
            // log.Printf("GetAverageFitness: found matching production %v with fitness %v", sp.Production, sp.Fitness)
        }
		
    }

    // If no match is found, return 0.0.
    if count == 0 {

        // -- Debug --
        // log.Printf("GetAverageFitness: no matching production found for %v", candidate)

        return 0.0
    }

    // Return the average fitness.
    averageFitness := totalFitness / float64(count)
    
    // -- Debug --
    // log.Printf("GetAverageFitness: average fitness for production %v is %v", candidate, averageFitness)

    return averageFitness
}

// GetDynamicRules returns a copy of the Genomizer's dynamic rules (ncRNAs).
func (g *Genomizer) GetDynamicRules() map[string]IRuleModel {
    g.mu.Lock()
    defer g.mu.Unlock()

    if g.dynamicRules == nil {
        return make(map[string]IRuleModel)  // Avoid returning nil
    }

    // Returns a copy to prevent external modifications.
    dynamicRulesCopy := make(map[string]IRuleModel, len(g.dynamicRules))
    
    for k, v := range g.dynamicRules {
        dynamicRulesCopy[k] = v
    }
    
    return dynamicRulesCopy
}

// GetDynamicRuleStack returns a copy of the Genomizer's stack of active ncRNAs.
func (g *Genomizer) GetDynamicRuleStack() []string {
    g.mu.Lock()
    defer g.mu.Unlock()

    if g.dynamicRuleStack == nil {
        return []string{}  // Avoid returning nil
    }
    
    // Returns a copy to prevent external modifications.
    stackCopy := make([]string, len(g.dynamicRuleStack))
    copy(stackCopy, g.dynamicRuleStack)
    return stackCopy
}

func (g *Genomizer) GetPhenotype() string {
    g.mu.Lock()
    defer g.mu.Unlock()
	return g.phenotype
}

// GetProductionHistory returns a deep copy of productionHistory.
func (g *Genomizer) GetProductionHistory() [][]IRuleModel {
    g.mu.Lock()
    defer g.mu.Unlock()
    return utils.DeepCopyProductionHistory(g.productionHistory)
}

// -- Not yet used --
func (g *Genomizer) GetRandomVowel() rune {
    vowels := []rune{'a', 'e', 'i', 'o', 'u', 'y'}
    return vowels[rand.Intn(len(vowels))]
}

// -- Not yet used --
func (g *Genomizer) GetRandomConsonant() rune {
    consonants := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'z'}
    return consonants[rand.Intn(len(consonants))]
}

// GetReductionChain returns the reduction chain for a given symbol.
func (g *Genomizer) GetReductionChain(symbol string, cache *ReductionChainCache) [][]IRuleModel {
    cache.mutex.Lock()
    defer cache.mutex.Unlock()
    return cache.cache[symbol]  // Returns a list of strings.
}

func (g *Genomizer) GetSuccessfulProductions() []ISuccessfulProduction {
    result := make([]ISuccessfulProduction, len(g.successfulProductions))
    
    for i, sp := range g.successfulProductions {
        result[i] = sp  // Implicit conversion
    }
    
    return result
}

func (g *Genomizer) GetSymbols() map[string]IRuleModel {
    return g.symbols
}

// GetSymbolIndex returns the index of a symbol in the grammar. If the symbol 
// does not exist, returns -1.
func (g *Genomizer) GetSymbolIndex(symbol string) int {
    
    // Retrieves the keys (symbols) from the g.symbols map.
    symbols := make([]string, 0, len(g.symbols))
    
    for sym := range g.symbols {
        symbols = append(symbols, sym)
    }

    // Find the symbol index.
    for i, sym := range symbols {
    
        if sym == symbol {
            return i
        }
    
    }

    return -1  // Symbol not found
}

func (g *Genomizer) GetUsedCodons() int {
	return g.usedCodons
}

func (g *Genomizer) GetUsedWraps() int {
	return g.usedWraps
}

// -- No longer in use --
// AverageSimilarityToSuccessSet calculates the average similarity of a 
// candidate production with the subset of "successful" productions.
func (g *Genomizer) AverageSimilarityToSuccessSet(candidate []IRuleModel) float64 {

	if len(g.successfulProductions) == 0 {
		return 0.0
	}

	var sum float64

	for _, success := range g.successfulProductions {
		sum += g.ProductionSimilarity(candidate, success.Production)
	}

	return sum / float64(len(g.successfulProductions))
}

// CanFactorizeAsImplicitRecursiveExpansion checks whether a sequence of 
// productions  can be factorized into a recursive expansion of a given 
// rule.
func (g *Genomizer) CanFactorizeAsImplicitRecursiveExpansion(
    sequence [][]IRuleModel,  // Sequence of productions (e.g., [[consonant 1], [vowel 1], [consonant 1]])
    ruleSymbol string,
) bool {

    // Verify that all symbols in the sequence are alternatives of ruleSymbol.
    rule, exists := g.GetSymbols()[ruleSymbol]
 
    if !exists {
        return false
    }

    for _, production := range sequence {
 
        // Each production must be a unique symbol.
        if len(production) != 1 {
            return false
        }
 
        prodSymbol := production[0].GetText()

        // Check if prodSymbol is a valid alternative to ruleSymbol.
        found := false
 
        for _, alt := range rule.GetSymbols() {
 
            if len(alt) == 1 && alt[0].GetText() == prodSymbol {
                found = true
                break
            }
 
        }
 
        if !found {
            return false
        }
 
    }
 
    return true
}

// CanReduceSequence checks if a sequence can be simplified, but only if no 
// iterative rule applies.
func (g *Genomizer) CanReduceSequence(subSequence []IRuleModel) ([]IRuleModel, bool) {

    if len(subSequence) == 0 {
        return nil, false
    }

    // 1. Check if an iterative rule exists for this sequence.
    for _, rule := range g.GetSymbols() {

        for _, prod := range rule.GetSymbols() {

            if len(prod) != len(subSequence) {
                continue
            }

            // Check if the production corresponds to the sub-sequence.
            match := true

            for j := 0; j < len(prod); j++ {

                if prod[j].GetText() != subSequence[j].GetText() {
                    match = false
                    break
                }

            }

            if match {

                // There is an iterative rule: do not simplify.
                return subSequence, false
            }

        }

    }

    // 2. Otherwise, apply the simplification of repetitions.
    reducedSequence := make([]IRuleModel, 0)
    i := 0
    n := len(subSequence)

    for i < n {
        currentSymbol := subSequence[i]
        j := i + 1

        // Find the length of the maximum repeating subsequence.
        for j < n && 
            subSequence[j].GetText() == currentSymbol.GetText() &&
            subSequence[j].GetSymbolType() == currentSymbol.GetSymbolType() {
            j++
        }

        // If a repetition is found (length > 1).
        if j - i > 1 {

            // Add a symbol representing repetition (e.g., "letter_3").
            // repeatedSymbol := fmt.Sprintf("%s_%d", currentSymbol.symbol, count)            
            reducedSequence = append(reducedSequence, model.NewRuleModelWithCount(
                currentSymbol.GetText(), 
                NonTerminal, 
                j - i,
            ))

            i = j
        } else {

            // Add the symbol as is.
            // reducedSequence = append(reducedSequence, subSequence[i])
            reducedSequence = append(reducedSequence, model.NewRuleModelWithCount(
                currentSymbol.GetText(),
                currentSymbol.GetSymbolType(),
                1,  // By default, count=1
            ))
            
            // Move to the next symbol.
            i++
        }

    }

    // Check if a discount has taken place.
    if len(reducedSequence) < len(subSequence) {
        return reducedSequence, true
    }

    return subSequence, false
}

func (g *Genomizer) CanReduceToInitialSymbol(phenotype string, baseMaxDepth int) bool {

    // -- Debug --
    // log.Printf("CanReduceToInitialSymbol: Phenotype initial %v", phenotype)

    // Verify that all terminals exist in the grammar.
    for _, terminal := range strings.Split(phenotype, "") {

        if !g.grammar.HasTerminal(terminal) {

            // -- Warning --
            log.Printf("CanReduceToInitialSymbol: Terminal %q not found in grammar", terminal)
            
            return false
        }

    }

    // Adjust maxDepth based on the phenotype length.
    maxDepth := baseMaxDepth + len(phenotype)/5  // Example: +1 for every 5 characters

    cacheKey := fmt.Sprintf("%s:%d", phenotype, maxDepth)  // Include maxDepth in the key

    if g.reduceCache != nil {

        // Check the cache.
        if cachedResult, ok := g.reduceCache.Get(cacheKey); ok {

            // If the result is in the cache, return it.
            return cachedResult.(bool)
        }

    }

    // Try to reduce the phenotype with limited depth.
    _, err := g.RebuildProductionSequence(phenotype, 20)
    result := err == nil

    if g.reduceCache != nil {

        // Cache the result.
        g.reduceCache.Add(cacheKey, result)
    }

    return result
}

// -- No longer in use --
// ClearSuccessfulGenomes clears the list of successful genomes.
func (g *Genomizer) ClearSuccessfulGenomes() {
    g.successfulGenomes = make([]SuccessfulGenome, 0)
}

// CleanSuccessfulProductions cleans the list of successful productions. 
// Removes entries with too low fitness or too low frequency.
func (g *Genomizer) CleanSuccessfulProductions() {

    // -- Debug --
    // log.Printf("CleanSuccessfulProductions: before cleaning, successfulProductions count = %v", len(g.successfulProductions))

    cleaned := make([]SuccessfulProduction, 0)

    for _, sp := range g.successfulProductions {
    
		// Keep only productions with:
		// - Sufficient fitness (e.g., > 0.3)
		// - Minimum frequency (e.g., used at least twice)
		if sp.Fitness > 0.3 && sp.Frequency > 1 {
            cleaned = append(cleaned, sp)
        }
    
	}
    
	g.successfulProductions = cleaned

	// Limit the maximum size (eg: 1000 entries).
    if len(g.successfulProductions) > 500 {
        g.successfulProductions = g.successfulProductions[:500]
    }

    // -- Debug --
    // log.Printf("CleanSuccessfulProductions: after cleaning, successfulProductions count = %v", len(g.successfulProductions))
}

// CleanUpSuccessfulGenomes cleans up the worst performing or oldest genomes.
func (g *Genomizer) CleanUpSuccessfulGenomes(maxSize int) {

    if len(g.successfulGenomes) <= maxSize {
        return  // No need to clean
    }

    // Sort genomes by decreasing fitness.
    sort.Slice(g.successfulGenomes, func(i, j int) bool {
        return g.successfulGenomes[i].Fitness > g.successfulGenomes[j].Fitness
    })
    
    // Keep the best performing maxSize genomes.
    g.successfulGenomes = g.successfulGenomes[:maxSize]
}

// CorrectByGenome applies CorrectGenome and returns true if fitness improves.
func (g *Genomizer) CorrectByGenome(
    ind IIndividual,
    population []IIndividual,
    fitnessThreshold float64,
    averageFitness float64,
    fitnessFunction FitnessFunc,
) (bool, error) {

    // Convert the individual and population parameters.
	concreteInd, ok := ind.(*Individual)
    
	if !ok {
        return false, fmt.Errorf("individual is not *Individual")
    }

	concretePopulation := make([]*Individual, len(population))
    
	for i, p := range population {
        concretePop, ok := p.(*Individual)

        if !ok {
            return false, fmt.Errorf("individual %d is not *Individual", i)
        }
    
		concretePopulation[i] = concretePop
    }

    // -- Debug -- Initial log: phenotype state BEFORE correction.
    // log.Printf(
    //     "CorrectByGenome: BEFORE: phenotype: %q, len: %d",
    //     concreteInd.GetPhenotype(), len(concreteInd.GetPhenotype().(string)),
    // )

    // Save the initial state.
    oldFitness := concreteInd.GetFitness()
    oldPhenotype := concreteInd.GetPhenotype()
    oldGenome := make([]int, len(concreteInd.GetGenome()))
    copy(oldGenome, concreteInd.GetGenome())
    oldHistory := utils.DeepCopyProductionHistory(concreteInd.GetProductionHistory())

    // Store the fitness of the productions before correction.
    for _, production := range concreteInd.GetProductionHistory() {
        key := fmt.Sprintf("%v", production)
        concreteInd.SetOldProductionFitness(key, g.GetAverageFitness(production))
    }

    // Apply corrections (modifies the genome and productionHistory).
    if err := g.CorrectGenome(concreteInd, fitnessThreshold, concretePopulation, averageFitness); err != nil {
        return false, err
    }

    // Rebuild the genome and synthesize ncRNAs.
    if err := g.RebuildGenome(concreteInd, false); err != nil {
        return false, fmt.Errorf("failed to rebuild genome after correction: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByGenome: Phenotype after RebuildGenome: %v\n, len: %d\n", 
    //     concreteInd.GetPhenotype(), len(concreteInd.GetPhenotype().(string)),
    // )

    // Regenerate the phenotype using ncRNAs (already synthesized in 
    // CorrectByLinguisticPatterns).
    if err := concreteInd.GeneratePhenotype(g); err != nil {

        // Restore the initial state in case of an error.
        concreteInd.SetFitness(oldFitness)
        concreteInd.SetPhenotype(oldPhenotype)
        concreteInd.SetGenome(oldGenome)
        concreteInd.SetProductionHistory(oldHistory)
        
        return false, fmt.Errorf("failed to regenerate phenotype: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByGenome: Phenotype after Genomize: %v\n, len: %d\n", 
    //     concreteInd.GetPhenotype(),
    //     len(concreteInd.GetPhenotype().(string)),
    // )

    // Recalculate fitness.
    if err := concreteInd.Evaluate(fitnessFunction); err != nil {
        return false, err
    }

    // -- Debug --
    // log.Printf("CorrectByGenome: Phenotype after Evaluate: %v, %v", 
    //     concreteInd.GetPhenotype(), 
    //     concreteInd.GetFitness(),
    // )    

    // Verify the improvement.
    if concreteInd.GetFitness() > oldFitness {

        // -- Debug --
        // log.Printf("CorrectByGenome: SUCCESS: fitness improved from %v to %v\n", 
        //     oldFitness, 
        //     concreteInd.GetFitness(),
        // )

        return true, nil  // Improvement
    }

    // Restore the initial state if there is no improvement.
    concreteInd.SetFitness(oldFitness)
    concreteInd.SetPhenotype(oldPhenotype)
    concreteInd.SetGenome(oldGenome)
    concreteInd.SetProductionHistory(oldHistory)

    // -- Debug --
    // log.Printf("CorrectByGenome: FAILED: fitness not improved (restored old state)\n")
    // log.Printf("CorrectByGenome: phenotype after restoration: %v, %v", 
    //     concreteInd.GetPhenotype(), 
    //     concreteInd.GetFitness(),
    // ) 

    return false, nil
}

// CorrectByGrammaticalPaths corrects an individual's grammatical paths. 
// This corrector acts on grammatical paths (production sequences) to 
// ensure that the generated phenotype adheres to the grammar rules. It 
// corrects structural deviations (e.g., invalid sequences, malformed 
// recursive loops) by relying on production rules (much like an enzyme 
// that recognizes and repairs specific patterns in DNA).
// Repair enzymes (e.g., AP endonuclease) target specific motifs in DNA 
// (such as lesions or mismatches). Similarly, CorrectByGrammaticalPaths 
// targets invalid grammatical sequences in the phenotype. Just as enzymes 
// cut, modify, or resynthesize sections of DNA to restore genetic integrity,
// CorrectByGrammaticalPaths reconstructs or adjusts production paths to 
// ensure that the phenotype is grammatically valid. While enzymes act locally 
// (e.g., on a DNA strand), this corrector acts on sub-parts of the phenotype 
// (e.g., a branch of the derivation tree) without necessarily regenerating 
// the entire structure.
//
// Cellular repair follows the process:
// - Diagnosis: reconstruct the production history (productionSequence) = 
//   identify the phenotype "blocks."
// - Preparation: resynthesize ncRNAs = prepare the "molecular tools" (ncRNAs) 
//   required for repair.
// - Repair: CorrectByOptimalPaths = apply corrections with the appropriate 
//   tools (ncRNAs) in place.
func (g *Genomizer) CorrectByGrammaticalPaths(
    ind IIndividual,
    fitnessThreshold float64,
    fitnessFunction FitnessFunc,
) (bool, error) {
    concreteInd, ok := ind.(*Individual)
    
	if !ok {
        return false, fmt.Errorf("individual is not *Individual")
    }

    // Store the fitness of the productions before correction.
    for _, production := range concreteInd.GetProductionHistory() {
        key := fmt.Sprintf("%v", production)
        concreteInd.SetOldProductionFitness(key, g.GetAverageFitness(production))
    }

    // Save the initial state.
    oldState := model.SaveIndividualState(concreteInd)

    // Convert the phenotype to a string of characters.
    phenotypeStr, err := ConvertPhenotypeToString(concreteInd.GetPhenotype())

    if err != nil {
        return false, fmt.Errorf("unsupported phenotype type: %T", concreteInd.GetPhenotype())
    }

    // Define a maximum derivation depth.
    maxDepth := 20

    // -- Debug -- Start reconstruction of the production sequence.
    // log.Printf("CorrectByGrammaticalPaths: Rebuilding production sequence for phenotype: %q", phenotypeStr)

    // Find a production sequence for the phenotype.
    productionSequence, err := g.RebuildProductionSequenceFromPhenotype(phenotypeStr, maxDepth, concreteInd)
    
    if err != nil {
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("failed to rebuild production sequence: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByGrammaticalPaths: Production sequence rebuilt: %v", productionSequence)

    // hasFallbackMarker represents a cellular stress signal (such as 
    // HSPs – Heat Shock Proteins). It indicates that phenotype reduction 
    // has failed and that the individual must activate repair mechanisms 
    // (such as restoration via LastValidPhenotype).
    hasFallbackMarker := slices.Contains(g.dynamicRuleStack, "FALLBACK_MARKER")

    // If a fallback was generated, restore the phenotype and mark 
    // Exhausted.
    if hasFallbackMarker {
    
        // Restore the phenotype from immunological memory (Last Valid 
        // Phenotype)
        lastValidPhenotype := concreteInd.GetLastValidPhenotype()
        
        if lastValidPhenotype != nil && lastValidPhenotype != "" {
            concreteInd.SetPhenotype(lastValidPhenotype)
            concreteInd.SetExhausted(true)  // Mark reduction failure
            
            // -- Debug --
            // log.Printf("CorrectByGrammaticalPaths: Fallback detected. LastValidPhenotype set and Exhausted marked.") 

            // --- EXIT IF EXHAUSTED ---
            // No need to continue, because RebuildGenome will fail again.
            return false, nil
        }
    
    }

    // Update production history ONLY if productionSequence is not empty.
    if len(productionSequence) > 0 {
        
        // -- Debug --
        log.Printf(
            "[DEBUG] CorrectByGrammaticalPaths: Setting productionHistory to: %v (length: %d)",
                productionSequence,
                len(productionSequence),
        )

        concreteInd.SetProductionHistory(productionSequence)
    } else {
    
        // -- Warning --
        log.Printf("[WARNING] CorrectByGrammaticalPaths: productionSequence is empty. Skipping SetProductionHistory.")
    
        // Optional: Restore state if productionSequence is empty.
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("empty production sequence")
    }

    // Validate the current state.
    if !concreteInd.IsStateValid() {

        // -- Debug --
        log.Printf(
            "[DEBUG] CorrectByGrammaticalPaths: State is INVALID after SetProductionHistory. Reason: Phenotype=%v, History=%v, Genome=%v, Exhausted=%v, LastValidPhenotype=%v",
                concreteInd.GetPhenotype(),
                concreteInd.GetProductionHistory(),
                concreteInd.GetGenome(),
                concreteInd.GetExhausted(),
                concreteInd.GetLastValidPhenotype(),
        )

        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("invalid state after correction")
    }

    // Validate the correspondence between the generated phenotype and the 
    // initial phenotype.
    if err := g.ValidatePhenotypeFromProductionSequence(productionSequence, phenotypeStr); err != nil {

        // -- Warning --
        log.Printf("[WARNING] CorrectByGrammaticalPaths: Phenotype validation failed: %v", err)

        // Continue despite the error (the fallback sequence has already been applied).
    }

    // -- Debug --
    // log.Println("CorrectByGrammaticalPaths: Phenotype validation: SUCCESS")

    // Rebuild the genome and synthesize ncRNAs.
    if err := g.RebuildGenome(concreteInd, true); err != nil {  // usePhenotype = true
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("failed to rebuild genome after correction: %w", err)
    }

    // -- Debug --
    if !concreteInd.IsStateValid() {
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("invalid state after RebuildGenome")
    }

    // Regenerate the phenotype and production history.
    if err := concreteInd.GeneratePhenotype(g); err != nil {
        
        // Restore the initial state in case of an error.
        model.RestoreIndividualState(concreteInd, oldState)
        
        return false, fmt.Errorf("failed to regenerate phenotype: %w", err)
    }

    // -- Debug --
    if !concreteInd.IsStateValid() {
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("invalid state after GeneratePhenotype")
    }

    // -- Debug -- BEFORE CorrectByOptimalPaths.
    // log.Printf(
    //     "CorrectByGrammaticalPaths: BEFORE CorrectByOptimalPaths: Fitness: %f",
    //     concreteInd.GetFitness(),
    // )

    // Apply the corrections via optimal paths.
    if err := g.CorrectByOptimalPaths(concreteInd, fitnessThreshold); err != nil {

        // Restore the initial state in case of an error.
        model.RestoreIndividualState(concreteInd, oldState)

        return false, err
    }

    // -- Debug --
    if !concreteInd.IsStateValid() {
        model.RestoreIndividualState(concreteInd, oldState)
        return false, fmt.Errorf("invalid state after CorrectByOptimalPaths")
    }

    // Recalculate fitness after correction.
    if err := concreteInd.Evaluate(fitnessFunction); err != nil {
        
        // Restore the initial state in case of an error.
        model.RestoreIndividualState(concreteInd, oldState)

        return false, fmt.Errorf("failed to recalculate fitness: %w", err)
    }

    // Check for improvement.
    // if concreteInd.GetFitness() > oldFitness {
    if concreteInd.GetFitness() > oldState.GetFitness() {
        
        // -- Debug --
        // log.Println("CorrectByGrammaiticalPaths: Correction SUCCESS: Fitness improved")
        
        return true, nil
    }

    // Restore the initial state if there is no improvement.
    model.RestoreIndividualState(concreteInd, oldState)

    // -- Debug --
    // log.Println("CorrectByGrammaiticalPaths: Correction FAILED: No fitness improvement")

    return false, nil
}

// CorrectByLinguisticPatterns uses linguistic patterns when correcting 
// suboptimal genomes. This method scans the genome to identify poorly 
// performing codon sequences (e.g., blocks that do not generate a valid 
// phenotype or have low fitness). For each identified index, a new block 
// of codons is generated (e.g., by replacing the block with an optimal 
// pattern from the linguistic pattern library). Next the fitness of the 
// new block is evaluated.
func (g *Genomizer) CorrectByLinguisticPatterns(
    individual *Individual,
    population []*Individual,
    fitnessThreshold float64,
    averageFitness float64,
    blockLength int,
) error {

    // 1. Identify suboptimal blocks.
    suboptimalIndices := g.IdentifySuboptimalCodonBlocks(individual.GetGenome(), population, blockLength, fitnessThreshold)

    // 2. For each suboptimal block, look for an optimal pattern.
    for _, index := range suboptimalIndices {

        if index+blockLength > len(individual.GetGenome()) {
            continue
        }

        suboptimalBlock := individual.GetGenome()[index : index+blockLength]

        // 3. Search for an optimal pattern in the library.
        if pattern := g.PatternLibrary.FindPatternByCodons(suboptimalBlock); pattern != nil {

            // 4. If the pattern has a relevant semantic tag, prioritize its use.
            if pattern.SemanticTag == "vowel_transition" || pattern.SemanticTag == "consonant_transition" {
                optimalBlock := pattern.CodonBlock
                individual.SetGenome(g.ReplaceSuboptimalCodonBlock(individual.GetGenome(), index, blockLength, optimalBlock))
                continue  // Move to the next block
            }

        }

        // 5. Otherwise, use the classic correction.
        optimalBlocks := g.GetOptimalCodonBlocks(population, 0.9, blockLength, 2)

        if len(optimalBlocks) > 0 {
            optimalBlock := optimalBlocks[0]
            individual.SetGenome(g.ReplaceSuboptimalCodonBlock(individual.GetGenome(), index, blockLength, optimalBlock))
        }

    }

    // 6. Regenerate the phenotype.
    // if err := individual.GeneratePhenotype(g); err != nil {
    //     return err
    // }

    return nil
}

// CorrectByOptimalPaths corrects an individual's grammatical paths using 
// optimal paths based on successful productions.
func (g *Genomizer) CorrectByOptimalPaths(ind *Individual, fitnessThreshold float64) error {
    
    // Saves the initial state for restoration in case of failure.
    oldFitness := ind.GetFitness()
    oldPhenotype := ind.GetPhenotype()
    oldGenome := make([]int, len(ind.GetGenome()))
    copy(oldGenome, ind.GetGenome())
    oldHistory := utils.DeepCopyProductionHistory(ind.GetProductionHistory())

    // 1. Identify suboptimal productions in the production history.
    suboptimalIndices := make([]int, 0)

    for i, production := range ind.GetProductionHistory() {
        avgFitness := g.GetAverageFitness(production)
    
        if avgFitness < fitnessThreshold {
            suboptimalIndices = append(suboptimalIndices, i)
        }
    
    }

    // 2. Replace suboptimal productions with optimal productions.
    for _, index := range suboptimalIndices {
    
        if index >= len(ind.GetProductionHistory()) {
            continue
        }

        suboptimalProduction := ind.GetProductionHistory()[index]
        similarProductions := g.FindSimilarProductions(suboptimalProduction, fitnessThreshold)

        if len(similarProductions) > 0 {
            bestSimilar := g.SelectBestProductionWithBias(similarProductions, fitnessThreshold)
    
            if bestSimilar != nil {
                ind.SetProductionStep(index, bestSimilar)
            }
    
        }
    
    }

    // 3. Reconstruct the genome from the corrected production history.
    if err := g.SpliceGenomeFromHistory(ind); err != nil {
    
        // Restore the initial state in case of an error.
        ind.SetFitness(oldFitness)
        ind.SetPhenotype(oldPhenotype)
        ind.SetGenome(oldGenome)
        ind.SetProductionHistory(oldHistory)
    
        return fmt.Errorf("failed to rebuild genome after production correction: %v", err)
    }

    // 4. Regenerate the phenotype and production history.
    if err := ind.GeneratePhenotype(g); err != nil {
    
        // Restore the initial state in case of an error.
        ind.SetFitness(oldFitness)
        ind.SetPhenotype(oldPhenotype)
        ind.SetGenome(oldGenome)
        ind.SetProductionHistory(oldHistory)
    
        return fmt.Errorf("failed to regenerate phenotype: %w", err)
    }

    return nil
}

// CorrectByProductions corrects the genome using individual productions 
// (Refining). It replaces productions from productionHistory with low 
// fitness with productions from successfulProductions: 
// 
// An individual generates "g_xden" (fitness = 0.4). productionHistory 
// contains:
//
// [
//    ["string" → "letter" "string"],  // prod1
//    ["letter" → "'g'"],              // prod2 (good)
//    ["string" → "'_'" "string"],     // prod3 (bad)
//    ["letter" → "'x'"],              // prod4 (invalid)
//    ...
// ]
//
// Fitness is low (0.4), so no productions are added to successfulProductions. 
// CorrectGenome() identifies that prod3 and prod4 have a fitness < 0.5. It 
// replaces them with similar productions from successfulProductions 
// 
//      (eg: ["string" → "letter" "string"] for prod3). 
//
// The corrected phenotype is "gol_den" (fitness = 0.7). productionHistory is 
// updated, and successful productions are added to successfulProductions. A 
// production block is a sequence of grammatical rules 
// 
// (e.g., string → letter letter followed by letter → consonant). 
// 
// These blocks are identified as successful because they appear frequently in 
// high-performing individuals.
func (g *Genomizer) CorrectByProductions(
    individual *Individual, 
    fitnessThreshold, 
    averageFitness float64,
) error {

    // 1. Copy history to edit it.
    historyCopy := utils.DeepCopyProductionHistory(individual.GetProductionHistory())

    // 2. Saving the initial state for restoration in case of error.
    oldFitness := individual.GetFitness()
    oldPhenotype := individual.GetPhenotype()
    oldGenome := make([]int, len(individual.GetGenome()))
    copy(oldGenome, individual.GetGenome())
    oldHistory := utils.DeepCopyProductionHistory(individual.GetProductionHistory())

    // 3. Flag to track if any changes have been made.
    modified := false

    // 4. Browse the productions to apply the corrections.
    for i, production := range historyCopy {
        avgFitness := g.GetAverageFitness(production)

        if avgFitness < fitnessThreshold {

            // Find a similar but more efficient production.
            similarProductions := g.FindSimilarProductions(production, averageFitness)
        
            if len(similarProductions) > 0 {
                bestSimilar := g.SelectBestProductionWithBias(similarProductions, avgFitness)
        
                if bestSimilar != nil {

                    // Replace production in historyCopy.
                    historyCopy[i] = bestSimilar
                    modified = true
                }
        
            }
        
        }

    }

    // 5. Update productionHistory and rebuild the genome if any changes were 
    //    made.
    if modified {
        individual.SetProductionHistory(historyCopy)

        if err := g.SpliceGenomeFromHistory(individual); err != nil {
           
            // Restore the initial state in case of an error.
            individual.SetFitness(oldFitness)
            individual.SetPhenotype(oldPhenotype)
            individual.SetGenome(oldGenome)
            individual.SetProductionHistory(oldHistory)

            return fmt.Errorf("failed to rebuild genome after production correction: %v", err)
        }

    }

    return nil
}

// CorrectByTemplate corrects an individual using generic templates based on 
// the C/V structure. It applies the corrections to the phenotype, validates 
// the result, and reconstructs the genome.
func (g *Genomizer) CorrectByTemplate(
    ind IIndividual,
    templateFunction TemplateFunc,
    fitnessFunction FitnessFunc,
) (bool, error) {

    // 1. Convert the individual and population parameters.
    concreteInd, ok := ind.(*Individual)
    
	if !ok {
        return false, fmt.Errorf("individual is not *Individual")
    }

    // -- Debug -- Initial log: phenotype state BEFORE correction.
    // log.Printf(
    //     "CorrectByTemplate: BEFORE: phenotype: %q, len: %d",
    //     concreteInd.GetPhenotype(), len(concreteInd.GetPhenotype().(string)),
    // )

    // 2. Save the initial state for restoration in case of failure.
    oldFitness := concreteInd.GetFitness()
    oldPhenotype := concreteInd.GetPhenotype()
    oldGenome := make([]int, len(concreteInd.GetGenome()))
    copy(oldGenome, concreteInd.GetGenome())
    oldHistory := utils.DeepCopyProductionHistory(concreteInd.GetProductionHistory())

    // Store the fitness of the productions before correction.
    for _, production := range concreteInd.GetProductionHistory() {
        key := fmt.Sprintf("%v", production)
        concreteInd.SetOldProductionFitness(key, g.GetAverageFitness(production))
    }

    // 3. Apply the correction using a template.
    if !templateFunction(concreteInd) {

        // -- Debug --
        // log.Printf("CorrectByTemplate: templateFunction returned FALSE (phenotype not updated)\n")
        
        return false, nil
    }

    // -- Debug --
    // log.Printf("CorrectByTemplate: phenotype after template function: %v\n, len: %d\n", 
    //     concreteInd.GetPhenotype(),
    //     len(concreteInd.GetPhenotype().(string)),
    // )

    // 4. Verify that the phenotype has been updated.
    if concreteInd.GetPhenotype() == oldPhenotype {

        // -- Warning --
        // log.Printf("CorrectByTemplate: Phenotype not updated by template function")

        return false, nil
    }

    // 5. Reconstruct the genome from the corrected phenotype.
    if err := g.RebuildGenome(concreteInd, true); err != nil {
        return false, fmt.Errorf("failed to rebuild genome: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByTemplate: phenotype after RebuildGenome: %v\n, len: %d\n", 
    //     concreteInd.GetPhenotype(), len(concreteInd.GetPhenotype().(string)),
    // )

    // 6. Regenerate the phenotype and production history.
    if err := concreteInd.GeneratePhenotype(g); err != nil {

        // Restore the initial state in case of an error.
        concreteInd.SetFitness(oldFitness)
        concreteInd.SetPhenotype(oldPhenotype)
        concreteInd.SetGenome(oldGenome)
        concreteInd.SetProductionHistory(oldHistory)

        return false, fmt.Errorf("failed to regenerate phenotype: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByTemplate: phenotype after Genomize: %v\n, len: %d\n", 
    //     concreteInd.GetPhenotype(),
    //     len(concreteInd.GetPhenotype().(string)),
    // )

    // 7. Recalculate fitness.
    if err := concreteInd.Evaluate(fitnessFunction); err != nil {

        // Restore the initial state in case of an error.
        concreteInd.SetFitness(oldFitness)
        concreteInd.SetPhenotype(oldPhenotype)
        concreteInd.SetGenome(oldGenome)
        concreteInd.SetProductionHistory(oldHistory)
        
        return false, fmt.Errorf("failed to recalculate fitness: %w", err)
    }

    // -- Debug --
    // log.Printf("CorrectByTemplate: phenotype after Evaluate: %v, %v", 
    //     concreteInd.GetPhenotype(), 
    //     concreteInd.GetFitness(),
    // )

    // 8. Check for improvement.
    if concreteInd.GetFitness() > oldFitness {

        // -- Debug --
        // log.Printf("CorrectByTemplate: SUCCESS: fitness improved from %v to %v\n", 
        //     oldFitness, 
        //     concreteInd.GetFitness(),
        // )

        return true, nil
    }

    // 9. Restore the initial state if there is no improvement.
    concreteInd.SetFitness(oldFitness)
    concreteInd.SetPhenotype(oldPhenotype)
    concreteInd.SetGenome(oldGenome)
    concreteInd.SetProductionHistory(oldHistory)

    // -- Debug --
    // log.Printf("CorrectByTemplate: FAILED: fitness not improved (restored old state)\n")

    // 10. No applicable template found.
    return false, nil    
}

// CorrectGenome corrects the genome by combining codon blocks and individual 
// productions.
func (g *Genomizer) CorrectGenome(
    individual *Individual, 
    fitnessThreshold float64, 
    population []*Individual,
    averageFitness float64,
) error {

    // Iterative corrections with dynamic thresholds.
    for i := range 2 {  // i := 0; i < 2; i++
        currentThreshold := fitnessThreshold - 0.05*float64(i)    

        // Corrections by blocks of variable sizes (2, 3, 4 codons).
        for blockLength := 2; blockLength <= 4; blockLength++ {

            if err := g.CorrectByLinguisticPatterns(individual, population, currentThreshold, averageFitness, blockLength); err != nil {
                return err
            }            

        }    

        if err := g.CorrectByProductions(individual, currentThreshold, averageFitness); err != nil {
            return err
        }        

    }

    return nil
}

// -- No longer in use --
// DecodeCodonBlock decodes a block of codons into a sequence of productions.
func (g *Genomizer) DecodeCodonBlock(block []int) ([][]IRuleModel, error) {

    // -- Debug --
    log.Printf("DecodeCodonBlock: Starting decoding for block: %v, startRule: %q", block, g.startRule)

    productions := [][]IRuleModel{}
    currentSymbols := []string{g.startRule}
    codonIndex := 0

    for len(currentSymbols) > 0 && codonIndex < len(block) {
        currentSymbol := currentSymbols[0]
        currentSymbols = currentSymbols[1:]

        // Check if the codon is a STOP_CODON. STOP_CODON = CODON_SIZE = 127. 
        // if block[codonIndex] == STOP_CODON {
        //
        //     log.Printf("DecodeCodonBlock: STOP_CODON encountered. Stopping decoding.")
        //
        //     break
        // }

        // Handling of negative codons (-1 markers).
        if codonIndex < len(block) && block[codonIndex] == -1 {

            // Ignore -1 markers (ncRNA) in the first pass.
            codonIndex++
            
            continue
        }

        // Check if the current symbol has an associated rule (non-terminal).
        rule, exists := g.GetSymbols()[currentSymbol]
        
        if !exists {
            
            // -- Debug --
            log.Printf("DecodeCodonBlock: Terminal symbol %q (no rule, it is not a codon)", currentSymbol)
            
            // Terminal (no associated rule).
            continue
        }

        // Non-terminal (rule exists): check that the codon is valid for this 
        // rule.
        if codonIndex >= len(block) {

            // -- Debug --
            // log.Printf("DecodeCodonBlock: No more codons to process (codonIndex=%d, block length=%d)",
            //     codonIndex, len(block))
            
                break
        }
        
        codon := block[codonIndex]

        // Handle negative codons (should not happen here, but just in case).
        // (in case other negative codons exist).
        if codon < 0 {
            codonIndex++
            continue
        }

        // Apply the modulo to handle "out-of-bounds" codons.
        codon = max(codon % len(rule.GetSymbols()), 0) 

        // Select the corresponding production.
        production := rule.GetSymbols()[codon]
        
        // -- Debug --
        log.Printf("DecodeCodonBlock: Codon %d decoded to production by codon index %d: %v for symbol %q",
            codon, codonIndex, production, currentSymbol)

        // Add production to results.
        productions = append(productions, production)

        // Store the current recursive production (if it contains _tail).
        if HasTailSymbol(production) {
            g.currentRecursiveProduction = production
        } else {
            g.currentRecursiveProduction = nil
        }

        // Add production symbols to currentSymbols (FIFO order).
        for _, symbol := range production {
            currentSymbols = append(currentSymbols, symbol.GetText())
        }

        codonIndex++  // Consumes a codon for this production
    }

    // -- Debug --
    log.Printf("DecodeCodonBlock: Final productions: %v", productions)
    
    return productions, nil
}

func (g *Genomizer) DecodeCodonBlockWithDynamicRules(block []int) ([][]IRuleModel, error) {
    g.usedCodons = 0  // Reset before the countdown
    codonIndex := 0
    currentSymbols := []string{g.startRule}
    derivedSymbols := make(map[string]bool)
    symbolOccurrences := make(map[string][]int)  
    parentCodonIndex := make(map[string]int)     
    productions := [][]IRuleModel{}

    // -- Debug --
    // log.Printf("DecodeCodonBlockWithDynamicRules: Starting decoding for block: %v, startRule: %q, usedCodons: %d",
    //     block, g.startRule, g.usedCodons)

    for len(currentSymbols) > 0 && codonIndex < len(block) {
        currentSymbol := currentSymbols[0]
        currentSymbols = currentSymbols[1:]

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Processing symbol: %q, codonIndex: %d, usedCodons: %d, remaining symbols %v",
        //     currentSymbol, codonIndex, g.usedCodons, currentSymbols)

        // CHECK IF THE CURRENT CODON IS -1 (absolute priority).
        if codonIndex < len(block) && block[codonIndex] == -1 {

            if len(g.dynamicRuleStack) == 0 {
                return nil, fmt.Errorf("no non-coding RNA in stack for codon -1 at position %d", codonIndex)
            }

            // Retrieve the dynamic rule from the stack.
            ruleName := g.dynamicRuleStack[len(g.dynamicRuleStack)-1]

            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: Using non-coding RNA: %s (stack after pop: %v)",
            //     ruleName, g.dynamicRuleStack)

            rule, exists := g.dynamicRules[ruleName]

            if !exists {
                return nil, fmt.Errorf("non-coding RNA rule %q not found", ruleName)
            }

            // Retrieve RNA production.
            rnaProduction := rule.GetSymbols()[0]
            productions = append(productions, rnaProduction)

            // Remove ALL symbols from the recursive production of currentSymbols.
            for _, sym := range g.currentRecursiveProduction {

                if len(currentSymbols) > 0 && currentSymbols[0] == sym.GetText() {
                    currentSymbols = currentSymbols[1:]

                    // -- Debug --
                    log.Printf("DecodeCodonBlockWithDynamicRules: Current symbols: %v", currentSymbols)
                }

            }

            // Add the RNA symbols to currentSymbols.
            for _, symbol := range rnaProduction {
                currentSymbols = append(currentSymbols, symbol.GetText())

                // -- Debug --
                // log.Printf("DecodeCodonBlockWithDynamicRules: Current symbols: %v", currentSymbols)
            }

            g.usedCodons++
            codonIndex++  // Consumes the -1 marker

            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: Consumed -1 marker (codonIndex now: %d, usedCodons remains: %d)",
            //     codonIndex, g.usedCodons)

            continue  // Move to the next iteration
        }

        // Initialize a counter for non-terminal symbols.
        nonTerminalCount := 0  // Reset to zero for each new one

        // If the symbol is marked as derived, use the index stored in derivedSymbols.
        if derivedSymbols[currentSymbol] {

            if g.currentRecursiveProduction != nil {

                parentIndex, parentExists := parentCodonIndex[currentSymbol]

                if parentExists && parentIndex+1 < len(block) && block[parentIndex+1] == -1 {

                    // -- Debug --
                    // log.Printf(
                    //     "DecodeCodonBlockWithDynamicRules: SKIPPING derived symbol %q (recursive production in progress and codon=-1 at parent index %d + 1)",
                    //     currentSymbol,
                    //     parentIndex,
                    // )

                    // Filter the current Symbols to retain only those not
                    // derived from recursive production.
                    newSymbols := []string{}

                    for _, sym := range currentSymbols {

                        if !derivedSymbols[sym] {
                            newSymbols = append(newSymbols, sym)
                        }
                    
                    }

                    currentSymbols = newSymbols

                    // Add "__MARKER__" only if currentSymbols is empty.
                    if len(currentSymbols) == 0 {
                        currentSymbols = append(currentSymbols, "__MARKER__")  // Dummy symbol to force the loop to continue
                    }

                    // Restore codonIndex to the parent's index + 1.
                    codonIndex = parentIndex + 1

                    // log.Printf(
                    //     "DecodeCodonBlockWithDynamicRules: Restored codonIndex to %d (parent index %d + 1) for %q",
                    //     codonIndex,
                    //     parentIndex,
                    //     currentSymbol,
                    // )

                    // Remove the derived symbol to prevent a re-attempt at processing.
                    delete(derivedSymbols, currentSymbol)
                    delete(symbolOccurrences, currentSymbol)
                    delete(parentCodonIndex, currentSymbol)

                    continue
                }

            }

            // Use the codon following the parent's.
            var nextCodonIndex int
            
            if len(symbolOccurrences[currentSymbol]) > 0 {
                nextCodonIndex = symbolOccurrences[currentSymbol][0]
                symbolOccurrences[currentSymbol] = symbolOccurrences[currentSymbol][1:]  // Removes the used index
            } else {
                return nil, fmt.Errorf("no more codon indices for derived symbol %q", currentSymbol)
            }

            if nextCodonIndex < len(block) {
                codon := block[nextCodonIndex]

                // -- Debug --
                // log.Printf("DecodeCodonBlockWithDynamicRules: Current codon: %d at new relative index %d, from absolute index %d",
                //     codon, nextCodonIndex, codonIndex)
                // log.Printf("DecodeCodonBlockWithDynamicRules: Using codon index %d for derived symbol %q",
                //     nextCodonIndex, currentSymbol)

                // Find the rule for the derived symbol.
                rule, exists := g.GetSymbols()[currentSymbol]
            
                if !exists {
                    continue  // Terminal symbol, ignore
                }

                // Apply the modulo to handle "out-of-bounds" codons.
                codon = max(codon % len(rule.GetSymbols()), 0)
                production := rule.GetSymbols()[codon]

                // Add production to the results.
                productions = append(productions, production)

                // log.Printf("DecodeCodonBlockWithDynamicRules: Codon %d decoded to production by codon index %d: %v for symbol %q",
                //     codon, nextCodonIndex, production, currentSymbol)

                // Increment usedCodons for this derivative production.
                g.usedCodons++  // Increment usedCodons for each decoded production

                // Mark production symbols as derived with their adjusted index
                // ONLY if the production is multi-symbol.
                if len(production) > 1 {
            
                    for rank, symbol := range production {
            
                        // Check if the symbol has an associated rule (non-terminal).
                        if _, symbolExists := g.GetSymbols()[symbol.GetText()]; symbolExists {
                            currentSymbols = append(currentSymbols, symbol.GetText())
                            derivedSymbols[symbol.GetText()] = true
                            symbolOccurrences[symbol.GetText()] = append(
                                symbolOccurrences[symbol.GetText()],
                                nextCodonIndex+rank+1, 
                            )
                            parentCodonIndex[symbol.GetText()] = nextCodonIndex
                        }
            
                    }
            
                } else {

                    // For single-symbol productions, simply add the symbol to currentSymbols.
                    currentSymbols = append(currentSymbols, production[0].GetText())
                }
        
            }

            // Cleans up derivedSymbols only if no occurrences remain.
            if len(symbolOccurrences[currentSymbol]) == 0 {
                delete(derivedSymbols, currentSymbol)  // Remove the marker after processing
                delete(symbolOccurrences, currentSymbol)  // Clears symbolOccurrences
                delete(parentCodonIndex, currentSymbol)
            }

            continue  // Do not increment usedCodons or codonIndex
        }

        // Check if the current symbol has an associated rule (non-terminal).
        rule, exists := g.GetSymbols()[currentSymbol]

        if !exists {

            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: Terminal symbol %q (no rule, it is not a codon)", currentSymbol)
            
            // Terminal (no associated rule).
            continue
        }

        // Non-terminal (rule exists): check that the codon is valid for this rule.
        if codonIndex >= len(block) {
            
            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: No more codons to process (codonIndex=%d, block length=%d)",
            //     codonIndex, len(block))
            
            break
        }

        codon := block[codonIndex]

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Current codon: %d at index %d", codon, codonIndex)

        // Normal case: codon ≥ 0 (coding DNA).
        // Apply modulo to handle "out-of-bounds" codons.
        codon = max(codon % len(rule.GetSymbols()), 0)

        // Select the corresponding production.
        production := rule.GetSymbols()[codon]

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Codon %d decoded to production by codon index %d: %v for symbol %q",
        //     codon, codonIndex, production, currentSymbol)

        // Add production to results.
        productions = append(productions, production)

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Current productions: %v", productions)

        // Store the current recursive production (if it contains _tail).
        if HasTailSymbol(production) {
            g.currentRecursiveProduction = production
        } else {
            g.currentRecursiveProduction = nil
        }

        // Mark production symbols as derived ONLY if it is a multi-symbol production.
        if len(production) > 1 {
            nonTerminalRank := 0
            nextCodonIndex := codonIndex + 1  // Starting index for derived symbols

            // Add production symbols to currentSymbols (FIFO order).
            for _, symbol := range production {

                // Check if the symbol has an associated rule (non-terminal).
                if _, symbolExists := g.GetSymbols()[symbol.GetText()]; symbolExists {
                    currentSymbols = append(currentSymbols, symbol.GetText())
                    derivedSymbols[symbol.GetText()] = true
                    symbolOccurrences[symbol.GetText()] = append(
                        symbolOccurrences[symbol.GetText()],
                        nextCodonIndex+nonTerminalRank,
                    )
                    parentCodonIndex[symbol.GetText()] = codonIndex  // Store the index of the parent codon
                    nonTerminalCount++  // Increment the counter for non-terminals
                    nonTerminalRank++
                }
                
            }

            codonIndex = codonIndex + nonTerminalCount
        } else {

            // For single-symbol productions, simply add the symbol to currentSymbols.
            currentSymbols = append(currentSymbols, production[0].GetText())
        }

        g.usedCodons++  // 1 codon = 1 production
        codonIndex++    // Consumes a codon for this production

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Consumed codon %d (usedCodons now: %d, codonIndex now: %d)",
        //     codon, g.usedCodons, codonIndex)
    }

    // -- Debug --
    // log.Printf("DecodeCodonBlockWithDynamicRules: Finished decoding. Final productions: %v, usedCodons: %d",
    //     productions, g.usedCodons)

    return productions, nil
}

// DecodeCodonBlockWithDynamicRules decodes a block of codons into a 
// sequence of productions. Handle non-coding RNAs (-1 marker) via the 
// g.dynamicRuleStack, without modifying the overall state of the Genomizer. 
// Dynamic rules are read from the stack without being consumed (since they 
// can be reused elsewhere). Like a non-coding RNA that regulates gene 
// expression without being immediately destroyed (it can be reused for 
// multiple transcriptions).
func (g *Genomizer) DecodeCodonBlockWithDynamicRules0(block []int) ([][]IRuleModel, error) {
    g.usedCodons = 0  // Reset before the countdown
    codonIndex := 0
    currentSymbols := []string{g.startRule}
    derivedSymbols := make(map[string]bool)
    symbolOccurrences := make(map[string][]int)  // Stores the relative indices for each occurrence of a symbol
    symbolRank  := make(map[string]int)
    parentCodonIndex := make(map[string]int)  // Derived symbol → parent codon index
    productions := [][]IRuleModel{}  

    // -- Debug --
    log.Printf("DecodeCodonBlockWithDynamicRules: Starting decoding for block: %v, startRule: %q, usedCodons: %d", 
        block, g.startRule, g.usedCodons)

    for len(currentSymbols) > 0 && codonIndex < len(block) {
        currentSymbol := currentSymbols[0]
        currentSymbols = currentSymbols[1:]

        // -- Debug --
        log.Printf("DecodeCodonBlockWithDynamicRules: Processing symbol: %q, codonIndex: %d, usedCodons: %d, remaining symbols %v", 
            currentSymbol, codonIndex, g.usedCodons, currentSymbols)

        // CHECK IF THE CURRENT CODON IS -1 (absolute priority).
        if codonIndex < len(block) && block[codonIndex] == -1 {
           
            if len(g.dynamicRuleStack) == 0 {
                return nil, fmt.Errorf("no non-coding RNA in stack for codon -1 at position %d", codonIndex)
            }

            // Retrieve the dynamic rule from the stack.
            ruleName := g.dynamicRuleStack[len(g.dynamicRuleStack)-1]
           
            // -- Debug --
            log.Printf("DecodeCodonBlockWithDynamicRules: Using non-coding RNA: %s (stack after pop: %v)",
                ruleName, g.dynamicRuleStack)

            rule, exists := g.dynamicRules[ruleName]
           
            if !exists {
                return nil, fmt.Errorf("non-coding RNA rule %q not found", ruleName)
            }

            // Retrieve RNA production.
            rnaProduction := rule.GetSymbols()[0]
            productions = append(productions, rnaProduction)

            // Remove ALL symbols from the recursive production of currentSymbols.
            for _, sym := range g.currentRecursiveProduction {
           
                if len(currentSymbols) > 0 && currentSymbols[0] == sym.GetText() {
                    currentSymbols = currentSymbols[1:]
           
                    // -- Debug --
                    log.Printf("DecodeCodonBlockWithDynamicRules: Current symbols: %v", currentSymbols)
                }
           
            }

            // Add the RNA symbols to currentSymbols.
            for _, symbol := range rnaProduction {
                currentSymbols = append(currentSymbols, symbol.GetText())
           
                // -- Debug --
                log.Printf("DecodeCodonBlockWithDynamicRules: Current symbols: %v", currentSymbols)
            }

            g.usedCodons++
            codonIndex++  // Consumes the -1 marker
           
            // -- Debug --
            log.Printf("DecodeCodonBlockWithDynamicRules: Consumed -1 marker (codonIndex now: %d, usedCodons remains: %d)",
                codonIndex, g.usedCodons)
           
            continue  // Move to the next iteration
        }

        // Initialize a counter for non-terminal symbols.
        nonTerminalCount := 0  // Reset to zero for each new one

        // If the symbol is marked as derived, use the index stored in
        // derivedSymbols.
        if derivedSymbols[currentSymbol] {

            if g.currentRecursiveProduction != nil {

                parentIndex, parentExists := parentCodonIndex[currentSymbol]

                if parentExists && parentIndex + 1 < len(block) && block[parentIndex + 1] == -1 {
            
                    // -- Debug --
                    log.Printf(
                        "DecodeCodonBlockWithDynamicRules: SKIPPING derived symbol %q (recursive production in progress and codon=-1 at parent index %d + 1)",
                        currentSymbol,
                        parentIndex,
                    )
            
                    // Filter the current Symbols to retain only those not 
                    // derived from recursive production.
                    newSymbols := []string{}
        
                    for _, sym := range currentSymbols {
        
                        if !derivedSymbols[sym] {
                            newSymbols = append(newSymbols, sym)
                        }

                    }
        
                    currentSymbols = newSymbols

                    // Add "__MARKER__" only if currentSymbols is empty.
                    if len(currentSymbols) == 0 {
                        currentSymbols = append(currentSymbols, "__MARKER__")  // Dummy symbol to force the loop to continue
                    }  

                    // Restore codonIndex to the parent's index + 1.
                    codonIndex = parentIndex + 1
            
                    log.Printf(
                        "DecodeCodonBlockWithDynamicRules: Restored codonIndex to %d (parent index %d + 1) for %q",
                            codonIndex,
                            parentIndex,
                            currentSymbol,
                    )                
                
                    // Remove the derived symbol to prevent a re-attempt at processing.
                    delete(derivedSymbols, currentSymbol)
                    delete(symbolRank, currentSymbol)
                    delete(parentCodonIndex, currentSymbol)

                    continue
                }

            }
        
            // Use the codon following the parent's.
            nextCodonIndex := symbolRank[currentSymbol] 
        
            if nextCodonIndex < len(block) {
                codon := block[nextCodonIndex]

                // -- Debug --
                log.Printf("DecodeCodonBlockWithDynamicRules: Current codon: %d at new relative index %d, from absolute index %d", 
                    codon, nextCodonIndex, codonIndex)
                log.Printf("DecodeCodonBlockWithDynamicRules: Using codon index %d for derived symbol %q",
                    nextCodonIndex, currentSymbol)

                // Find the rule for the derived symbol.
                rule, exists := g.GetSymbols()[currentSymbol]
            
                if !exists {
                    continue  // Terminal symbol, ignore
                }

                // Apply the modulo to handle "out-of-bounds" codons.
                codon = max(codon % len(rule.GetSymbols()), 0)
                production := rule.GetSymbols()[codon]

                // Add production to the results.
                productions = append(productions, production)
            
                log.Printf("DecodeCodonBlockWithDynamicRules: Codon %d decoded to production by codon index %d: %v for symbol %q",
                    codon, nextCodonIndex, production, currentSymbol)

                // Increment usedCodons for this derivative production.
                g.usedCodons++  // Increment usedCodons for each decoded production
                
                // Mark production symbols as derived with their adjusted index
                // ONLY if the production is multi-symbol.
                if len(production) > 1 {

                    for rank, symbol := range production {

                        // Check if the symbol has an associated rule (non-terminal).
                        if _, symbolExists := g.GetSymbols()[symbol.GetText()]; symbolExists {
                            currentSymbols = append(currentSymbols, symbol.GetText())
                            derivedSymbols[symbol.GetText()] = true
                            // Stocke l'index relatif pour cette occurrence spécifique
                            symbolOccurrences[symbol.GetText()] = append(symbolOccurrences[symbol.GetText()], codonIndex + rank + 1)
                            symbolRank[symbol.GetText()] = nextCodonIndex + rank + 1  // Update with the next available index
                        }

                        // Otherwise, ignore terminal symbols
                    
                    }

                } else {

                    // For single-symbol productions, simply add the symbol to currentSymbols.
                    currentSymbols = append(currentSymbols, production[0].GetText())
                }

            }
        
            delete(derivedSymbols, currentSymbol)  // Remove the marker after processing
            delete(symbolRank, currentSymbol)
            continue  // Do not increment usedCodons or codonIndex
        }

        // Check if the current symbol has an associated rule (non-terminal).
        rule, exists := g.GetSymbols()[currentSymbol]

        if !exists {

            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: Terminal symbol %q (no rule, it is not a codon)", currentSymbol)

            // Terminal (no associated rule).
            continue
        }

        // Non-terminal (rule exists): check that the codon is valid for this 
        // rule.
        if codonIndex >= len(block) {

            // -- Debug --
            // log.Printf("DecodeCodonBlockWithDynamicRules: No more codons to process (codonIndex=%d, block length=%d)",
            //     codonIndex, len(block))

            break
        }

        codon := block[codonIndex]

        // -- Debug --
        log.Printf("DecodeCodonBlockWithDynamicRules: Current codon: %d at index %d", codon, codonIndex)

        // Normal case: codon ≥ 0 (coding DNA).
        // Apply modulo to handle "out-of-bounds" codons.
        codon = max(codon % len(rule.GetSymbols()), 0)

        // Select the corresponding production.
        production := rule.GetSymbols()[codon]

        // -- Debug --
        log.Printf("DecodeCodonBlockWithDynamicRules: Codon %d decoded to production by codon index %d: %v for symbol %q",
            codon, codonIndex, production, currentSymbol)

        // Add production to results.
        productions = append(productions, production)

        // -- Debug --
        log.Printf("DecodeCodonBlockWithDynamicRules: Current productions: %v", productions)

        // Store the current recursive production (if it contains _tail).
        if HasTailSymbol(production) {
            g.currentRecursiveProduction = production
        } else {
            g.currentRecursiveProduction = nil
        }

        // -- Debug --
        // log.Printf("DecodeCodonBlockWithDynamicRules: Current recursive production: %v", g.currentRecursiveProduction)

        // Mark production symbols as derived ONLY if it is a multi-symbol 
        // production.
        if len(production) > 1 {

            nonTerminalRank := 0

            // Add production symbols to currentSymbols (FIFO order).
            for _, symbol := range production {

                // Check if the symbol has an associated rule (non-terminal).
                if _, symbolExists := g.GetSymbols()[symbol.GetText()]; symbolExists {

                    currentSymbols = append(currentSymbols, symbol.GetText())
                    derivedSymbols[symbol.GetText()] = true
                    symbolRank[symbol.GetText()] = codonIndex + nonTerminalRank + 1  // rank starts at 0, so +1 to get the actual rank
                    parentCodonIndex[symbol.GetText()] = codonIndex  // Store the index of the parent codon
                    nonTerminalCount++  // Increment the counter for non-terminals
                    nonTerminalRank++

                    // -- Debug --
                    log.Printf("DecodeCodonBlockWithDynamicRules: Current symbols: %v", currentSymbols)
                }

                // Otherwise, ignore terminal symbols

            }

            codonIndex = codonIndex + nonTerminalCount

        } else {
            
            // For single-symbol productions, simply add the symbol to 
            // currentSymbols.
            currentSymbols = append(currentSymbols, production[0].GetText())
        }

        g.usedCodons++  // 1 codon = 1 production
        codonIndex++    // Consumes a codon for this production
        
        // -- Debug --
        log.Printf("DecodeCodonBlockWithDynamicRules: Consumed codon %d (usedCodons now: %d, codonIndex now: %d)", codon, g.usedCodons, codonIndex)
    }

    // -- Debug --
    log.Printf("DecodeCodonBlockWithDynamicRules: Finished decoding. Final productions: %v, usedCodons: %d", productions, g.usedCodons)
    
    // -- Debug -- Check that the RNA stack isn't empty.
    // log.Printf("DecodeCodonBlockWithDynamicRules: Dynamic rule stack at end: %v", g.dynamicRuleStack)

    return productions, nil
}

// DevelopRecursiveSequences expands sequences containing recursive symbols 
// (like "string" in "string → letter string") into sequences of non-recursive 
// symbols (like "letter letter letter").
func (g *Genomizer) DevelopRecursiveSequences(symbols []IRuleModel, maxDepth int) ([]IRuleModel, error) {
    depth := 0
    changed := true

    // -- Debug -- Display the input symbol sequence.
    // log.Printf("DevelopRecursiveSequences: Starting with the sequence: %v", symbols)

    // Continue developing as long as there are changes and maxDepth is not 
    // reached.
    for changed && depth < maxDepth {
        changed = false
        depth++

        // -- Debug --
        // log.Printf("DevelopRecursiveSequences: Iteration %d, depth %d", depth, depth)

        // Iterate through each symbol in the sequence.
        for i := 0; i < len(symbols); i++ {
            symbol := symbols[i].GetText()

            // -- Debug --
            // log.Printf("DevelopRecursiveSequences: Processing the symbol '%s' at the position %d", symbol, i)

            // Find recursive rules for this symbol.
            for ntSymbol, rule := range g.GetSymbols() {
             
                for _, prod := range rule.GetSymbols() {
             
                    // Check if the rule is recursive (e.g., "string → letter
                    // string"). A rule is recursive if the LHS (ntSymbol) 
                    // appears in the RHS (prod).
                    if ContainsSymbol(prod, ntSymbol) && prod[0].GetText() == symbol {
             
                        // -- Debug --
                        // log.Printf("DevelopRecursiveSequences: Recursive rule found: %s → %v", ntSymbol, prod)
                        // log.Printf("DevelopRecursiveSequences: The symbol '%s' corresponds to the beginning of the rule %s → %v", symbol, ntSymbol, prod)
                    
                        // Check if the current symbol corresponds to the 
                        // first symbol of the RHS.
                        // if prod[0].symbol == symbol {

                        // Replace the recursive symbol with its decomposition. 
                        // Example: "string" → "letter string", 
                        //          becomes ["letter", "string"].

                        // symbols = append(symbols[:i], append([]string{prod[0].symbol, prod[1].symbol}, symbols[i+1:]...)...)
                        newSymbols := make([]IRuleModel, 0)
                        newSymbols = append(newSymbols, symbols[:i]...)

                        for _, s := range prod {
                            newSymbols = append(newSymbols, model.NewRuleModel(s.GetText(), s.GetSymbolType(), nil))
                        }

                        newSymbols = append(newSymbols, symbols[i+1:]...)
                        symbols = newSymbols

                        // -- Debug --            
                        // log.Printf("DevelopRecursiveSequences: New sequence after development: %v", symbols)

                        changed = true
                        break          
                    }

                }

                if changed {
                    break
                }                    
             
            }
             
            if changed {
                break
            }

        }

    }

    if depth >= maxDepth {
        return nil, fmt.Errorf("maximum bypass depth reached (maxDepth=%d)", maxDepth)
    }

    // -- Debug -- Display the expanded symbol sequence.
    // log.Printf("DevelopRecursiveSequences: Final result : %v", symbols)

    return symbols, nil
}

// EncodeProductionHistoryToCodons encodes the production history into 
// codons without handling recursive expansions as dynamic rules. Used 
// for non-optimized grammars.
func (g *Genomizer) EncodeProductionHistoryToCodons(history [][]IRuleModel) []int {
    
    // -- Debug --
    // log.Printf("EncodeProductionHistoryToCodons: Starting encoding for history: %v", history)    
    
    genome := make([]int, 0)

    // Get the initial symbol (e.g., "string").
    initialProduction, err := g.FindInitialProduction()
    initialSymbol := ""

    if err == nil && len(initialProduction) > 0 {
        initialSymbol = strings.Split(initialProduction[0].GetText(), " ")[0]
        
        // -- Debug --
        // log.Printf("EncodeProductionHistoryToCodons: Initial symbol: %q", initialSymbol)
    }

    // for _, production := range history {
    for i, production := range history {

        // -- Debug --
        // log.Printf("EncodeProductionHistoryToCodons: Processing production[%d]: %v", i, production)

        if len(production) == 0 {
            continue
        }

        // Case 1: Production with a single symbol (terminal or non-terminal).
        if len(production) == 1 {
            symbol := production[0]
            symbolText := strings.TrimSuffix(
                strings.TrimSpace(strings.Split(symbol.GetText(), " ")[0]),
                "_tail",
            )
            symbolType := symbol.GetSymbolType()

            // -- Debug --
            // log.Printf("EncodeProductionHistoryToCodons: Single symbol production - Text: %q, Type: %v", symbol.GetText(), symbolType)

            if symbolType == Terminal {

                // -- Debug --
                // log.Printf("EncodeProductionHistoryToCodons: Terminal symbol %q, finding parent...", symbolText)

                // Find the index of the terminal within its parent rule.
                _, _, altIndex, found := g.FindTerminalIndexInParent(symbolText)
                
                if found {

                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: Terminal %q found in parent rule at index %d", symbolText, altIndex)

                    genome = append(genome, altIndex)
                } else {

                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: No parent found for terminal %q, using default codon 0", symbolText)

                    genome = append(genome, 0)
                }

            } else {

                // -- Debug --
                // log.Printf("EncodeProductionHistoryToCodons: Non-terminal symbol %q", symbolText)

                // For a non-terminal, check if it is the start symbol.
                if symbolText == initialSymbol {

                    // -- Debug --
                    log.Printf("EncodeProductionHistoryToCodons: Symbol %q is initial symbol, using codon 0", symbolText)

                    genome = append(genome, 0)
                } else {

                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: Finding parent for non-terminal %q...", symbolText)

                    // Find the parent rule and the index of the alternative.
                    // parentSymbol, _, altIndex, found := g.FindParentForNonTerminal(symbolText)
                    _, _, altIndex, found := g.FindParentForNonTerminal(symbolText)

                    if found {

                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodons: Non-terminal %q found in parent rule %q at index %d",
                        //     symbolText, parentSymbol, altIndex)

                        genome = append(genome, altIndex)
                    } else {

                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodons: No parent found for non-terminal %q, using default codon 0", symbolText)

                        genome = append(genome, 0)
                    }

                }

            }

        } else {  
            
            // Case 2: Production with multiple symbols.
            isRecursiveProduction := HasTailSymbol(production)
            isRecursiveExpansion := !isRecursiveProduction && IsRecursiveExpansion(production, history, i, g)

            // -- Debug --
            // log.Printf("EncodeProductionHistoryToCodons: isRecursiveProduction for %v: %v", production, isRecursiveProduction)

            if isRecursiveProduction {

                // 2.1 Recursive production (ex: [letter 1 string_tail 1]).
                foundParent := false
                
                // for parentSymbol, parentRule := range g.GetSymbols() {
                for _, parentRule := range g.GetSymbols() {
            
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: Checking parent rule %q for recursive production %v", 
                    //     parentSymbol, 
                    //     production,
                    // )
            
                    for altIndex, alt := range parentRule.GetSymbols() {
                        
                        if len(alt) != len(production) {
                            continue
                        }

                       // Compare the basic symbols term by term (ignoring _tail).
                        match := true
                        
                        for j := range alt {
                            altSymbolName := strings.TrimSuffix(
                                strings.TrimSpace(strings.Split(alt[j].GetText(), " ")[0]),
                                "_tail",
                            )
                            prodSymbolName := strings.TrimSuffix(
                                strings.TrimSpace(strings.Split(production[j].GetText(), " ")[0]),
                                "_tail",
                            )

                            // -- Debug --
                            // log.Printf("EncodeProductionHistoryToCodons: Comparing alt[%d] %q (base: %q) with prod[%d] %q (base: %q)",
                            //     i, alt[i].GetText(), 
                            //     altSymbolName, 
                            //     i, 
                            //     production[i].GetText(), 
                            //     prodSymbolName,
                            // )

                            if altSymbolName != prodSymbolName || alt[j].GetSymbolType() != production[j].GetSymbolType() {
                                match = false
                                break
                            }

                        }

                        if match {

                            // -- Debug --
                            // log.Printf("EncodeProductionHistoryToCodons: Production %v matches recursive alternative %v at index %d in rule %q",
                            //     production, 
                            //     alt, 
                            //     altIndex, 
                            //     parentSymbol,
                            // )

                            genome = append(genome, altIndex)
                            foundParent = true
                            break
                        }

                    }

                    if foundParent {
                        break
                    }

                }

                if !foundParent {

                    //  -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: No parent found for recursive production %v, using default codon 0", production)

                    genome = append(genome, 0)
                }
                
            } else if isRecursiveExpansion {
                
                // 2.2 Recursive expansion (e.g., [letter 1 letter 1 ...] 
                // generated by [letter 1 string_tail 1]).

                // -- Debug --
                // log.Printf("EncodeProductionHistoryToCodons: Production %v is a recursive expansion, skipping encoding", production)
                
                continue
            } else {  
                
                // Case 3: Direct production (e.g., [letter 1 letter 1 
                // letter 1] generated by [letters 1])
                foundParent := false
                
                // for parentSymbol, parentRule := range g.GetSymbols() {
                for _, parentRule := range g.GetSymbols() {

                    //  -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodons: Checking parent rule %q for production %v", parentSymbol, production)

                    for altIndex, alt := range parentRule.GetSymbols() {
                
                        if len(alt) != len(production) {
                            continue
                        }

                        // Compare the symbols term by term.
                        match := true
                
                        for j := range alt {
                            altSymbolName := strings.TrimSpace(strings.Split(alt[j].GetText(), " ")[0])
                            prodSymbolName := strings.TrimSpace(strings.Split(production[j].GetText(), " ")[0])

                            if altSymbolName != prodSymbolName || alt[j].GetSymbolType() != production[j].GetSymbolType() {
                                match = false
                                break
                            }

                        }

                        if match {

                            // -- Debug --
                            // log.Printf("EncodeProductionHistoryToCodons: Production %v matches direct alternative %v at index %d in rule %q",
                            //     production, alt, altIndex, parentSymbol)

                            genome = append(genome, altIndex)
                            foundParent = true
                            break
                        }
                
                    }
                
                    if foundParent {
                        break
                    }
                
                }
                
                if !foundParent {
                    genome = append(genome, 0)
                }

            }

        }

    }

    //  -- Debug --
    log.Printf("EncodeProductionHistoryToCodons: Final encoded genome: %v", genome)
    
    return genome
}

// EncodeProductionHistoryToCodonsWithDynamicRules encodes the production 
// history into a codon sequence (genome), by dynamically managing recursive 
// expansions (non-coding RNA or ncRNA) for grammars optimized with _tail 
// symbols.
//
// Technical details:
//   - Uses -1 markers in the genome to flag recursive expansions (ncRNA).
//   - Maintains a dynamic rule stack (g.dynamicRuleStack) to track context 
//     for these expansions.
//   - Resets the stack (g.dynamicRuleStack = []string{}) at the start to 
//     ensure transient behavior, mirroring the biological degradation of 
//     ncRNA.
//
// Biological analogy:
//   - Acts as an "RNA synthesizer": the grammar is DNA, the genome is mRNA + 
//     ncRNA, and -1 markers are ncRNA-like regulators.
//   - Recursive expansions (ncRNA) are not translated into terminal symbols 
//     but regulate the derivation process, similar to how biological ncRNAs 
//     (e.g., miRNA) regulate gene expression without encoding proteins.
func (g *Genomizer) EncodeProductionHistoryToCodonsWithDynamicRules(history [][]IRuleModel) []int {
    genome := make([]int, 0)

    // Get the start symbol (e.g., "grammar" or "string").
    initialProduction, err := g.FindInitialProduction()
    initialSymbol := ""

    if err == nil && len(initialProduction) > 0 {
        initialSymbol = strings.Split(initialProduction[0].GetText(), " ")[0]

        // -- Debug --
        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Initial symbol: %q", initialSymbol)
    } else {

        // -- Debug --
        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: No initial symbol found")
    }

    for i, production := range history {
        
        // -- Debug --
        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Processing production[%d]: %v", i, production)

        if len(production) == 0 {

            // -- Debug --
            // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Skipping empty production at index %d", i)
            
            continue
        }

        // Case 1: Production with a single symbol (terminal or non-terminal).
        if len(production) == 1 {
            symbol := production[0]
            symbolText := strings.TrimSuffix(
                strings.TrimSpace(strings.Split(symbol.GetText(), " ")[0]),
                "_tail",
            )
            symbolType := symbol.GetSymbolType()
            
            // -- Debug --
            // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Single symbol production - Text: %q, Type: %v", symbol.GetText(), symbolType)

            if symbolType == Terminal {
                
                // -- Debug --
                // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Terminal symbol %q, finding parent...", symbolText)
                
                _, _, altIndex, found := g.FindTerminalIndexInParent(symbolText)
                
                if found {
                
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Terminal %q found in parent rule at index %d", symbolText, altIndex)
                    
                    genome = append(genome, altIndex)
                } else {
                    
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: No parent found for terminal %q, using default codon 0", symbolText)
                    
                    genome = append(genome, 0)
                }

            } else {
                
                // -- Debug --
                // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Non-terminal symbol %q", symbolText)
                
                if symbolText == initialSymbol {
                    
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Symbol %q is initial symbol, using codon 0", symbolText)
                    
                    genome = append(genome, 0)
                } else {
          
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Finding parent for non-terminal %q...", symbolText)
                    
                    // -- Debug -- parentSymbol, _, altIndex, found := g.FindParentForNonTerminal(symbolText)
                    _, _, altIndex, found := g.FindParentForNonTerminal(symbolText)
                    
                    if found {
                    
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Non-terminal %q found in parent rule %q at index %d",
                        //     symbolText, parentSymbol, altIndex)
                    
                            genome = append(genome, altIndex)
                    } else {
                    
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: No parent found for non-terminal %q, using default codon 0", symbolText)
                    
                        genome = append(genome, 0)
                    }

                }

            }

        } else {
            
            // Case 2: Production with multiple symbols.
            isRecursiveProduction := HasTailSymbol(production)
            
            // -- Debug --
            // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: isRecursiveProduction for %v: %v", production, isRecursiveProduction)

            if isRecursiveProduction {
                
                // Case 2.1: Recursive production (e.g., [letter 1 string_tail 1]).
                foundParent := false
                
                // -- Debug -- for parentSymbol, parentRule := range g.GetSymbols() {
                for _, parentRule := range g.GetSymbols() {
                
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Checking parent rule %q for recursive production %v", parentSymbol, production)
                
                    for altIndex, alt := range parentRule.GetSymbols() {
                
                        if len(alt) != len(production) {
                            continue
                        }

                        // Compare the base symbols term by term (ignoring _tail).
                        match := true
                        
                        for j := range alt {
                            altSymbolName := strings.TrimSuffix(
                                strings.TrimSpace(strings.Split(alt[j].GetText(), " ")[0]),
                                "_tail",
                            )
                            prodSymbolName := strings.TrimSuffix(
                                strings.TrimSpace(strings.Split(production[j].GetText(), " ")[0]),
                                "_tail",
                            )

                            // -- Debug --
                            // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Comparing alt[%d] %q (base: %q) with prod[%d] %q (base: %q)",
                            //     j, alt[j].GetText(), altSymbolName, j, production[j].GetText(), prodSymbolName)

                            if altSymbolName != prodSymbolName || alt[j].GetSymbolType() != production[j].GetSymbolType() {
                                match = false
                                break
                            }

                        }

                        if match {
                            
                            // -- Debug --
                            // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Production %v matches recursive alternative %v at index %d in rule %q",
                            //     production, alt, altIndex, parentSymbol)
                            
                            genome = append(genome, altIndex)
                            foundParent = true
                            break
                        }

                    }

                    if foundParent {
                        break
                    }

                }

                if !foundParent {
                    
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: No parent found for recursive production %v, using default codon 0", production)
                    
                    genome = append(genome, 0)
                }

            } else {
                
                // Case 2.2: Recursive expansion 
                // (e.g., [letter 1 letter 1 letter 1] 
                //        of [letter 1 string_tail 1]).
                isRecursiveExpansion := false
                
                if i > 0 {
                    prevProduction := history[i-1]

                    if HasTailSymbol(prevProduction) {

                        isRecursiveExpansion = IsReductionOf(prevProduction, production, g)
                        
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: isRecursiveExpansion for %v: %v", production, isRecursiveExpansion)
                    }
                
                }

                if isRecursiveExpansion {  // Non-coding RNA: add a dynamic rule and encode -1
                
                    // Generate a unique name for the dynamic rule.
                    ruleName := GenerateDynamicRuleName(production)
                    
                    if _, exists := g.dynamicRules[ruleName]; !exists {
                        newRule := model.NewRuleModel(
                            ruleName, 
                            NonTerminal, 
                            [][]IRuleModel{production},
                        )
                        g.dynamicRules[ruleName] = newRule
                    
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Added dynamic rule %s -> %v", ruleName, production)
                    }

                    // A non-coding RNA is non-coding by definition, but it is 
                    // used to regulate coding. The -1 marker is non-coding in 
                    // itself (it does not produce a terminal symbol), but it 
                    // triggers a recursive expansion (like a non-coding RNA 
                    // activating a gene).
                    g.dynamicRuleStack = append(g.dynamicRuleStack, ruleName)  // Add ncRNA to stack
                    genome = append(genome, -1)  // Marker for non-coding RNA
                    
                    // -- Debug --
                    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Production %v is a recursive expansion, encoded as dynamic rule %s", production, ruleName)
                } else {

                    // Case 2.3: Direct production. (e.g., [letter 1 letter 1 
                    // letter 1] of [letters 1]).
                    foundParent := false
                    
                    // -- Debug -- for parentSymbol, parentRule := range g.GetSymbols() {
                    for _, parentRule := range g.GetSymbols() {
                    
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Checking parent rule %q for production %v", parentSymbol, production)
                    
                        for altIndex, alt := range parentRule.GetSymbols() {
                            
                            if len(alt) != len(production) {
                                continue
                            }

                            match := true
                            
                            for j := range alt {
                                altSymbolName := strings.TrimSpace(strings.Split(alt[j].GetText(), " ")[0])
                                prodSymbolName := strings.TrimSpace(strings.Split(production[j].GetText(), " ")[0])

                                // -- Debug --
                                // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Comparing alt[%d] %q with prod[%d] %q",
                                //     j, alt[j].GetText(), j, production[j].GetText())

                                if altSymbolName != prodSymbolName || alt[j].GetSymbolType() != production[j].GetSymbolType() {
                                    match = false
                                    break
                                }

                            }

                            if match {
                                
                                // -- Debug --
                                // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Production %v matches direct alternative %v at index %d in rule %q",
                                //     production, alt, altIndex, parentSymbol)
                                
                                genome = append(genome, altIndex)
                                foundParent = true
                                break
                            }

                        }

                        if foundParent {
                            break
                        }

                    }

                    if !foundParent {
                        
                        // -- Debug --
                        // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: No parent found for production %v, using default codon 0", production)
                        
                        genome = append(genome, 0)
                    }

                }

            }

        }

    }

    // -- Debug --
    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Final encoded genome: %v", genome)
    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Dynamic rules: %v", g.dynamicRules)
    // log.Printf("EncodeProductionHistoryToCodonsWithDynamicRules: Dynamic rule stack at end: %v", g.dynamicRuleStack)
    // log.Fatal("EncodeProductionHistoryToCodonsWithDynamicRules: Forced stop at this location")
    
    return genome
}

// EncodeProductionHistoryToGenome encodes a production history into a COMPLETE 
// genome (adjusted to CODON_SIZE).
func (g *Genomizer) EncodeProductionHistoryToGenome(history [][]IRuleModel) []int {

    // 1. Encodes the history into codons.
    codons := g.EncodeProductionHistoryToCodonsWithDynamicRules(history)  // [0, 0, 0, 0, 1]

    // 2. Create a genome of size CODONS_SIZE with random values.
    genome := make([]int, CODONS_SIZE)
    
    for i := 0; i < CODONS_SIZE; i++ {
        genome[i] = rand.Intn(CODONS_SIZE)
    }

    // 3. Copy the encoded codons to the beginning of the genome.
    copy(genome, codons)

    // Adds a STOP codon right after the history.
    // if len(codons) < CODONS_SIZE {
    //     genome[len(codons)] = STOP_CODON  // const STOP_CODON = CODON_SIZE = 127
    // }

    return genome
}

// EncodeProductionHistoryToGenomeSegment encodes a production history into a 
// genome SEGMENT (without size adjustment). Used by SpliceGenomeFromHistory 
// to insert targeted corrections.
func (g *Genomizer) EncodeProductionHistoryToGenomeSegment(history [][]IRuleModel) []int {
    return g.EncodeProductionHistoryToCodonsWithDynamicRules(history)
}

// EvaluateCodonBlockFitness evaluates the fitness of a codon block.
func (g *Genomizer) EvaluateCodonBlockFitness(
    block []int, 
    population []*Individual, 
    averageFitness float64,
) float64 {
    totalFitness := 0.0

    for _, codon := range block {
        totalFitness += g.GetCodonFitness(codon, population, averageFitness)
    }

    return totalFitness / float64(len(block))
}

func (g *Genomizer) GetCodonFitness(
    codon int, 
    population []*Individual, 
    averageFitness float64,
) float64 {
    
    // Retourne une fitness basée sur la fréquence du codon dans la population
    // Exemple : plus un codon est fréquent dans les individus performants, plus sa fitness est élevée
    count := 0

    for _, ind := range population {

        for _, c := range ind.GetGenome() {

            if c == codon {
                count++
            }

        }

    }

    // Normalise par la taille totale des génomes.
    totalCodons := len(population) * len(population[0].GetGenome())
    return float64(count) / float64(totalCodons)
}

// ExplicitFactorizableSequences inserts implicit recursive expansions into 
// history after the recursive production and before the factorizable sequence.
// Returns a new history with the expansions inserted.
func (g *Genomizer) ExplicitFactorizableSequences(history [][]IRuleModel) [][]IRuleModel {
    
    // 1. Identify the sequences to factor out.
    factorizableSeqs := g.IdentifyFactorizableSequences(history)
    
    if len(factorizableSeqs) == 0 {
    
        // -- Debug --
        // log.Printf("ExplicitFactorizableSequences: No factorizable sequences found.")
    
        return history
    }

    // 2. Create a modifiable copy of the history.
    explicitHistory := make([][]IRuleModel, 0, len(history)+len(factorizableSeqs))
    explicitHistory = append(explicitHistory, history...)

    // 3. Iterate through the sequences to be factored and insert their 
    //    expansions. Iterate from the end to the beginning to avoid index 
    //    shifts.
    for i := len(factorizableSeqs) - 1; i >= 0; i-- {
        seq := factorizableSeqs[i]

        // 4. Construct the explicit recursive expansion ([]IRuleModel).
        expansion := g.BuildRecursiveExpansion(seq)

        // 5. Insert the expansion JUST BEFORE the factorable sequence (at seq.StartIndex).
        insertIndex := seq.StartIndex
        explicitHistory = append(
            explicitHistory[:insertIndex],  // Everything preceding the factorable sequence
            append([][]IRuleModel{expansion}, explicitHistory[insertIndex:]...)...,  // Expansion + factorable sequence + remainder
        )

        // -- Debug --
        // log.Printf(
        //     "ExplicitFactorizableSequences: Inserted expansion at index %d for sequence %v",
        //     insertIndex,
        //     seq.Sequence,
        // )
    
    }

    return explicitHistory
}

// ExtractLinguisticPatterns automatically extracts the linguistic patterns.
//
// DEBUG ExtractLinguisticPatterns: Starting extraction with minFitness=0.50, blockLength=4
// DEBUG ExtractLinguisticPatterns: Average fitness of population: 0.75
// DEBUG ExtractLinguisticPatterns: Found 3 recurrent codon blocks
// DEBUG ExtractLinguisticPatterns: Processing block [0 0 0 0] (frequency: 5)
// DEBUG ExtractLinguisticPatterns: Block [0 0 0 0] fitness=0.80
// DEBUG ExtractLinguisticPatterns: Start symbol "string" has index 1
// DEBUG ExtractLinguisticPatterns: Temporary genome for decoding: [1 0 0 0 0]
// DEBUG ExtractLinguisticPatterns: Decoded production block: [[string 1] [syllable 1] [letter 1 letter 1] [a 0]]
// DEBUG ExtractLinguisticPatterns: Adjusted production block (removed start symbol): [[syllable 1] [letter 1 letter 1] [a 0]]
// DEBUG ExtractLinguisticPatterns: Inferred semantic tag: "syllable_pattern"
// DEBUG ExtractLinguisticPatterns: Created pattern: {CodonBlock:[0 0 0 0] ProductionBlock:[[syllable 1] [letter 1 letter 1] [a 0]] Fitness:0.80 Frequency:5 SemanticTag:"syllable_pattern"}
// DEBUG ExtractLinguisticPatterns: Processing block [0 0 0 1] (frequency: 3)
// DEBUG ExtractLinguisticPatterns: Block [0 0 0 1] fitness=0.60
// DEBUG ExtractLinguisticPatterns: Start symbol "string" has index 1
// DEBUG ExtractLinguisticPatterns: Temporary genome for decoding: [1 0 0 0 1]
// DEBUG ExtractLinguisticPatterns: Decoded production block: [[string 1] [syllable 1] [letter 1 letter 1] [a 0] [b 0]]
// DEBUG ExtractLinguisticPatterns: Adjusted production block (removed start symbol): [[syllable 1] [letter 1 letter 1] [a 0] [b 0]]
// DEBUG ExtractLinguisticPatterns: Inferred semantic tag: "syllable_pattern"
// DEBUG ExtractLinguisticPatterns: Created pattern: {CodonBlock:[0 0 0 1] ProductionBlock:[[syllable 1] [letter 1 letter 1] [a 0] [b 0]] Fitness:0.60 Frequency:3 SemanticTag:"syllable_pattern"}
// DEBUG ExtractLinguisticPatterns: Extracted 2 linguistic patterns
func (g *Genomizer) ExtractLinguisticPatterns(
    individuals []*Individual, 
    minFitness float64, 
    blockLength int,
    individual IIndividual,
) []LinguisticPattern {

    // Save the current ncRNAs from Genomizer.
    oldDynamicRules := g.dynamicRules
    defer func() {

        // Restore ncRNAs after use.
        g.dynamicRules = oldDynamicRules
    }()

    // Load ncRNAs from the individual.
    g.dynamicRules = individual.GetDynamicRules()
    g.dynamicRuleStack = individual.GetDynamicRuleStack()

    // Rebuild the stack if it is empty but rules exist:
    // → g.dynamicRuleStack is now populated with the dynamic rules.
    if len(g.dynamicRuleStack) == 0 && len(g.dynamicRules) > 0 {

        ruleNames := make([]string, 0, len(g.dynamicRules))

        for ruleName := range g.dynamicRules {
            ruleNames = append(ruleNames, ruleName)
        }

        sort.Strings(ruleNames)  // Sort in a deterministic order
        g.dynamicRuleStack = ruleNames
    }

    // -- Debug -- State of ncRNAs after loading.
    // log.Printf(
    //     "ExtractLinguisticPatterns [ARNnc]: Loaded from Individual - DynamicRules: %v, DynamicRuleStack: %v",
    //     g.dynamicRules, g.dynamicRuleStack,
    // )

    // -- Debug --
    // log.Printf("ExtractLinguisticPatterns: Starting extraction with minFitness=%.2f, blockLength=%d", minFitness, blockLength)

    // 1. Calculate the average fitness of the population.
    averageFitness := g.CalculateAverageFitness(individuals)

    // -- Debug --
    // log.Printf("ExtractLinguisticPatterns: Average fitness of population: %.2f", averageFitness)

    // 2. Find the recurring codon blocks.
    codonBlocks := g.FindRecurrentCodonBlocks(individuals, minFitness, blockLength)

    // -- Debug --
    // log.Printf("ExtractLinguisticPatterns: Found %d recurrent codon blocks: %v", len(codonBlocks), codonBlocks)

    // 3. Convert each recurring block into a linguistic pattern.
    var patterns []LinguisticPattern

    for blockStr, freq := range codonBlocks {
        
        // Parse the blockStr string in the format "[0 1 121]".
        blockStr = strings.Trim(blockStr, "[]")
        codonStrs := strings.Split(blockStr, " ")
        block := make([]int, len(codonStrs))
        
        for i, codonStr := range codonStrs {
            codon, err := strconv.Atoi(codonStr)
        
            if err != nil {
       
                // -- Error --
                log.Printf("ExtractLinguisticPatterns: Failed to parse codon %q: %v", codonStr, err)
       
                continue
            }
        
            block[i] = codon
        }

        // -- Debug -- Before decoding the block.
        // log.Printf(
        //     "ExtractLinguisticPatterns [ARNnc]: Decoding block %v (frequency: %d) - Current DynamicRules: %v, DynamicRuleStack: %v",
        //     block, freq, g.dynamicRules, g.dynamicRuleStack,
        // )

        // 4. Calculate the average fitness of the block.
        fitness := g.EvaluateCodonBlockFitness(block, individuals, averageFitness)

        // -- Debug --
        // log.Printf("ExtractLinguisticPatterns: Block %v fitness=%.2f", block, fitness)

        // 5. Decode the block into grammatical productions.

        productionBlock, err := g.DecodeCodonBlockWithDynamicRules(block)

        if err != nil {

            // -- Error --
            log.Printf(
                "ExtractLinguisticPatterns [ARNnc]: ERROR decoding block %v: %v - DynamicRules: %v, DynamicRuleStack: %v",
                block, err, g.dynamicRules, g.dynamicRuleStack,
            )
            
            continue
        }        

        // -- Debug -- After decoding.
        // log.Printf(
        //     "ExtractLinguisticPatterns [ARNnc]: Decoded block %v -> %v - DynamicRules: %v, DynamicRuleStack: %v",
        //     block, productionBlock, g.dynamicRules, g.dynamicRuleStack,
        // )

        // 6. Infer an advanced semantic tag.
        semanticTag := g.InferSemanticTag(productionBlock)

        // -- Debug --
        // log.Printf("ExtractLinguisticPatterns: Inferred semantic tag: %q", semanticTag)

        // 7. Create the linguistic pattern.
        pattern := LinguisticPattern{
            CodonBlock:      block,
            ProductionBlock: productionBlock,
            Fitness:         fitness,
            Frequency:       freq,
            SemanticTag:     semanticTag,
        }

        // -- Debug --
        // log.Printf("ExtractLinguisticPatterns: Created pattern: %+v", pattern)

        patterns = append(patterns, pattern)
    }

    // -- Debug --
    // log.Printf("ExtractLinguisticPatterns: Extracted %d linguistic patterns", len(patterns))

    return patterns
}

// FindInitialProduction finds an initial production to begin deriving a 
// phenotype from the grammar.
//
// - length: the length of the phenotype.
// - return: an initial production (right-hand member of the initial rule), 
//   or an error.
func (g *Genomizer) FindInitialProduction() ([]IRuleModel, error) {

    // 1. Check that the initial rule (g.startRule) exists.
    startRule, exists := g.GetSymbols()[g.startRule]
    if !exists {
        return nil, fmt.Errorf("initial rule '%s' not found", g.startRule)
    }

    // 2. Verify that the initial rule has outputs.
    if len(startRule.GetSymbols()) == 0 {
        return nil, fmt.Errorf("The initial rule '%s' has no valid output", g.startRule)
    }

    // 3. Return the first right-hand member of the initial rule, 
    //      (e.g., [string 1]).
    return startRule.GetSymbols()[0], nil
}

// FindMatchingRule checks if a subsequence matches a grammar rule. It uses 
// the cache to check the origin of terminals.
func (g *Genomizer) FindMatchingRule(subSequence []IRuleModel, cache *ReductionChainCache) (string, IRuleModel) {

    // -- Debug -- Initial log: Display the tested sequence.
    // log.Printf("=== FindMatchingRule: Testing subsequence: %v ===", subSequence)

    for ntSymbol, rule := range g.GetSymbols() {

        // -- Debug --
        // log.Printf("FindMatchingRule: Checking rule for ntSymbol: %s", ntSymbol)

        for _, prod := range rule.GetSymbols() {
            
            // -- Debug --
            // log.Printf("FindMatchingRule:   Production: %v (length: %d)", prod, len(prod))

            if len(prod) != len(subSequence) {
                
                // -- Debug --
                // log.Printf("FindMatchingRule:   Skipping (length mismatch: %d != %d)", len(prod), len(subSequence))
                
                continue
            }

            // Determine if the rule is mixed (terminal + non-terminal).
            hasTerminalInProd := false
            hasNonTerminalInProd := false
            
            for _, sym := range prod {
            
                if g.IsTerminal(sym.GetText()) {
                    hasTerminalInProd = true
                } else {
                    hasNonTerminalInProd = true
                }
            
            }
            
            // -- Debug --
            // log.Printf("FindMatchingRule:   Rule type: hasTerminal=%t, hasNonTerminal=%t",
            //     hasTerminalInProd, hasNonTerminalInProd)

            match := true
            
            for i := range prod {
                prodSymbol := prod[i].GetText()
                subSymbol := subSequence[i].GetText()
            
                // -- Debug --
                // log.Printf("FindMatchingRule:     Comparing prodSymbol: %s with subSymbol: %s",
                //     prodSymbol, subSymbol)

                // If the rule is mixed and prodSymbol is a terminal, use prodSymbol.
                if hasTerminalInProd && hasNonTerminalInProd && g.IsTerminal(prodSymbol) {
                    chains := g.GetReductionChain(prodSymbol, cache)
                    
                    // -- Debug --
                    // log.Printf("FindMatchingRule:       Using GetReductionChain(prodSymbol=%s, cache). Chains: %v",
                    //     prodSymbol, chains)
                    
                    if len(chains) == 0 {
                    
                        if prodSymbol != subSymbol {
                        
                            // -- Debug --
                            // log.Printf("FindMatchingRule:       No chains and direct mismatch: %s != %s", prodSymbol, subSymbol)
                        
                            match = false
                            break
                        } else {

                            // -- Debug --
                            // log.Printf("FindMatchingRule:       No chains but direct match: %s == %s", prodSymbol, subSymbol)
                        }

                    } else {
                        found := false
                        
                        for _, chain := range chains {
                        
                            // -- Debug --
                            // log.Printf("FindMatchingRule:         Checking chain: %v", chain)
                        
                            for _, symbolInChain := range chain {
                                
                                if symbolInChain.GetText() == prodSymbol {
                                    found = true
                                
                                    // -- Debug --
                                    // log.Printf("FindMatchingRule:           Found %s in chain", prodSymbol)
                                    
                                    break
                                }

                            }

                            if found {
                                break
                            }

                        }

                        if !found {
                        
                            // -- Debug --
                            // log.Printf("FindMatchingRule:       %s not found in any chain for %s", prodSymbol, prodSymbol)
                        
                            match = false
                            break
                        }

                    }

                } else {
                    
                    // For homogeneous rules, use a direct comparison.
                    if g.IsTerminal(prodSymbol) {
                    
                        // -- Debug --
                        // log.Printf("FindMatchingRule:       Homogeneous rule: direct comparison for terminal %s vs %s",
                        //     prodSymbol, subSymbol)
                        
                        if prodSymbol != subSymbol {
                        
                            // -- Debug --
                            // log.Printf("FindMatchingRule:       Mismatch: %s != %s", prodSymbol, subSymbol)
                            
                            match = false
                            break
                        }

                    } else {
                        
                        // -- Debug --
                        // log.Printf("FindMatchingRule:       Homogeneous rule: direct comparison for non-terminal %s vs %s",
                        //     prodSymbol, subSymbol)
                        
                        if prodSymbol != subSymbol {
                        
                            // -- Debug --
                            // log.Printf("FindMatchingRule:       Mismatch: %s != %s", prodSymbol, subSymbol)
                            
                            match = false
                            break
                        }

                    }

                }

            }

            if match {
                
                // -- Debug --
                // log.Printf("FindMatchingRule:   >>> MATCH FOUND: ntSymbol=%s, rule=%v <<<", ntSymbol, rule)

                return ntSymbol, rule
            } else {
                
                // -- Debug --
                // log.Printf("FindMatchingRule:   No match for production %v", prod)
            }

        }

    }

    // -- Debug --
    // log.Printf("FindMatchingRule: No matching rule found for subsequence: %v", subSequence)
    
    return "", nil
}

// FindNonTerminalForSequence finds a nonterminal that produces exactly the 
// given sequence of terminals. This method is designed to handle:
// - Regular terminals (e.g., "a", "b", "_")
// - The special symbol ε (epsilon) representing an empty production
// - Terminal sequences (e.g., ["a", "b"] for a rule like "syllable = 'a' 'b'")
//
// Priorities:
// 1. Rules with the shortest productions take precedence (e.g., "voxel = 'a'" 
//    before "string_tail = 'a' ...")
// 2. In case of a tie, the order is determined by the order in which the rules
//    were traversed.
// 3. Returns an error if no rule matches the sequence.
func (g *Genomizer) FindNonTerminalForSequence(sequence []string) (IRuleModel, error) {
    symbols := g.GetSymbols()
    var candidates []IRuleModel

    // -- Debug -- Displays the searched terminal sequence
    // log.Printf("FindNonTerminalForSequence: Searching for non-terminals for the sequence %v", sequence)

    // Browse all the grammar rules.
    for ntSymbol, rule := range symbols {

        // -- Debug --
        // log.Printf("[FindNonTerminalForSequence]   Rule analysis: %s → %v", ntSymbol, rule.GetSymbols())

        // Browse all productions of the rule.
        for _, prod := range rule.GetSymbols() {

            // -- Debug --
            // log.Printf("    Production: %v", prod)

            // Check if the production length corresponds to the sequence.
            if len(prod) != len(sequence) {

                // -- Debug --
                // log.Printf("[FindNonTerminalForSequence]       Incompatible length (production: %d, sequence: %d)",
                //    len(prod), len(sequence))

                continue
            }

            // Check if each symbol in the production corresponds to the sequence.
            match := true

            for i, symbol := range prod {
                terminal := sequence[i]

                // Special case for ε (epsilon).
                if symbol.GetText() == "ε" && sequence[i] == "ε" {

                    // -- Debug --
                    // log.Printf("[FindNonTerminalForSequence]       Match ε found at position %d", i)

                    continue  // ε corresponds to ε
                }

                if symbol.GetText() != terminal {

                    // -- Debug --
                    // log.Printf("[FindNonTerminalForSequence]       No match found for '%s' (expected: '%s', found: '%s'))",
                    //    terminal, terminal, symbol.GetText())

                    match = false
                    break
                }

                // -- Debug --
                // log.Printf("[FindNonTerminalForSequence]       Match found for '%s' at position %d",
                //    terminal, i)

            }

            // If all symbols match, add the non-terminal to the candidates.
            if match {

                // -- Debug --
                // log.Printf("[FindNonTerminalForSequence]     Rule '%s' corresponds to the sequence %v", ntSymbol, sequence)

                candidates = append(candidates, model.NewRuleModel(ntSymbol, NonTerminal, nil))

                // -- Debug --
                // log.Printf("[FindNonTerminalForSequence]     Affiche la liste des candidats %v", candidates)
            }

        }

    }

    if len(candidates) == 0 {
        return nil, fmt.Errorf("no non-terminal produces the sequence %v", sequence)
    }

    // Sort candidates by production length (priority given to the shortest 
    // rules).

    // -- Debug --
    // log.Printf("[FindNonTerminalForSequence] Candidates sorted by specific skill: %v", candidates)

    sort.Slice(candidates, func(i, j int) bool {
        ruleI := g.GetSymbols()[candidates[i].GetText()]
        ruleJ := g.GetSymbols()[candidates[j].GetText()]

        // Calculate the average length of the productions for each candidate.
        avgLenI := 0
        
        for _, prod := range ruleI.GetSymbols() {
            avgLenI += len(prod)
        }
        
        avgLenI /= len(ruleI.GetSymbols())

        avgLenJ := 0
        
        for _, prod := range ruleJ.GetSymbols() {
            avgLenJ += len(prod)
        }
        
        avgLenJ /= len(ruleJ.GetSymbols())

        // -- Debug --
        // log.Printf("[FindNonTerminalForSequence]   Comparison: %s (average length: %.2f) vs %s (average length: %.2f)",
        //    candidates[i].GetText(), float64(avgLenI), candidates[j].GetText(), float64(avgLenJ))

        return avgLenI < avgLenJ  // Priority given to rules with shorter productions
    })

    // Return the best candidate found.
    bestCandidate := candidates[0]

    // -- Debug --
    // log.Printf("[FindNonTerminalForSequence] Best candidate found: %s (for the sequence %v)",
    //    bestCandidate.GetText(), sequence)

    return bestCandidate, nil
}

// -- To be deleted --
// FindNonTerminalsForTerminals find the non-terminals associated with a list 
// of terminals or special symbols
func (g *Genomizer) FindNonTerminalsForTerminals(terminals []string) ([]IRuleModel, error) {
    var nonTerminals []IRuleModel
    symbols := g.GetSymbols()

    // -- Debug --
    log.Printf("FindNonTerminalsForTerminals: Recherche des non-terminaux pour les terminaux %v", terminals)
    
    for _, terminal := range terminals {
           
        if terminal == "ε" {

            // -- Debug --
            log.Printf("Terminal spécial 'ε' détecté → ajout de 'empty'")
            
            // Return the non-terminal 'empty' directly for ε.
            nonTerminals = append(nonTerminals, model.NewRuleModel("empty", NonTerminal, nil))
            continue
        }

        // -- Debug --
        log.Printf("Parcourir les règles pour le terminal '%s'", terminal)

        // Browse through all the rules to find those that contain the terminal.
        for ntSymbol, rule := range symbols {

            // -- Debug --
            log.Printf("  Règle '%s' → %v", ntSymbol, rule.GetSymbols())

            // Check if the terminal appears in an output of this rule.
            for _, prod := range rule.Clone().GetSymbols() {

                // -- Debug --
                log.Printf("    Production: %v", prod)
    
                // if len(prod) == 1 && prod[0].GetText() == terminal {
                //    nonTerminals = append(nonTerminals, model.NewRuleModel(ntSymbol, NonTerminal, nil))
                // }

                for _, symbol := range prod {
                
                    if symbol.GetText() == terminal {

                        // -- Debug --
                        log.Printf("    Match trouvé pour '%s' dans la production", terminal)
                
                        // Ajouter le non-terminal parent (éviter les doublons)
                        alreadyAdded := false
                
                        for _, nt := range nonTerminals {
                
                            if nt.GetText() == ntSymbol {
                                alreadyAdded = true
                                break
                            }
                
                        }
                
                        if !alreadyAdded {

                            // -- debug --
                            log.Printf("    Ajout du non-terminal '%s' pour le terminal '%s'", ntSymbol, terminal)

                            nonTerminals = append(nonTerminals, model.NewRuleModel(ntSymbol, NonTerminal, nil))
                        }
                
                        break // Passer à la production suivante une fois le terminal trouvé
                    }

                }
    
            }
    
        }
    
    }

    if len(nonTerminals) == 0 {
        return nil, fmt.Errorf("no rule produces the terminal or special symbol '%s'", terminals[0])
    }

    // -- Debug --
    log.Printf("FindNonTerminalsForTerminals: Non-terminals associated with terminals: %v", nonTerminals)
    
    return nonTerminals, nil
}

// FindParentForNonTerminal returns the parent rule of a non-terminal (including _tail symbols)
// and the index of the alternative containing it.
func (g *Genomizer) FindParentForNonTerminal(nonTerminal string) (parentSymbol string, parentRule IRuleModel, altIndex int, found bool) {
    
    // -- Debug --
    // log.Printf("FindParentForNonTerminal: Searching parent for non-terminal %q", nonTerminal)

    // Handle optimized symbols (e.g., "string_tail" → "string").
    baseNonTerminal := strings.TrimSuffix(nonTerminal, "_tail")

    // -- Debug --
    // log.Printf("FindParentForNonTerminal: Base non-terminal after trimming '_tail': %q", baseNonTerminal)

    for parentSymbol, rule := range g.GetSymbols() {
        
        // Ignore optimized rules (e.g., "string_tail").
        if strings.HasSuffix(parentSymbol, "_tail") {
        
            // -- Debug --
            // log.Printf("FindParentForNonTerminal: Skipping optimized rule %q", parentSymbol)

            continue
        }

        // -- Debug --
        // log.Printf("FindParentForNonTerminal: Checking parent rule %q", parentSymbol)
        
        for altIndex, alt := range rule.GetSymbols() {
            
            // -- Debug --
            // log.Printf("FindParentForNonTerminal: Checking alternative[%d]: %v", altIndex, alt)
            
            for _, symbol := range alt {
            
                // Extract the symbol name (ignore the index).
                symbolName := strings.TrimSpace(strings.Split(symbol.GetText(), " ")[0])

                // -- Debug --
                // log.Printf("FindParentForNonTerminal: Comparing symbol %q (type: %v) with base non-terminal %q",
                //     symbolName, symbol.GetSymbolType(), baseNonTerminal)

                if symbolName == baseNonTerminal && symbol.GetSymbolType() == NonTerminal {
                    
                    // -- Debug --
                    // log.Printf("FindParentForNonTerminal: Found match! Parent rule: %q, Alternative index: %d",
                    //     parentSymbol, altIndex)
                        
                    return parentSymbol, rule, altIndex, true
                }

            }
        
        }
    
    }

    // -- Debug --
    // log.Printf("FindParentForNonTerminal: No parent found for non-terminal %q", nonTerminal)
    
    return "", nil, 0, false
}

// Helper function to find the parent of a terminal.
func (g *Genomizer) FindParentForTerminal(terminal string) (string, IRuleModel, bool) {
    
    for ntSymbol, rule := range g.GetSymbols() {
    
        for _, alt := range rule.GetSymbols() {
    
            if len(alt) == 1 && alt[0].GetText() == terminal && alt[0].GetSymbolType() == Terminal {
                return ntSymbol, rule, true
            }
    
        }
    
    }
    
    return "", nil, false
}

// -- No longer in use --
// FindProductionForTerminal finds a production that generates the specified 
// terminal (e.g., [letter→'g']). It uses successful productions if available, 
// otherwise the grammar rules. 
func (g *Genomizer) FindProductionForTerminal(terminal string) ([]IRuleModel, error) {

    // 1. Search among successful productions.
    for _, sp := range g.successfulProductions {
    
        if len(sp.Production) == 1 && sp.Production[0].GetText() == terminal {
    
            // Find the non-terminal symbol that can generate this terminal.
            for ntSymbol, rule := range g.GetSymbols() {
    
                for _, prod := range rule.GetSymbols() {
    
                    if len(prod) == 1 && prod[0].GetText() == terminal {
                        return []IRuleModel{
                            model.NewRuleModel(ntSymbol, NonTerminal, nil),
                        }, nil
                    }
    
                }
    
            }
    
        }
    
    }

    // 2. If no successful production matches, use grammar.
    for ntSymbol, rule := range g.GetSymbols() {
    
        for _, prod := range rule.GetSymbols() {
    
            if len(prod) == 1 && prod[0].GetText() == terminal {
                return []IRuleModel{
                    model.NewRuleModel(ntSymbol, NonTerminal, nil),
                }, nil
            }
    
        }
    
    }

    // 3. If no production directly generates the terminal, use a default 
    //    production. Find a non-terminal symbol that could generate this 
    //    terminal.
    for ntSymbol := range g.GetSymbols() {
        return []IRuleModel{
            model.NewRuleModel(ntSymbol, NonTerminal, nil),
        }, nil
    }

    // 4. If no output directly generates the terminal, return an error.
    return nil, fmt.Errorf("aucune production ne génère '%s'", terminal)
}

// -- No longer in use --
// A downward traversal of derivations can potentially lead to a combinatorial 
// explosion, especially if the grammar is complex and contains many production
// rules. To avoid this, we can:
//
// 1. Limit the Depth of Recursion:
// Limit the number of bypass levels you explore. For example, you can define a
// maximum bypass depth.
//
// 2. Use Memoization:
// Store the results of the derivations already explored to avoid recalculating
// them.
//
// 3. Prioritize Promising Paths:
// Use heuristics to favor the derivation paths that seem the most promising 
// (for example, those that have already led to high-fitness phenotypes).
//
// 4. Prune the Unpromising Paths:
// Eliminate detours that do not appear promising (e.g., those that have led
// to low fitness phenotypes in the past).
//
// FindProductionSequenceForPhenotype generates a sequence of productions for a
// given phenotype.
func (g *Genomizer) FindProductionSequenceForPhenotype(
    target string, 
    maxDepth int,
) ([][]IRuleModel, error) {

    if target == "" {
        return nil, fmt.Errorf("empty target phenotype")
    }

    var productionSequence [][]IRuleModel
    var terminalBuffer []IRuleModel

    // Find the initial production to reconstruct the production history.
    initialProduction, err := g.FindInitialProduction()

    if err != nil {
        return nil, fmt.Errorf(
            "Unable to find an initial production for the given length: %v", 
            err,
        )
    }

    // Add the initial production to the history.
    productionSequence = append(productionSequence, initialProduction)

    // Extract the symbols from initial Production to initialize current 
    // Symbols.
    currentSymbols := make([]string, len(initialProduction))

    for i, rule := range initialProduction {
        currentSymbols[i] = rule.GetText()
    }

    depth := 0
    index := 0

    for index < len(target) {
        
        if depth >= maxDepth {
            return nil, fmt.Errorf("maximum bypass depth reached")
        }

        // -- Debug --
        log.Printf("FindProductionSequenceForPhenotype: currentSymbols %v", currentSymbols)

        if len(currentSymbols) == 0 {
            return nil, fmt.Errorf("no non-terminal symbol available to generate the phenotype")
        }

        var nextSymbols []string
        
        for _, symbol := range currentSymbols {

            // -- Debug --
            log.Printf("FindProductionSequenceForPhenotype: symbol %v", symbol)

            rule, exists := g.GetSymbols()[symbol]
        
            if !exists {
                return nil, fmt.Errorf("rule for the symbol '%s' not found", symbol)
            }

            // -- Debug --
            log.Printf("FindProductionSequenceForPhenotype: rule rhs %v", rule.GetSymbols())

            // Distinguishing between sequences and alternatives.
            for _, prod := range rule.GetSymbols() {

                // -- Debug --
                log.Printf("FindProductionSequenceForPhenotype: production %v, length %v", prod, len(prod))
        
                if len(prod) > 1 {
        
                    // It's a sequence.
                    productionSequence = append(productionSequence, prod)

                    // -- Debug --
                    log.Printf("FindProductionSequenceForPhenotype: production sequence for prod > 1 %v", productionSequence)
        
                    for _, s := range prod {
                        nextSymbols = append(nextSymbols, s.GetText())

                        // -- Debug --
                        log.Printf("FindProductionSequenceForPhenotype: nextSymbols sequence %v", nextSymbols)
                    }
        
                } else if len(prod) == 1 {

                    // -- Debug --
                    log.Printf("FindProductionSequenceForPhenotype: production sequence for prod = 1a %v", productionSequence)
        
                    if index < len(target) {  // Index check
                        charStr := string(target[index])

                        // -- Debug --
                        log.Printf("FindProductionSequenceForPhenotype: charStr %v", charStr)
        
                        if prod[0].GetText() == charStr {

                            // -- Debug --
                            log.Printf("FindProductionSequenceForPhenotype: production sequence for prod = 1b %v", productionSequence)
        
                            // Add non-terminal production to the history.
                            productionSequence = append(productionSequence, []IRuleModel{
                                model.NewRuleModel(symbol, NonTerminal, nil),
                            })
        
                            // Add the terminal to the buffer.
                            terminalBuffer = append(terminalBuffer, model.NewRuleModel(charStr, Terminal, nil))

                            index++
                        } else {
        
                            // Clear nextSymbols to avoid accumulating alternatives.
                            nextSymbols = []string{}                            

                            // nextSymbols = append(nextSymbols, prod[0].symbol)
                            // Randomly choose an alternative.
                            alternatives := rule.GetSymbols()
                            chosenProd := alternatives[rand.Intn(len(alternatives))]

                            // This is an alternative, adding non-terminal 
                            // production to the history.
                            productionSequence = append(productionSequence, []IRuleModel{
                                model.NewRuleModel(chosenProd[0].GetText(), NonTerminal, nil),
                            })

                            // Add only the symbol of the chosen alternative.
                            nextSymbols = append(nextSymbols, chosenProd[0].GetText())

                            // -- Debug --
                            log.Printf("FindProductionSequenceForPhenotype: nextSymbols alternative for prod = 1 %v", nextSymbols)

                            break  // Breaking free from the loop after choosing an alternative
                        }
        
                    }
        
                }
        
            }
        
        }
        
        currentSymbols = nextSymbols
        depth++
    }

    // Add the terminals to the end of the productionSequence.
    for _, terminal := range terminalBuffer {
        productionSequence = append(productionSequence, []IRuleModel{terminal})
    }

    log.Printf("FindProductionSequenceForPhenotype: productionSequence %v", productionSequence)

    return productionSequence, nil
}

// FindRecurrentCodonBlocks identifies recurring codon sequences in a given
// population.
func (g *Genomizer) FindRecurrentCodonBlocks(
    individuals []*Individual,
    minFitness float64,
    blockLength int,
) map[string]int {

    // -- Debug --
    // log.Printf("FindRecurrentCodonBlocks: Starting with minFitness=%.2f, blockLength=%d", minFitness, blockLength)
    // log.Printf("FindRecurrentCodonBlocks: Processing %d individuals", len(individuals))

    // 1. Initialize a map to count the occurrences of each block.
    blockCounts := make(map[string]int)

    // 2. Scan each individual in the population.
    // -- Debug -- for i, individual := range individuals {
    for _, individual := range individuals {
        genome := individual.GetGenome()
        fitness := individual.GetFitness()
        
        // -- Debug --
        // log.Printf("FindRecurrentCodonBlocks: Processing individual %d (fitness=%.2f, genome length=%d)",
        //     i, fitness, len(genome))

        // 3. Check if the individual has sufficient fitness.
        if fitness < minFitness {
            
            // -- Debug --
            // log.Printf("FindRecurrentCodonBlocks: Skipping individual %d (fitness=%.2f < minFitness=%.2f)",
            //     i, fitness, minFitness)
            
            continue
        }

        // 4. Traverse the genome in blocks of size `blockLength`.
        for j := 0; j <= len(genome)-blockLength; j += blockLength {
            block := genome[j : j+blockLength]
            blockStr := fmt.Sprintf("%v", block)

            // 5. Increment the counter for this block.
            blockCounts[blockStr]++
            
            // -- Debug --
            // log.Printf("FindRecurrentCodonBlocks: Found block %s at index %d in individual %d (count=%d)",
            //     blockStr, j, i, blockCounts[blockStr])
        }
    
    }

    // -- Debug --
    // log.Printf("FindRecurrentCodonBlocks: Found %d unique blocks before filtering", len(blockCounts))

    // 6. Filters blocks that are too rare and too frequent.
    minOccurrences := 2  // Ignore blocks that appear less than 2 times
    maxOccurrences := 10  // Ignore blocks that appear more than 10 times (too common)
    filteredBlocks := make(map[string]int)
    
    for blockStr, count := range blockCounts {
        
        if count >= minOccurrences && count <= maxOccurrences {
            filteredBlocks[blockStr] = count
     
            // -- Debug --
            // log.Printf("FindRecurrentCodonBlocks: Kept block %s (count=%d)", blockStr, count)
        } else {
    
            // -- Debug --
            // log.Printf("FindRecurrentCodonBlocks: Discarded block %s (count=%d, min=%d, max=%d)",
            //     blockStr, count, minOccurrences, maxOccurrences)
        }

    }

    // -- Debug --
    // log.Printf("FindRecurrentCodonBlocks: Returning %d recurrent blocks", len(filteredBlocks))
    
    return filteredBlocks
}

// FindRecursiveRule searches for a recursive rule for a given subsequence. It 
// returns the corresponding non-terminal symbol, the corresponding production,
// and the complete recursive rule.
func (g *Genomizer) FindRecursiveRule(subSequence []IRuleModel) (string, []IRuleModel, []IRuleModel) {
    
    // Go through all the grammar rules.
    for ntSymbol, rule := range g.GetSymbols() {
    
        for _, prod := range rule.GetSymbols() {
    
            // Check if all the symbols of the subsequence are present in the 
            // production.
            symbolFound := true
    
            for _, s := range subSequence {
                found := false
    
                for _, p := range prod {
    
                    if s.GetText() == p.GetText() {
                        found = true
                        break
                    }
    
                }
    
                if !found {
                    symbolFound = false
                    break
                }
    
            }

            if !symbolFound {
                continue
            }

            // Check if the rule is recursive.
            var recursiveRule  []IRuleModel
            isRecursive := false

            for _, s := range prod {
    
                if _, isNonTerminal := g.GetSymbols()[s.GetText()]; isNonTerminal {
    
                    if strings.HasPrefix(s.GetText(), ntSymbol+"_") || s.GetText() == ntSymbol {
                        recursiveRule  = prod
                        isRecursive = true
                        break
                    }
    
                }
    
            }

            if isRecursive {
                return ntSymbol, prod, recursiveRule 
            }
    
        }
    
    }

    return "", nil, nil
}

// -- Not used yet --
// FindSimilarOptimalCodonBlock finds an optimal codon block similar to a suboptimal block.
func (g *Genomizer) FindSimilarOptimalCodonBlock(suboptimalBlock []int, optimalBlocks [][]int) []int {

    // Compare the suboptimal block with the optimal blocks (this comparison 
    // can be based on the similarity of codon sequences or resulting outputs). 
    // For simplicity, we return the first optimal block found (refine as 
    // needed).
    if len(optimalBlocks) > 0 {
        return optimalBlocks[0]
    }

    return nil
}

// FindSimilarProductions finds productions similar to a candidate production, 
// with a dynamic threshold.
func (g *Genomizer) FindSimilarProductions(production []IRuleModel, averageFitness float64) [][]IRuleModel {
    var similarProductions [][]IRuleModel

    // Dynamically adapt the threshold based on the average fitness of the 
    // population.
    dynamicThreshold := 0.7 - (0.3 * (1 - averageFitness))  // Reduces the threshold if the average fitness is low
    similarityThreshold := math.Max(0.4, dynamicThreshold)  // Minimum threshold of 0.4

    for _, sp := range g.successfulProductions {
        similarity := g.ProductionSimilarity(production, sp.Production)

        if similarity >= similarityThreshold && sp.Fitness < 1.0 {
            similarProductions = append(similarProductions, sp.Production)
        }

    }

    return similarProductions
}

// FindTerminalIndexInParent returns the index of a terminal within its parent 
// rule. 
// Example: For "a", returns 0 if vowel ::= a | e | i | ...
func (g *Genomizer) FindTerminalIndexInParent(terminal string) (parentSymbol string, parentRule IRuleModel, altIndex int, found bool) {

    // Find the parent rule of the terminal.
    parentSymbol, parentRule, found = g.FindParentForTerminal(terminal)

    if !found {
        return "", nil, 0, false
    }

    // Find the index of the alternative containing the terminal.
    for altIndex, alt := range parentRule.GetSymbols() {

        if len(alt) == 1 && alt[0].GetText() == terminal && alt[0].GetSymbolType() == Terminal {
            return parentSymbol, parentRule, altIndex, true
        }

    }

    return "", nil, 0, false
}

func (g *Genomizer) FrequencyOfProduction(candidate []IRuleModel) int {

    for _, sp := range g.successfulProductions {
    
		if reflect.DeepEqual(sp.Production, candidate) {
            return sp.Frequency
        }
    
	}
    
	return 0
}

// -- No longer in use --
// Generate produces a phenotype from a genome and a grammar.
// Parameters:
// - output: Generated result (phenotype).
// - usedInput: Index of the codon used in the genome.
// - input: Genome (list of codons).
func (g *Genomizer) Generate(
    output *string,
    usedInput *int,
    usedWraps *int,
    input []int,
) error {
    *usedInput = 0
    *usedWraps = 0
    *output = ""
    wraps := 0
    
    // FIFO queue for the symbols to be processed.
    unexpandedSymbols := utils.NewQueue[string]()
    unexpandedSymbols.Enqueue(g.startRule)

    // Stack for waiting terminals (left erivation).
    pendingTerminals := utils.NewStack[string]()

    var outputSlice []string

    // -- Debug --
    log.Printf("=== GENERATE START ===")
    log.Printf("Input genome: %v (length: %d)", input, len(input))
    log.Printf("Start rule: %q", g.startRule)
    decodedProductions, err := g.DecodeCodonBlock(input)

    if err != nil {

        // -- Error --
        log.Printf("Failed to decode genome (first pass): %v", err)
        
        return err
    }

    // log.Printf("RebuildGenome: Decoding genome=%v -> productions=%v", input, decodedProductions)

	// As long as the wraps loop counter iteration is less than the MAX_WRAPS 
    // constant and there are still unexpanded non-terminal symbols.
    for wraps <= MAX_WRAPS && (!unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty()) {
        
        // Wrap management (bounces on the genome). If these conditions are 
        // met, increment the wrap counter to indicate that a "rebound" must 
        // be made in the selection of productions. This makes it possible 
        // to diversify the generator outputs using different productions to 
        // generate a given character string.
        if *usedInput%len(input) == 0 && *usedInput > 0 && (!unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty()) {
            wraps++
        
            // -- Debug --
            // log.Printf("--- WRAP %d --- (usedInput: %d, remaining symbols: %d, pending terminals: %d)",
            //     wraps, *usedInput, unexpandedSymbols.Size(), pendingTerminals.Size())
        }

        // Priority is given to non-terminal symbols in the queue.
        if !unexpandedSymbols.IsEmpty() {
            currentSymbol := unexpandedSymbols.Dequeue()
            codonIndex := *usedInput % len(input)
            codon := input[codonIndex]

            // -- Debug --
            // log.Printf("Generate: Processing symbol: %q (codonIndex: %d, codon: %d, usedInput: %d)",
            //     currentSymbol, codonIndex, codon, *usedInput)

            // Retrieve the rule associated with the symbol.
            rule, isNonTerminal := g.GetSymbols()[currentSymbol]
            
            // Check if the current symbol is not a codon. If there is no 
            // correspondence in the codons table, it means that the symbol 
            // is a terminal, and that it can be added to the outputslice 
            // result.
            if !isNonTerminal {  
            
                // Exclude the empty symbol "ε" from the phenotype.
                if currentSymbol != "ε" {
                
                    // Terminal: add to the phenotype WITHOUT consuming a codon.
                    outputSlice = append(outputSlice, currentSymbol)
            
                    // -- Debug --
                    // log.Printf("  → Terminal %q added to output. Output so far: %q",
                    //     currentSymbol, strings.Join(outputSlice, ""))
                }
            
                // Release a waiting terminal (if branching to the left).
                if !pendingTerminals.IsEmpty() {
                    terminal := pendingTerminals.Pop()
                   
                    // Check if the terminal has a parent rule.
                    codon := g.GetCodonForTerminal(terminal)
    
                    if codon == -1 {
        
                        // Error: the terminal has no parent rule in the 
                        // grammar.
                        return fmt.Errorf(
                            "terminal %q has no parent rule in grammar. "+
                            "Ensure all terminals are defined in the grammar (e.g., letter = 'a' | 'b')",
                            terminal,
                        )
    
                    }

                    // -- Debug --
                    // log.Printf("  → Terminal %q has valid codon %d in grammar", terminal, codon)
                   
                    unexpandedSymbols.Enqueue(terminal)
                    
                    // -- Debug --
                    // log.Printf("  → Released pending terminal: %q (queue size: %d)",
                    //     terminal, unexpandedSymbols.Size())
                }

                continue
            }

            // Non-terminal: apply the modulo operation to the codon.
            codon = max(codon % len(rule.GetSymbols()), 0)

            // Hybrid selection.
            var selectedProduction []IRuleModel
            randomValue := rand.Float64()

            // Unified selection: Random (1%) / Fitness (69%) / Genome (30%).
            if randomValue < 0.01 {  // 1% random exploration
                selectedProduction = g.SelectProductionByRandom(codon, rule.GetSymbols())
                
                // -- Debug --
                // log.Printf("  → Random exploration: selected production %v", selectedProduction)
            } else if randomValue < 0.7 {  // 69% selection by fitness
                selectedProduction = g.SelectProductionByFitness(codon, rule.GetSymbols())
            
                // -- Debug --
                // log.Printf("  → Exploration: selected production by fitness: %v", selectedProduction)
            } else {  // 30% deterministic
                selectedProduction = g.SelectProductionByGenome(codon, rule.GetSymbols())
            
                // -- Debug --
                // log.Printf("  → Deterministic: selected production by genome: %v", selectedProduction)
            }

            if selectedProduction == nil {
                return fmt.Errorf("no production selected for symbol %q", currentSymbol)
            }
            
            // -- Debug --
            // log.Printf("  → Non-terminal %q: selected production %v (codon: %d, num alternatives: %d)",
            //     currentSymbol, selectedProduction, codon, len(rule.GetSymbols()))

            // Add the production to the history.
            g.productionHistory = append(g.productionHistory, selectedProduction)

            // Retrieve the type of production derivation.
            derivationType := g.GetDerivationType(selectedProduction)
            
            // -- Debug --
            // log.Printf("  → Derivation type: %v", derivationType)

            // Treat according to the type of derivation.
            switch derivationType {
            case LeftDerivation:
            
                // Left-hand derivation: add the symbols in reverse order
                // and store the pending terminals.
                for i := len(selectedProduction) - 1; i >= 0; i-- {
                    symbol := selectedProduction[i]
            
                    if symbol.GetSymbolType() == Terminal {
                        pendingTerminals.Push(symbol.GetText())
            
                        // -- Debug --
                        // log.Printf("  → Pending terminal (left derivation): %q (pending stack size: %d)",
                        //     symbol.GetText(), pendingTerminals.Size())
            
                    } else {
                        unexpandedSymbols.Enqueue(symbol.GetText())
    
                        // -- Debug --
                        // log.Printf("  → Enqueued non-terminal (left derivation): %q (queue size: %d)",
                        //     symbol.GetText(), unexpandedSymbols.Size())
                    }
    
                }
    
            case RightDerivation:
    
                // Right-hand derivative: add the symbols in natural order.
                for _, symbol := range selectedProduction {

                    if symbol.GetSymbolType() == Terminal {

                        // Validate the terminal here (no duplicate).
                        codon := g.GetCodonForTerminal(symbol.GetText())

                        if codon == -1 {
                            return fmt.Errorf(
                                "terminal %q has no parent rule in grammar. "+
                                "Ensure all terminals are defined in the grammar (e.g., letter = 'a' | 'b')",
                                symbol.GetText(),
                            )
                        }

                        // -- Debug --
                        // log.Printf("  → Terminal %q has valid codon %d in grammar", symbol.GetText(), codon)
                    }
                    
                    unexpandedSymbols.Enqueue(symbol.GetText())
    
                    // -- Debug --
                    // log.Printf("  → Enqueued symbol (right derivation): %q (queue size: %d)",
                    //     symbol.GetText(), unexpandedSymbols.Size())
                }

                *usedInput++  // Consume one codon for each symbol in a mixed production
                
                // -- Debug --
                // log.Printf("  → Codon consumed for mixed rule symbol. usedInput: %d", *usedInput)
    
            default:  // HomogeneousDerivation

                // Homogeneous production: add the symbols in natural order.
                for _, symbol := range selectedProduction {
                    unexpandedSymbols.Enqueue(symbol.GetText())

                    // -- Debug --
                    // log.Printf("  → Enqueued symbol (homogeneous): %q (queue size: %d)",
                    //     symbol.GetText(), unexpandedSymbols.Size())
                }

            }

            // Consume ONE codon for non-terminal production.
            *usedInput++

            // -- Debug --
            // log.Printf("  → Codon consumed for production. usedInput: %d", *usedInput)

        } else if !pendingTerminals.IsEmpty() {

            // Process pending terminals (for left branching).
            terminal := pendingTerminals.Pop()

            // Terminal validation.
            codon := g.GetCodonForTerminal(terminal)

            if codon == -1 {
                return fmt.Errorf(
                    "terminal %q has no parent rule in grammar. "+
                    "Ensure all terminals are defined in the grammar (e.g., letter = 'a' | 'b')",
                    terminal,
                )
            }
            
            // -- Debug --
            // log.Printf("  → Terminal %q has valid codon %d in grammar", terminal, codon)

            unexpandedSymbols.Enqueue(terminal)
            
            // -- Debug --
            // log.Printf("Processing pending terminal: %q (queue size: %d)",
            //     terminal, unexpandedSymbols.Size())
        }

    }

    *usedWraps = wraps  // Update usedWraps via pointer.

    // -- Debug --
    // log.Printf("Total wraps: %d", wraps)

    if !unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty() {
        
        // Replace with ε (empty string).
        for !unexpandedSymbols.IsEmpty() {
            unexpandedSymbols.Dequeue()
        }

        for !pendingTerminals.IsEmpty() {
            pendingTerminals.Pop()
        }        

        // Add a warning message (optional).
        // If there are always non-terminal symbols not developed after having 
        // traveled all the symbols, returning an error message specifying that 
        // the generator could not completely develop grammar. If not completely 
        // expanded, return error.
        // log.Printf("Warning: %d unexpanded symbols and %d pending terminals remaining. Forcing termination.",
        //     unexpandedSymbols.Size(), pendingTerminals.Size())

        // Alternative: force the correction of the problems.
        // return fmt.Errorf(
        //     "generation incomplete after %d wraps: %d unexpanded symbols and %d pending terminals remaining",
        //     wraps, unexpandedSymbols.Size(), pendingTerminals.Size(),
        // )
    }

    // If there are no non-terminal symbols not developed after having traveled all the
	// symbols, returning the output result by converting the slice into a character 
    // string and deleting the space between each symbol.
    *output = strings.Join(outputSlice, "")
    
    // -- Debug --
    log.Printf("=== GENERATE END ===")
    log.Printf("Final phenotype: %q", *output)
    log.Printf("Final production history: %v", g.productionHistory)
    log.Printf("Production history vs decode codon block: %v, %v", g.productionHistory, decodedProductions)
    log.Printf("Used codons: %d, Used wraps: %d", *usedInput, *usedWraps)

    return nil
}

func (g *Genomizer) GenerateWithDynamicRules(
    output *string,
    usedInput *int,
    usedWraps *int,
    input []int,
) error {
    *usedInput = 0
    *usedWraps = 0
    *output = ""
    wraps := 0
    indexInput := 0  // Index for traversing input[] (includes -1s)

    unexpandedSymbols := utils.NewQueue[string]()
    unexpandedSymbols.Enqueue(g.startRule)
    pendingTerminals := utils.NewStack[string]()
    var outputSlice []string

    // -- Debug --
    // log.Printf("GenerateWithDynamicRules: Input genome: %v (length: %d)", input, len(input))
    // log.Printf("GenerateWithDynamicRules: Start rule: %q", g.startRule)
    
    // -- Debug -- Check that the RNA stack isn't empty.
    // log.Printf("GenerateWithDynamicRules: Dynamic rule stack at start: %v", g.dynamicRuleStack)

    for wraps <= MAX_WRAPS && (!unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty()) {

        if indexInput%len(input) == 0 && indexInput > 0 && (!unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty()) {
            wraps++
        }

        if !unexpandedSymbols.IsEmpty() {
            currentSymbol := unexpandedSymbols.Dequeue()
            codonIndex := indexInput % len(input)
            codon := input[codonIndex]

            rule, isNonTerminal := g.GetSymbols()[currentSymbol]

            if !isNonTerminal {

                if currentSymbol != "ε" {
                    outputSlice = append(outputSlice, currentSymbol)
                }

                if !pendingTerminals.IsEmpty() {
                    terminal := pendingTerminals.Pop()
                    codon := g.GetCodonForTerminal(terminal)

                    if codon == -1 {
                        return fmt.Errorf("terminal %q has no parent rule in grammar", terminal)
                    }

                    unexpandedSymbols.Enqueue(terminal)
                }

                continue
            }

            // Handling of non-coding RNAs (codon == -1).
            if codon == -1 {

                if len(g.dynamicRuleStack) == 0 {
                    
                    // -- debug --
                    // log.Printf("GenerateWithDynamicRules: Ignoring codon -1 (no non-coding RNA in stack)")
                    
                    indexInput++  // Consume the -1 and continue
                    continue
                }

                // Retrieve the dynamic rule from the stack.
                ruleName := g.dynamicRuleStack[len(g.dynamicRuleStack)-1]

                // ncRNAs are degraded after use; the stack is cleared after 
                // each decoding/generation step.
                g.dynamicRuleStack = g.dynamicRuleStack[:len(g.dynamicRuleStack)-1]  // Remove from stack

                rule, exists := g.dynamicRules[ruleName]
                
                if !exists {
                    return fmt.Errorf("non-coding RNA rule %q not found", ruleName)
                }

                // Retrieve the RNA production.
                selectedProduction := rule.GetSymbols()[0]
                
                // -- Debug --
                // log.Printf("GenerateWithDynamicRules: Using non-coding RNA %s: %v", ruleName, selectedProduction)

                // Remove ALL symbols of the recursive production from 
                // unexpandedSymbols. The recursive production is stored 
                // in g.currentRecursiveProduction.
                if g.currentRecursiveProduction != nil {

                    // Create a list of symbols to remove (in reverse order to 
                    // avoid index problems).
                    symbolsToRemove := make([]string, len(g.currentRecursiveProduction))
                    
                    for i, sym := range g.currentRecursiveProduction {
                        symbolsToRemove[len(symbolsToRemove)-1-i] = sym.GetText()
                    }

                    // Remove each symbol from unexpandedSymbols if it is at 
                    // the head of the queue.
                    for _, symToRemove := range symbolsToRemove {
                    
                        // Replacing Peek with a manual check.
                        if !unexpandedSymbols.IsEmpty() {
                    
                            // Retrieve the first element without removing it.
                            firstElement := unexpandedSymbols.Dequeue()
                    
                            // If the first element matches, do not put it 
                            // back on.
                            if firstElement == symToRemove {
                    
                                // Do not put it back on, as it is the symbol to be removed.
                                continue
                            } else {
                    
                                // Otherwise, put it back on.
                                unexpandedSymbols.Enqueue(firstElement)
                            }
                    
                        }
                    
                    }
                    
                }

                // Add RNA symbols to unexpandedSymbols.
                for _, symbol := range selectedProduction {
                    unexpandedSymbols.Enqueue(symbol.GetText())
                }

                indexInput++  // Consumes the -1 marker
                continue
            }

            // END OF NON-CODING RNA MANAGEMENT.

            // Existing code for codons ≥ 0.
            codon = max(codon % len(rule.GetSymbols()), 0)
            randomValue := rand.Float64()
            var selectedProduction []IRuleModel

            if randomValue < 0.01 {
                selectedProduction = g.SelectProductionByRandom(codon, rule.GetSymbols())
            } else if randomValue < 0.7 {
                selectedProduction = g.SelectProductionByFitness(codon, rule.GetSymbols())
            } else {
                selectedProduction = g.SelectProductionByGenome(codon, rule.GetSymbols())
            }

            if selectedProduction == nil {
                return fmt.Errorf("no production selected for symbol %q", currentSymbol)
            }

            // Store the current recursive production (if it contains _tail).
            if HasTailSymbol(selectedProduction) {
                g.currentRecursiveProduction = selectedProduction
            } else {
                g.currentRecursiveProduction = nil
            }

            g.productionHistory = append(g.productionHistory, selectedProduction)
            derivationType := g.GetDerivationType(selectedProduction)

            switch derivationType {
            case LeftDerivation:

                for i := len(selectedProduction) - 1; i >= 0; i-- {
                    symbol := selectedProduction[i]

                    if symbol.GetSymbolType() == Terminal {
                        pendingTerminals.Push(symbol.GetText())
                    } else {
                        unexpandedSymbols.Enqueue(symbol.GetText())
                    }

                }

            case RightDerivation:

                for _, symbol := range selectedProduction {

                    if symbol.GetSymbolType() == Terminal {
                        codon := g.GetCodonForTerminal(symbol.GetText())

                        if codon == -1 {
                            return fmt.Errorf("terminal %q has no parent rule in grammar", symbol.GetText())
                        }

                    }

                    unexpandedSymbols.Enqueue(symbol.GetText())
                }

                indexInput++

            default:

                for _, symbol := range selectedProduction {
                    unexpandedSymbols.Enqueue(symbol.GetText())
                }

            }

            *usedInput++
            indexInput++
        } else if !pendingTerminals.IsEmpty() {
            terminal := pendingTerminals.Pop()
            codon := g.GetCodonForTerminal(terminal)

            if codon == -1 {
                return fmt.Errorf("terminal %q has no parent rule in grammar", terminal)
            }

            unexpandedSymbols.Enqueue(terminal)
        }
    
    }

    *usedWraps = wraps

    if !unexpandedSymbols.IsEmpty() || !pendingTerminals.IsEmpty() {

        for !unexpandedSymbols.IsEmpty() {
            unexpandedSymbols.Dequeue()
        }

        for !pendingTerminals.IsEmpty() {
            pendingTerminals.Pop()
        }
    
    }

    *output = strings.Join(outputSlice, "")

    // -- Debug --
    // log.Printf("GenerateWithDynamicRules: Final phenotype: %q", *output)

    return nil
}

// GeneratePhenotypeFromProductionSequence generates the phenotype from the 
// production sequence.
func (g *Genomizer) GeneratePhenotypeFromProductionSequence(productionSequence [][]IRuleModel) (string, error) {
    var phenotype strings.Builder

    // -- Debug -- Log initial : Afficher la séquence de production complète.
    // log.Printf("GeneratePhenotypeFromProductionSequence: Starting with productionSequence: %v", productionSequence)

    // Browse the production sequence to extract the terminals.
    // -- Debug -- for i, production := range productionSequence {
    for _, production := range productionSequence {

        // -- Debug -- Log pour chaque production dans la séquence.
        // log.Printf("GeneratePhenotypeFromProductionSequence: Processing production[%d]: %v", i, production)
        
        // -- Debug -- for j, rule := range production {
        for _, rule := range production {

            // -- Debug -- Log pour chaque règle dans la production.
            // log.Printf(
            //     "GeneratePhenotypeFromProductionSequence: Rule[%d.%d] - Text: %q, Type: %v",
            //     i, j, rule.GetText(), rule.GetSymbolType(),
            // )
        
            // Ignore ε when generating the phenotype.
            if rule.GetText() == "ε" {

                // -- Debug --
                // log.Printf("GeneratePhenotypeFromProductionSequence: Skipping ε at position [%d.%d]", i, j)

                continue
            }

            if rule.GetSymbolType() == Terminal {
                phenotype.WriteString(rule.GetText())

                // -- Debug --
                // log.Printf(
                //     "GeneratePhenotypeFromProductionSequence: Added terminal %q to phenotype (current: %q)",
                //     rule.GetText(), phenotype.String(),
                // )

            } else {
            
                // -- Debug --
                // log.Printf(
                //     "GeneratePhenotypeFromProductionSequence: Non-terminal %q ignored (not added to phenotype)",
                //     rule.GetText(),
                // )

            } 
        
        }
    
    }

    finalPhenotype := phenotype.String()    
    
    // -- Debug -- Log final : Phénotype généré.
    // log.Printf("GeneratePhenotypeFromProductionSequence: Final phenotype generated: %q", finalPhenotype)    

    return finalPhenotype, nil
}

// Genomize generates the phenotype.
func (g *Genomizer) Genomize(genome []int, individual IIndividual) error {
    g.mu.Lock()
    defer g.mu.Unlock()

    // Save only dynamicRules (not the stack).
    oldDynamicRules := g.dynamicRules
    defer func() {
        g.dynamicRules = oldDynamicRules
        // Do not restore g.dynamicRuleStack (transient)
    }()

    // Load ncRNAs from the individual.
    g.dynamicRules = individual.GetDynamicRules()
    g.dynamicRuleStack = individual.GetDynamicRuleStack()

    // Rebuild the stack if it is empty but rules exist:
    // → g.dynamicRuleStack is now populated with the dynamic rules.
    if len(g.dynamicRuleStack) == 0 && len(g.dynamicRules) > 0 {
        
        // Extract the rule names.
        ruleNames := make([]string, 0, len(g.dynamicRules))
    
        for ruleName := range g.dynamicRules {
            ruleNames = append(ruleNames, ruleName)
        }
    
        // Sort the names (alphabetically or according to business logic).
        sort.Strings(ruleNames)
        g.dynamicRuleStack = ruleNames
    }

    // -- Debug --
    // log.Printf("=== GENOMIZE START ===")
    // log.Printf("Genomize: Input genome: %v (len=%d)", genome, len(genome))
    // log.Printf("Genomize: Current g.productionHistory: %v (len=%d)", g.productionHistory, len(g.productionHistory))
    // log.Printf("Genomize: Current g.dynamicRules: %v (len=%d)", g.dynamicRules, len(g.dynamicRules))
    // log.Printf("Genomize: Current g.dynamicRuleStack: %v (len=%d)", g.dynamicRuleStack, len(g.dynamicRuleStack))

    // Genome length verification.
    if len(genome) == 0 {

        // -- Warning --
        log.Printf("Genomize: empty genome")

        return fmt.Errorf("genome is empty")
    }

    // Decoding with loaded ncRNAs:
    // → Retrieves the COMPLETE history (with recursive expansions).
    productions, err := g.DecodeCodonBlockWithDynamicRules(genome)

    if err != nil {

        // -- Error --
        log.Printf("Failed to decode genome: %v", err)
        
        return err
    }

    // -- Debug --
    // log.Printf("Genomize: Decoding with loaded ncRNAs: %v (len=%d)", productions, len(productions))

    // Update the history (only at the end).
    g.productionHistory = productions

    // -- Debug --
    // log.Printf("Genomize: Updated g.productionHistory: %v (len=%d)", g.productionHistory, len(g.productionHistory))

    // Generate the phenotype.

    // -- Debug --
    // log.Printf("Genomize: Before GenerateWithDynamicRules - g.phenotype: %q, g.usedCodons: %v, g.usedWraps: %v",
    //     g.phenotype, g.usedCodons, g.usedWraps)
	
    // Reset g.productionHistory BEFORE GenerateWithDynamicRules.
    g.productionHistory = [][]IRuleModel{}

    if err := g.GenerateWithDynamicRules(&g.phenotype, &g.usedCodons, &g.usedWraps, genome); err != nil {
        
        // -- Warning --
        log.Printf("Failed to generate phenotype: %v", err)
		
        return err
	}

    // -- Debug --
    // log.Printf("Genomize: Current g.productionHistory after GenerateWithDynamicRules: %v (len=%d)", 
    //     g.productionHistory, 
    //     len(g.productionHistory),
    // )

    // -- Debug --
    // log.Printf("Genomize: After GenerateWithDynamicRules - g.phenotype: %q, g.usedCodons: %v, g.usedWraps: %v",
    //     g.phenotype, g.usedCodons, g.usedWraps)

    if g.phenotype == "" {

        // -- Warning --
        log.Printf("Genomize: empty phenotype generated from genome %v", genome)
    }

    // After generating the phenotype, empty the stack into the individual.
    individual.ClearDynamicRuleStack()

    // -- Debug --
    // log.Printf("=== GENOMIZE END ===")
    // log.Printf("Genomize: Current genome: %v (len=%d)", genome, len(genome))
    // log.Printf("Genomize: Current g.productionHistory: %v (len=%d)", g.productionHistory, len(g.productionHistory))
    // log.Printf("Genomize: Current g.dynamicRules: %v (len=%d)", g.dynamicRules, len(g.dynamicRules))
    // log.Printf("Genomize: Current g.dynamicRuleStack: %v (len=%d)", g.dynamicRuleStack, len(g.dynamicRuleStack))

	return nil
}

// GenomizeFromPhenotype reconstructs the production history using the given 
// phenotype. This ensures that the reconstructed genome is consistent with 
// the grammar rules used during the initial generation.
func (g *Genomizer) GenomizeFromPhenotype(phenotype any) error {

    // 1. Clear current history.
    g.productionHistory = [][]IRuleModel{}

    // 2. Manage the different types of phenotypes.
    switch p := phenotype.(type) {
    case string:

        // 2.1. Case of character strings.
        return g.GenomizeFromStringPhenotype(p)

    case int, int32, int64, float32, float64:
        
        // 2.2. Case of numbers: convert to a string of characters and process 
        //      each digit.
        numStr := fmt.Sprintf("%v", p)
        
        return g.GenomizeFromStringPhenotype(numStr)

    default:
        return fmt.Errorf("unsupported phenotype type : %T", phenotype)
    }

}

// GenomizeFromStringPhenotype reconstructs the production history from a 
// string, starting with the initial derivation.
func (g *Genomizer) GenomizeFromStringPhenotype(phenotypeStr string) error {

    // Clear current history.
    g.productionHistory = [][]IRuleModel{}

    productionSequence, err := g.RebuildProductionSequence(phenotypeStr, 20)
    
    if err != nil {
        return err
    }
    
    g.productionHistory = productionSequence
    
    // -- Debug -- Check the reconstruction.
    // log.Printf("GenomizeFromStringPhenotype: phenotype=%s, history %v", 
    //      phenotypeStr, 
    //      g.productionHistory,
    // )

    // -- Debug -- Also check the phenotype generated from this history.
    // regeneratedPhenotype, err := g.GeneratePhenotypeFromProductionSequence(g.productionHistory)
    //
    // if err != nil {
    //     return err
    // }
    //
    // log.Printf(
    //     "GenomizeFromStringPhenotype: regenerated phenotype=%s (expected=%s)",
    //     regeneratedPhenotype,
    //     phenotypeStr,
    // )

    return nil
}

// GetBaseSymbol extracts the base name of a symbol 
// (e.g., "y 0" → "y", "letter_tail 1" → "letter").
func GetBaseSymbol(symbolText string) string {
    
    // Remove the "_tail" suffix if present.
    base := strings.TrimSuffix(symbolText, "_tail")
    
    // Extract the first part (before the space or the numeric suffix).
    return strings.TrimSpace(strings.Split(base, " ")[0])
}

// GetCodonForTerminal finds the codon that produces a given terminal in its 
// parent rule. Returns the index of the codon, or -1 if the terminal is not 
// produced by a rule in the grammar.
func (g *Genomizer) GetCodonForTerminal(terminal string) int {

    for _, rule := range g.GetSymbols() {
    
        for codon, production := range rule.GetSymbols() {
    
            if len(production) == 1 && production[0].GetText() == terminal {
                return codon
            }
    
        }
    
    }
    
    return -1  // Terminal not found in the grammar (e.g., static terminal in a mixed production)
}

// GetDerivationType determines the derivation type of a production.
func (g *Genomizer) GetDerivationType(production []IRuleModel) DerivationType {

    if len(production) == 0 {
        return HomogeneousDerivation
    }

    hasTerminal := false
    hasNonTerminal := false

    for _, symbol := range production {
    
        if symbol.GetSymbolType() == Terminal {
            hasTerminal = true
        } else {
            hasNonTerminal = true
        }
    
    }

    if !hasTerminal || !hasNonTerminal {
        return HomogeneousDerivation
    }

    // Check the position of the first non-terminal.
    for i, symbol := range production {

        if symbol.GetSymbolType() == NonTerminal {

            if i == 0 {
                return LeftDerivation
            }

            return RightDerivation
        }

    }

    return RightDerivation
}

// GetOptimalCodonBlocks returns the most recurring codon sequences in a given 
// population.
//
// Parameters:
// - population: The current population of individuals.
// - minFitness: Fitness threshold for considering an individual as efficient.
// - blockLength: Length of the codon blocks to analyze.
// - minFrequency: Minimum frequency for a block to be considered optimal.
func (g *Genomizer) GetOptimalCodonBlocks(
    population []*Individual, 
    minFitness float64, 
    blockLength int, 
    minFrequency int,
) [][]int {
    codonBlocks := g.FindRecurrentCodonBlocks(population, minFitness, blockLength)
    var optimalBlocks [][]int

    for blockStr, freq := range codonBlocks {

        if freq >= minFrequency {
            var block []int
            _, err := fmt.Sscanf(blockStr, "%v", &block)  // Convert string to []int
        
            if err == nil {
                optimalBlocks = append(optimalBlocks, block)
            }
        
        }

    }

    return optimalBlocks
}

// IdentifyFactorizableSequences identifies factorizable sequences in history.
// Returns a list of sequences with their positions and associated rules.
func (g *Genomizer) IdentifyFactorizableSequences(history [][]IRuleModel) []struct {
    Sequence          [][]IRuleModel  // Sequence to factor out (e.g., [[vowel 1] [consonant 1]])
    StartIndex        int             // Start index in history
    EndIndex          int             // End index in history
    FactorizedRule    string          // Associated rule (e.g., "string_tail")
    SequenceLength    int             // Length of the factored sequence
} {
    factorizableSequences := []struct {
        Sequence       [][]IRuleModel
        StartIndex     int
        EndIndex       int
        FactorizedRule string
        SequenceLength int
    }{}

    i := 0
    
    for i < len(history) {
        production := history[i]
    
        if len(production) == 0 {
            i++
            continue
        }

        // Detect if it is an explicit recursive expansion (to be ignored for 
        // factorization).
        isRecursiveExpansion := false
    
        if i > 0 {
            prevProduction := history[i-1]
        
            if HasTailSymbol(prevProduction) {
                isRecursiveExpansion = IsReductionOf(prevProduction, production, g)
            }
        
        }

        // Detect if it is a factorable SEQUENCE following a recursive production.
        if !isRecursiveExpansion && i > 0 && HasTailSymbol(history[i-1]) {
            maxPossibleLength := len(history) - i
        
            for sequenceLength := maxPossibleLength; sequenceLength >= 2; sequenceLength-- {
        
                if i+sequenceLength > len(history) {
                    continue
                }
        
                sequence := history[i : i+sequenceLength]
        
                // Check if the sequence can be factored using a grammar rule.
                for ruleSymbol := range g.GetSymbols() {

                    if g.CanFactorizeAsImplicitRecursiveExpansion(sequence, ruleSymbol) {
                    
                        // -- Debug -- Log of the factorizable sequence.
                        // log.Printf(
                        //     "IdentifyFactorizableSequences: Found factorizable sequence at index %d to %d (length %d): %v. Matched rule: %s",
                        //     i, i+sequenceLength-1, sequenceLength, sequence, ruleSymbol,
                        // )
                    
                        factorizableSequences = append(factorizableSequences, struct {
                            Sequence       [][]IRuleModel
                            StartIndex     int
                            EndIndex       int
                            FactorizedRule string
                            SequenceLength int
                        }{
                            Sequence:       sequence,
                            StartIndex:     i,
                            EndIndex:       i + sequenceLength - 1,
                            FactorizedRule: ruleSymbol,
                            SequenceLength: sequenceLength,
                        })
                        
                        // Moving to the end of the factored sequence.
                        i += sequenceLength - 1
                        break
                    }

                }

            }

        }

        i++
    }

    return factorizableSequences
}

// IdentifySuboptimalCodonBlocks identifies suboptimal codon sequences in a 
// genome.
func (g *Genomizer) IdentifySuboptimalCodonBlocks(
    genome []int, 
    population []*Individual,
    blockLength int, 
    fitnessThreshold float64,
) []int {
    var suboptimalBlocks []int
    averageFitness := g.CalculateAverageFitness(population)

    // Browse the genome in blocks of size blockLength.
    for i := 0; i <= len(genome)-blockLength; i++ {
        block := genome[i : i+blockLength]
        blockFitness := g.EvaluateCodonBlockFitness(block, population, averageFitness)
        
        if blockFitness < fitnessThreshold {
            suboptimalBlocks = append(suboptimalBlocks, i)  // Store the start index of the block
        }
    
    }

    return suboptimalBlocks
}

// InferSemanticTag assigns a semantic tag to a pattern based on its outputs.
func (g *Genomizer) InferSemanticTag(productionBlock [][]IRuleModel) string {

    if len(productionBlock) == 0 {
        return "empty"
    }

    // 1. Analyze the last production of the block.
    lastProduction := productionBlock[len(productionBlock)-1]

    if len(lastProduction) == 0 {
        return "unknown"
    }

    lastSymbol := lastProduction[0].GetText()

    // 2. Tags based on terminal symbols.
    switch lastSymbol {
    case "a", "e", "i", "o", "u", "y":
        return "vowel_transition"
    case "b", "c", "d", "f", "g", "h", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "v", "w", "x", "z":
        return "consonant_transition"
    case "ε":
        return "termination"
    }

    // 3. Tags based on grammatical structure.
    if len(productionBlock) >= 2 {
        firstProduction := productionBlock[0]

        if len(firstProduction) > 0 {
            firstSymbol := firstProduction[0].GetText()

            switch firstSymbol {
            case "string":
                return "string_initiation"
            case "letter":
                return "letter_sequence"
            case "word":
                return "word_construction"
            }
        
        }

    }

    // 4. Tags based on block length.
    switch len(productionBlock) {
    case 1:
        return "single_production"
    case 2:
        return "double_production"
    default:
        return "complex_sequence"
    }

}

// -- Not used yet --
// Initialise productionHistory pour un nouvel individu.
func (g *Genomizer) InitializeProductionHistory(ind *Individual) error {

    if err := g.Genomize(ind.GetGenome(), ind); err != nil {
        return fmt.Errorf("failed to genomize individual: %v", err)
    }
    
    ind.SetPhenotype(g.phenotype)
    ind.SetProductionHistory(utils.DeepCopyProductionHistory(g.productionHistory))
    return nil
}

// -- Not used yet --
// IsFailedProduction checks if a production is known to be inefficient.
func (g *Genomizer) IsFailedProduction(production []IRuleModel) bool {

    for _, fp := range g.failedProductions {
    
		if reflect.DeepEqual(fp.Production, production) && fp.Fitness < 0.3 {
            return true
        }
    
	}
    
	return false
}

// -- Not used yet --
// IsValidReductionChain checks if the first element of the recursive 
// production corresponds to an element of the stored reduction chain.
func (g *Genomizer) IsValidReductionChain(prod []*RuleModel, symbol string, cache *ReductionChainCache) bool {
    
    // Retrieve all reduction strings for the symbol.
    chains := g.GetReductionChain(symbol, cache)

    if len(chains) == 0 {
        return false
    }

    // Check if the first element of the production corresponds to the first 
    // element of at least one of the reduction chains.
    firstElement := prod[0].GetText()

    for _, chain := range chains {

        if len(chain) > 0 && chain[0].GetText() == firstElement  {
            return true
        }

    }

    return false
}

func (g *Genomizer) IsValidReductionRule(ntSymbol string, prod []IRuleModel) bool {

    // Retrieve the grammar rules for ntSymbol.
    rules, exists := g.GetSymbols()[ntSymbol]
    
    if !exists {
        return false
    }

    // Check if prod is a valid production for ntSymbol.
    for _, rule := range rules.GetSymbols() {
    
        if len(rule) == len(prod) {
            match := true
    
            for i := range rule {
    
                if rule[i].GetText() != prod[i].GetText() {
                    match = false
                    break
                }
    
            }
    
            if match {
                return true
            }
    
        }
    
    }
    
    return false
}

// -- Not used yet --
// NumericPhenotypeToGenome converts a numerical phenotype into a genome. Each
// digit of the number is mapped to a codon (e.g., 123 → [1, 2, 3]).
func (g *Genomizer) NumericPhenotypeToGenome(phenotype float64) ([]int, error) {
    str := fmt.Sprintf("%.0f", phenotype)  // Convert to string without decimals
    genome := make([]int, 0, len(str))

    for _, char := range str {

        // Convert each character into a codon (e.g., '1' → 1, '2' → 2, etc.).
        codon := int(char - '0')
        genome = append(genome, codon)
    }

    return genome, nil
}

// IsTerminal checks if a symbol is a terminal.
func (g *Genomizer) IsTerminal(symbol string) bool {
    _, isNonTerminal := g.GetSymbols()[symbol]
    return !isNonTerminal  // If the symbol is not a key in g.symbols, it is a terminal
}

// ProductionSimilarity calculates the similarity between two productions.
func (g *Genomizer) ProductionSimilarity(p1, p2 []IRuleModel) float64 {

	// Convert productions to strings.
	s1 := utils.RuleModelSliceToString(p1)
	s2 := utils.RuleModelSliceToString(p2)

	// Calculate the Levenshtein distance between the two strings.
	distance := utils.LevenshteinDistance(s1, s2)
	maxLen := max(float64(len(s1)), float64(len(s2)))

	if maxLen == 0 {
		return 0.0
	}

	// Return the normalized similarity.
	return 1.0 - float64(distance)/float64(maxLen)
}

// RebuildGenome reconstructs an individual's genome from either:
//   - The phenotype (explicit history via GenomizeFromPhenotype), or
//   - The production history (implicit history, preprocessed to explicit 
//     recursive expansions via ExplicitFactorizableSequences).
//
// In both cases, it ensures individual.history is synchronized with the 
// encoded history. The genome is then generated, and the individual's 
// phenotype and history are updated.
func (g *Genomizer) RebuildGenome(individual *Individual, usePhenotype bool) error {
    
    // Reset the ncRNAs for this new individual.
    g.dynamicRules = make(map[string]IRuleModel)  // ncRNA resetting
    g.dynamicRuleStack = []string{}               // Stack reset

    // -- Debug --
    // log.Printf("RebuildGenome: START - usePhenotype=%t, individual.phenotype=%v, individual.history=%v",
    //     usePhenotype, individual.GetPhenotype(), individual.GetProductionHistory())

    var genome []int
    var history [][]IRuleModel

    if usePhenotype {
    
        // 1. Reconstruct from the phenotype (RebuildGenomeFromPhenotype 
        //    approach).
        if individual.GetPhenotype() == nil || individual.GetPhenotype() == "" {
            return fmt.Errorf("empty or invalid phenotype")
        }

        // Generate productionHistory from the phenotype.
        if err := g.GenomizeFromPhenotype(individual.GetPhenotype()); err != nil {
            return fmt.Errorf("failed to generate production history from phenotype: %w", err)
        }

        // -- Debug --
        // log.Printf("RebuildGenome: After GenomizeFromPhenotype - g.history=%v", g.GetProductionHistory())

        history = g.GetProductionHistory()
    } else {
        
        // 2. Reconstruct from the production history (RebuildGenomeFromHistory
        //    approach).
        if len(individual.GetProductionHistory()) == 0 {
            return fmt.Errorf("empty production history")
        }

        // -- Debug --
        // log.Printf("RebuildGenome: Using implicit production history: %v", g.productionHistory)

        // Preprocess the implicit history to make recursive expansions 
        // explicit.
        history = g.ExplicitFactorizableSequences(g.productionHistory)

        // -- Debug --
        // log.Printf("RebuildGenome: Explicit history after factorization: %v", history)
    }

    // Always update individual.history with the history used for encoding.
    individual.SetProductionHistory(utils.DeepCopyProductionHistory(history))

    // Encode the history (explicit or pre-processed) into a genome.
    genome = g.EncodeProductionHistoryToGenome(history)

    // -- Debug --
    // log.Printf("RebuildGenome: Encoded genome=%v from history=%v", genome, history)

    // Store ncRNAs in the individual (already synthesized by 
    // EncodeProductionHistoryToGenome).
    individual.SetDynamicRules(g.GetDynamicRules())
    individual.SetDynamicRuleStack(g.GetDynamicRuleStack())

    // Update the individual's genome.
    individual.SetGenome(genome)

    // -- Debug --
    // decodedProductions, _ := g.DecodeCodonBlockWithDynamicRules(genome)
    // log.Printf("RebuildGenome: Decoding genome=%v -> productions=%v", genome, decodedProductions)

    // Regenerate the phenotype from the reconstructed genome.
    if err := g.Genomize(individual.GetGenome(), individual); err != nil {
        return fmt.Errorf("failed to regenerate phenotype: %w", err)
    }

    // -- Debug --
    // log.Printf("RebuildGenome: After Genomize - g.phenotype=%v, g.history=%v",
    //     g.GetPhenotype(), g.GetProductionHistory())

    // Update the phenotype and production history.
    individual.SetPhenotype(g.GetPhenotype())
    individual.SetProductionHistory(g.GetProductionHistory())

    // -- Debug --
    // log.Printf("RebuildGenome: FINAL - individual.phenotype=%v, individual.history=%v",
    //     individual.GetPhenotype(), individual.GetProductionHistory())

    return nil
}

// -- Not used yet --
// Reconstruct the entire genome (CODON_SIZE) of an individual from 
// productionHistory. Prerequisite: EncodeProductionHistoryToGenome() must 
// return a genome of size CODON_SIZE.
func (g *Genomizer) RebuildGenomeFromHistory(individual *Individual) error {

    // 1. Encode productionHistory into a genome of size CODON_SIZE.
    genome := g.EncodeProductionHistoryToGenome(individual.GetProductionHistory())

    // 2. Security check (optional, if EncodeProductionHistoryToGenome is not trusted).
    if len(genome) != CODONS_SIZE {
        return fmt.Errorf("genome size is %d, expected %d (EncodeProductionHistoryToGenome failed)", len(genome), CODONS_SIZE)
    }

    // 3. Update the individual's genome.
    individual.SetGenome(genome)

    // 4. Reconstruct the phenotype from the new genome.
    if err := g.Genomize(individual.GetGenome(), individual); err != nil {
        return fmt.Errorf("failed to regenerate phenotype after genome rebuild: %v", err)
    }

    individual.SetPhenotype(g.GetPhenotype())

    return nil
}

// RebuildProductionSequence reconstructs the production sequence for a given 
// phenotype. It updates the cache for terminal and associated non-terminals.
func (g *Genomizer) RebuildProductionSequence(target string, maxDepth int) ([][]IRuleModel, error) {
    
    // -- Debug --
    // log.Printf("RebuildProductionSequence: Beginning with target=%s, maxDepth=%d", target, maxDepth)

    if target == "" {
        return nil, fmt.Errorf("target phenotype is empty")
    }

    var productionSequence [][]IRuleModel

    // 1. Extract the terminals of the phenotype.
    terminals := strings.Split(target, "")

    // -- Debug --
    // log.Printf("RebuildProductionSequence: Terminals extracted: %v", terminals)

    if len(terminals) == 0 {
        return nil, fmt.Errorf("no terminal extracted from the phenotype")
    }

    // 2. Initialize the reduction cache.
    cache := NewReductionChainCache()

    // 3. Find the non-terminals for all sequences of terminals. We are now 
    //    dealing with sequences (not just individual terminals).
    var currentSymbols []IRuleModel
    remainingTerminals := make([]string, len(terminals))
    copy(remainingTerminals, terminals)

    // Find the non-terminals for the terminal sequences.
    for len(remainingTerminals) > 0 {
        found := false

        // Try to find the longest sequence of terminals that corresponds to 
        // a production.
        for seqLength := len(remainingTerminals); seqLength >= 1; seqLength-- {
            seq := remainingTerminals[:seqLength]

            // Find a non-terminal that produces this sequence.
            nonTerminal, err := g.FindNonTerminalForSequence(seq)
            
            if err == nil {
            
                // Add the non-terminal to currentSymbols.
                currentSymbols = append(currentSymbols, nonTerminal)

                // Update the cache for each terminal in the sequence. The 
                // reduction string is: terminal → non-terminal.
                for _, terminal := range seq {
                    chain := []IRuleModel{
                        model.NewRuleModel(terminal, Terminal, nil),
                        nonTerminal,
                    }
                    g.UpdateReductionChain(terminal, chain, cache, false)
                }

                // Update the cache for the non-terminal itself. The reduction 
                // string is: non-terminal → (existing or empty string).
                existingChain := g.GetReductionChain(nonTerminal.GetText(), cache)
    
                if len(existingChain) == 0 {
                    chain := []IRuleModel{nonTerminal}
                
                    for _, terminal := range seq {
                        chain = append(chain, model.NewRuleModel(terminal, Terminal, nil))
                    }
                    
                    g.UpdateReductionChain(nonTerminal.GetText(), chain, cache, false)
                }

                // Remove the processed sequence from remainingTerminals.
                remainingTerminals = remainingTerminals[seqLength:]
                found = true
                break
            }
    
        }

        if !found {
            return nil, fmt.Errorf("no production found for remaining terminals: %v", remainingTerminals)
        }
    
    }

    // 4. Obtain the initial symbol.
    initialProduction, err := g.FindInitialProduction()
    
    if err != nil {
        return nil, fmt.Errorf("failed to find an initial production: %v", err)
    }
    
    initialSymbol := initialProduction[0].GetText()

    // -- Debug --
    // log.Printf("RebuildProductionSequence: Initial symbol: %s", initialSymbol)

    // 5. Develop recursive sequences.
    developedSymbols, err := g.DevelopRecursiveSequences(currentSymbols, maxDepth)
    
    if err != nil {
        return nil, fmt.Errorf("failed to develop recursive sequences: %v", err)
    }

    // 6. Reduce to the initial symbol using the cache.
    _, sequenceProductions, err := g.ReduceToInitialSymbol(developedSymbols, initialSymbol, maxDepth, cache)
    
    if err != nil {
        return nil, fmt.Errorf("failed to reduce to initial symbol: %v", err)
    }

    // 7. Construct the final sequence.
    productionSequence = append(productionSequence, sequenceProductions...)
    productionSequence = append([][]IRuleModel{initialProduction}, productionSequence...)

    // -- Debug --
    // log.Printf("RebuildProductionSequence: After adding initialProduction - productionSequence=%v", productionSequence)

    // 8. Add the terminals at the end.
    for _, terminal := range terminals {
        productionSequence = append(productionSequence, []IRuleModel{
            model.NewRuleModel(terminal, Terminal, nil),
        })
    }

    // -- Debug --
    // log.Printf("RebuildProductionSequence: After adding terminals - productionSequence=%v", productionSequence)

    // -- Debug --
    // log.Printf("RebuildProductionSequence: productionSequence before ε check: %v", productionSequence)

    // 9. Add ε explicitly if present in productionHistory.
    for _, prod := range g.GetProductionHistory() {

        for _, rule := range prod {
        
            if rule.GetText() == "ε" {

                // -- Debug --
                // log.Printf("RebuildProductionSequence: Found ε in productionHistory, adding to productionSequence")
                
                productionSequence = append(productionSequence, []IRuleModel{
                    model.NewRuleModel("ε", Terminal, nil),
                })
                break
            }
    
        }

    }

    // -- Debug --
    // log.Printf("RebuildProductionSequence: productionSequence after ε check: %v", productionSequence)
    // log.Printf("RebuildProductionSequence: Before validation - productionSequence=%v", productionSequence)

    // 10. Validate the sequence.
    if err := g.ValidatePhenotypeFromProductionSequence(productionSequence, target); err != nil {
        return nil, fmt.Errorf("invalid production sequence: %v", err)
    }

    // -- Debug --
    // if len(productionSequence) == 0 {
    //     log.Fatalf(
    //         "RebuildProductionSequence: productionSequence is EMPTY! target=%s, maxDepth=%d, terminals=%v, currentSymbols=%v, developedSymbols=%v, sequenceProductions=%v",
    //         target, maxDepth, terminals, currentSymbols, developedSymbols, sequenceProductions,
    //     )
    // }

    // -- Debug --
    // log.Printf("RebuildProductionSequence: Final result: %v", productionSequence)

    return productionSequence, nil
}

// RebuildProductionSequenceFromPhenotype attempts to reduce the phenotype 
// back to the initial symbol. If the reduction fails (maxDepth reached), 
// it generates a fallback and triggers a stress signal (FALLBACK_MARKER).
// Like a cell that, under stress (e.g., DNA damage), activates heat shock 
// proteins (HSPs) to signal the problem.
func (g *Genomizer) RebuildProductionSequenceFromPhenotype(
    target string, 
    maxDepth int,
    ind IIndividual,
) ([][]IRuleModel, error) {

    // --- Save a DEEP COPY of the current state of the ncRNAs ---
    // Deep copy of dynamicRules (map)
    oldDynamicRules := make(map[string]IRuleModel, len(g.dynamicRules))

    for k, v := range g.dynamicRules {

        // Clone each IRuleModel in the map.
        oldDynamicRules[k] = v.Clone()
    }

    // Restore the initial state at the end of the function.
    defer func() {
        g.dynamicRules = oldDynamicRules
    }()

    // -- Debug --
    // log.Printf("RebuildProductionSequenceFromPhenotype: Starting for target %q (maxDepth: %d)", target, maxDepth)

    // Verification of reducibility.
    if !g.CanReduceToInitialSymbol(target, maxDepth) {

        // -- Debug --
        // log.Printf("RebuildProductionSequenceFromPhenotype: Phenotype %q cannot be reduced to initial symbol. Skipping.", target)

        // --- Update LastValidPhenotype with the target phenotype ---
        if ind != nil {

            concreteInd, ok := ind.(*Individual)
            
            if ok {
                concreteInd.SetLastValidPhenotype(target)  // Save the original phenotype
                concreteInd.SetExhausted(true)  // Mark as exhausted
                
                // -- Debug --
                // log.Printf("RebuildProductionSequenceFromPhenotype: Individual exhausted. Fallback generated.")
                // log.Printf("RebuildProductionSequenceFromPhenotype: Last individual valid phenotype: %v", concreteInd.GetLastValidPhenotype())
            }
        
        }

        // Generate the fallback: [[string] [ε]].
        fallbackSequence := [][]IRuleModel{
            {model.NewRuleModel(g.startRule, NonTerminal, nil)},
            {model.NewRuleModel("ε", Terminal, nil)},
        }

        // Mark the failure in dynamicRuleStack.
        g.dynamicRuleStack = append(g.dynamicRuleStack, "FALLBACK_MARKER")

        // -- Debug --
        // log.Printf("RebuildProductionSequenceFromPhenotype: Fallback sequence generated: %v", fallbackSequence) 

        // --- Store the original phenotype in LastValidPhenotype (in 
        //     CorrectByGrammaticalPaths) ---.
        return fallbackSequence, nil 
    }

    // If the reduction succeeds, reset isExhausted.
    if ind != nil {
        concreteInd, ok := ind.(*Individual)
        
        if ok {
            concreteInd.SetExhausted(false)  // Reset on success
        }

    }

    // Call the internal function for the reduction.
    return g.RebuildProductionSequence(target, maxDepth)
}

// RecentSuccessScore is a measure of a production's recent performance, 
// based on recent generations. It is calculated using a history of the last 
// successful productions, and reflects the recent quality of a production, 
// focusing on the latest generations. It allows for favoring productions that 
// have recently led to high-quality phenotypes, which is particularly useful 
// for quickly adapting the system to recent improvements. recentSuccess allows 
// the system to react quickly to recent improvements. If a production starts 
// to generate better phenotypes, it will quickly be favored. By focusing on 
// recent performance, recentSuccess helps avoid local pitfalls where 
// historically good production might no longer be optimal. New productions 
// that begin to perform well are quickly identified and promoted, which 
// encourages innovation and the exploration of new solutions. 
// The learning curve is improved by allowing the system to react quickly to 
// recent performance and avoid local pitfalls. This promotes faster and more 
// accurate convergence to the target.
func (g *Genomizer) RecentSuccessScore(candidate []IRuleModel) float64 {
    score := 0.0

    for _, sp := range g.recentSuccessfulProductions {
    
		if reflect.DeepEqual(sp.Production, candidate) {
            score += sp.Fitness
        }
    
	}

    return score
}

// -- No longer in use --
// ReduceTerminalToNonTerminal reduces a terminal to its associated 
// non-terminal.
func (g *Genomizer) ReduceTerminalToNonTerminal(terminalSymbol string) (IRuleModel, error) {
    
    nonTerminals, err := g.FindNonTerminalsForTerminals([]string{terminalSymbol})
    
    if err != nil {
        return nil, err
    }
    
    if len(nonTerminals) > 0 {
        return nonTerminals[0], nil
    }
    
    return nil, fmt.Errorf("No non-terminals found for the terminal '%s'", terminalSymbol)
}

func (g *Genomizer) ReduceToInitialSymbol(
    symbols []IRuleModel,
    initialSymbol string,
    maxDepth int,
    cache *ReductionChainCache,
) ([]IRuleModel, [][]IRuleModel, error) {
    currentSymbols := make([]IRuleModel, len(symbols))
    copy(currentSymbols, symbols)
    var productionSequence [][]IRuleModel
    depth := 0

    // -- Debug --
    // log.Printf("ReduceToInitialSymbol: Starting with the sequence: %v, initial symbol: %s, max depth: %d",
    //     currentSymbols, initialSymbol, maxDepth)

    for depth < maxDepth {
        reduced := false

        // 1. Reduction of first-rank non-terminals, 
        //      e.g., vowel → letter.
        if g.ApplyAtomicReductions(&currentSymbols, &productionSequence, &depth, cache) {
            reduced = true

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: [Phase 1] Atomic reduction applied. New sequence %v", currentSymbols)
        }

        // 2. Reduction of homogeneous sequences,
        //      e.g., [letter letter letter] → letters.
        if !reduced && g.ApplySequenceReduction(&currentSymbols, &productionSequence, &depth, cache) {
            reduced = true

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: [Phase 2] Homogeneous sequence reduction applied. New sequence %v", currentSymbols)            
        }

        // 3. Reduction of mixed sequences, e.g., [vowel consonant] → syllable.
        if !reduced && g.ApplyMixedSequenceReduction(&currentSymbols, &productionSequence, &depth, cache) {
            reduced = true

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: [Phase 3] Mixed Sequence Reduction Applied. New sequence %v", currentSymbols)            
        }

        // 4. Simplification of repetitive sequences,
        //      e.g., [syllable syllable] → syllable_2.
        if !reduced && g.ApplySequenceSimplification(&currentSymbols, &productionSequence, &depth, cache) {
            reduced = true

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: [Phase 4] Simplification applied. New sequence %v)", currentSymbols)
        }

        // 5. Recursive reductions, 
        //      e.g., "string → syllable string".
        if !reduced && g.ApplyDirectRecursiveMatches(&currentSymbols, &productionSequence, &depth, cache) {
            reduced = true

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: [Phase 5] Recursive reduction applied. New sequence %v", currentSymbols)
        }        

        // Check if the reduction is complete.
        if len(currentSymbols) == 1 && currentSymbols[0].GetText() == initialSymbol {

            // -- Debug --
            // log.Printf("ReduceToInitialSymbol: Reduction completed successfully, final sequence: %v, depth: %d",
            //     currentSymbols, depth)
                
            return currentSymbols, productionSequence, nil
        }

        if !reduced {
            depth++
        }

    }

    return nil, nil, fmt.Errorf("Maximum depth reached (maxDepth=%d)", maxDepth)
}

// RepairIndividual repairs an individual whose genome or history is corrupted.
func (g *Genomizer) RepairIndividual(ind IIndividual) error {

    // 1. Realization of the individual.
    concreteInd, ok := ind.(*Individual)

    if !ok {
        return fmt.Errorf("RepairIndividual: individual is not *Individual")
    }

    phenotype := concreteInd.GetPhenotype()

    if phenotype != nil && phenotype != "" {
        phenotypeStr, ok := phenotype.(string)
    
        if !ok {            
            return fmt.Errorf("RepairIndividual: phenotype is not a string")
        }
    
        if !g.CanReduceToInitialSymbol(phenotypeStr, 20) {
            
            // -- Debug --
            log.Printf("RepairIndividual: Phenotype %q cannot be reduced to initial symbol. Skipping repair.", phenotypeStr)

            return nil
        }
    
    }

    // 2. Save the initial state for restoration in case of failure.
    oldGenome := make([]int, len(concreteInd.GetGenome()))
    copy(oldGenome, concreteInd.GetGenome())
    oldPhenotype := concreteInd.GetPhenotype()
    oldHistory := utils.DeepCopyProductionHistory(concreteInd.GetProductionHistory())
    oldLastValidPhenotype := concreteInd.GetLastValidPhenotype()  // Preservation of immune memory
    oldExhausted := concreteInd.GetExhausted()

    // -- Debug --
    log.Printf("RepairIndividual: INITIAL STATE - ARNnc: %v, stack: %v, phenotype=%v, history=%v, genome=%v, last valid phenotype: %v, exhauted: %v",
        concreteInd.GetDynamicRules(), 
        concreteInd.GetDynamicRuleStack(),    
        concreteInd.GetPhenotype(), 
        concreteInd.GetProductionHistory(), 
        concreteInd.GetGenome(),
        concreteInd.GetLastValidPhenotype(),
        concreteInd.GetExhausted(),
    )

    // Restore the initial state in the event of an error or panic.
    success := false
    defer func() {

        if !success {
            concreteInd.SetGenome(oldGenome)
            concreteInd.SetPhenotype(oldPhenotype)
            concreteInd.SetProductionHistory(oldHistory)
            concreteInd.SetLastValidPhenotype(oldLastValidPhenotype)  // Restoration of immune memory
            concreteInd.SetExhausted(oldExhausted)
        }

    }()

    // 2. Check if productionHistory is empty.

    // -- Debug --
    log.Printf("RepairIndividual: history before Genomize n°1 %v", g.productionHistory)
    log.Printf("RepairIndividual: individual history before Genomize n°1 %v", ind.GetProductionHistory()) 
    log.Printf("RepairIndividual: last individual valid phenotype before Genomize n°1 %v", ind.GetLastValidPhenotype())        

    if len(concreteInd.GetProductionHistory()) == 0 {

        // Regenerate productionHistory from the genome.
        if err := g.Genomize(concreteInd.GetGenome(), concreteInd); err != nil {
            return fmt.Errorf("unable to regenerate production history: %v", err)
        }

        concreteInd.SetPhenotype(g.GetPhenotype())
        concreteInd.SetProductionHistory(utils.DeepCopyProductionHistory(g.GetProductionHistory()))

        // -- Debug --
        // log.Printf("RepairIndividual: history after Genomize n°1 %v", g.productionHistory)
        // log.Printf("RepairIndividual: individual history after Genomize n°1 %v", ind.GetProductionHistory())        
    }

    // 2. Check if the phenotype is valid.
    isPhenotypeValid := concreteInd.GetPhenotype() != nil && concreteInd.GetPhenotype() != ""

    // -- Debug --
    // log.Printf("RepairIndividual: Is phenotype %v valid before RebuildGenome: %v", ind.GetPhenotype(), isPhenotypeValid)

    // 3. Reconstruct the genome from the history or phenotype according to 
    //    isPhenotypeValid.
    if err := g.RebuildGenome(concreteInd, isPhenotypeValid); err != nil {
        return fmt.Errorf("failed to rebuild genome: %v", err)
    }
   
    // -- Debug --
    log.Printf("RepairIndividual: phenotype after RebuildGenome %v", ind.GetPhenotype())
    log.Printf("RepairIndividual: history after RebuildGenome, Genomize n°2 %v", g.productionHistory)
    log.Printf("RepairIndividual: individual history after RebuildGenome, Genomize n°2 %v", ind.GetProductionHistory())   

    // 4. Verify that the new genome generates a valid phenotype.
    if err := g.Genomize(concreteInd.GetGenome(), concreteInd); err != nil {
        return fmt.Errorf("invalid repaired genome: %v", err)
    }

    // 5. Update phenotype and history.
    concreteInd.SetPhenotype(g.GetPhenotype())

    // -- Debug --
    // log.Printf("RepairIndividual: phenotype after checking by Genomize n°3 %v, %v", isPhenotypeValid, ind.GetPhenotype())

    if len(concreteInd.GetProductionHistory()) > 0 {

        // -- Debug --
        // log.Printf("RepairIndividual: history after Genomize n°3 %v", concreteInd.productionHistory)

        concreteInd.SetProductionHistory(utils.DeepCopyProductionHistory(g.GetProductionHistory()))

        // -- Debug --
        // log.Printf("RepairIndividual: individual history after Genomize n°3 %v", concreteInd.GetProductionHistory())
    }

    // 6. Check the final consistency.
    if concreteInd.GetPhenotype() == nil || concreteInd.GetPhenotype() == "" {
        return fmt.Errorf("empty phenotype after repair")
    }

    success = true
    return nil
}

// ReplaceSuboptimalCodonBlock replaces a suboptimal codon block with an optimal
// one.
func (g *Genomizer) ReplaceSuboptimalCodonBlock(
    genome []int, 
    suboptimalIndex int, 
    blockLength int, 
    optimalBlock []int,
) []int {

    // Replace the suboptimal block with the optimal block.
    for i := 0; i < blockLength; i++ {

        if suboptimalIndex+i < len(genome) && i < len(optimalBlock) {
            genome[suboptimalIndex+i] = optimalBlock[i]
        }

    }

    return genome
}

// -- Not used yet --
// SelectBestProduction chooses the best performing production from a list.
func (g *Genomizer) SelectBestProduction(
    productions [][]IRuleModel, 
    currentFitness float64,
) []IRuleModel {
	var bestProduction []IRuleModel
	bestFitness := currentFitness

	for _, prod := range productions {
		fitness := g.GetAverageFitness(prod)
	
		if fitness > bestFitness {
			bestFitness = fitness
			bestProduction = prod
		}
	
	}
	
	return bestProduction
}

// SelectBestProductionWithBias chooses the best production from a list of 
// candidate productions, taking into account average fitness, recent success, 
// and frequency of use.
func (g *Genomizer) SelectBestProductionWithBias(
    productions [][]IRuleModel, 
    currentFitness float64,
) []IRuleModel {
    var bestProduction []IRuleModel
    bestScore := -1.0  // Initial composite score

    for _, prod := range productions {

        // 1. Calculate the average fitness of the production.
        fitness := g.GetAverageFitness(prod)

        // 2. Calculate the recent success of the production.
        recentSuccess := g.RecentSuccessScore(prod)

        // 3. Calculate the frequency of use of the production.
        frequency := float64(g.FrequencyOfProduction(prod))

        // 4. Calculate a composite score. The score is a weighted combination 
        //    of fitness, recent success, and frequency.
        score := fitness * (1.0 + 0.5*recentSuccess) * math.Log(1.0+frequency)

        // 5. Update the best production if the score is higher.
        if score > bestScore {
            bestScore = score
            bestProduction = prod
        }

    }

    return bestProduction
}

/* Deterministic selections */

// SelectProductionByGenome selects a production deterministically, using 
// the provided codon (already normalized in Generate).
// Parameters:
// - codon: the codon to use to select the production (already modulo 
//   len(choices)).
// - choices: the possible productions for the non-terminal symbol.
// Returns: the selected production.
func (g *Genomizer) SelectProductionByGenome(
    codon int,
    choices [][]IRuleModel,
) []IRuleModel {

    if len(choices) == 0 {
        return nil
    }
    
    return choices[codon]
}

/* Stochastic selections */

// SelectProductionByFitness sélectionne une production par fitness.
func (g *Genomizer) SelectProductionByFitness(
    codon int,
    choices [][]IRuleModel,
) []IRuleModel {
    
    if len(choices) == 0 {
        return nil
    }

    // -- Debug -- Verification log for each candidate production.
    // for _, production := range choices {
    //     fitness := g.GetAverageFitness(production)
    //     log.Printf("SelectProductionByFitness: production %v, fitness %v", production, fitness)
    // }    
    
    if len(g.successfulProductions) == 0 {
        return g.SelectProductionByGenome(codon, choices)
    }
    
    bestProduction := choices[0]
    bestFitness := g.GetAverageFitness(bestProduction)
    
    for _, production := range choices[1:] {
        currentFitness := g.GetAverageFitness(production)
    
        if currentFitness > bestFitness {
            bestFitness = currentFitness
            bestProduction = production
        }
    
    }
    
    if bestFitness == 0.0 {
        return g.SelectProductionByGenome(codon, choices)
    }
    
    return bestProduction
}

// SelectProductionByRandom sélectionne une production de manière aléatoire.
// Paramètres :
// - codon : non utilisé (mais gardé pour cohérence avec les autres méthodes).
// - choices : les productions possibles pour le symbole non-terminal.
// Retourne : une production sélectionnée aléatoirement.
func (g *Genomizer) SelectProductionByRandom(
    codon int, // Non utilisé, mais gardé pour cohérence
    choices [][]IRuleModel,
) []IRuleModel {

    if len(choices) == 0 {
        return nil
    }

    selectedIndex := rand.Intn(len(choices))
    return choices[selectedIndex]
}

// SpliceGenomeFromHistory generates a genome segment to be spliced ​​into 
// the existing genome. Used for targeted updating (preservation of genetic 
// information).
func (g *Genomizer) SpliceGenomeFromHistory(individual *Individual) error {

    // 1. Encode the new productionHistory into a sequence of codons.
    newCodons := g.EncodeProductionHistoryToGenomeSegment(individual.GetProductionHistory())
    // fmt.Printf("DEBUG: Encoded %d new codons from productionHistory\n", len(newCodons))

    // 2. Check that newCodons is not empty.
    if len(newCodons) == 0 {
        return fmt.Errorf("no codons encoded from productionHistory")
    }

    // 3. Check that the existing genome is the correct size.
    if len(individual.GetGenome()) != CODONS_SIZE {
        return fmt.Errorf("genome size is %d, expected %d", len(individual.GetGenome()), CODONS_SIZE)
    }

    // 4. Splice the genome: replace the first len(newCodons) codons with 
    //    newCodons.
    if len(newCodons) <= len(individual.GetGenome()) {

        // Replace the corresponding segment.
        individual.UpdateGenomeSegment(newCodons, 0)
        // fmt.Printf("DEBUG: Spliced genome: first %d codons replaced\n", len(newCodons))
    } else {

        // If newCodons is larger than the genome (unlikely case), truncate.
        individual.UpdateGenomeSegment(newCodons[:CODONS_SIZE], 0)
        // fmt.Println("DEBUG: newCodons larger than genome, truncated")
    }

    return nil
}

func (g *Genomizer) UpdatePatternLibrary(individuals []IIndividual) {

    concreteIndividuals := make([]*Individual, len(individuals))
    
	for i, p := range individuals {
        concretePop, ok := p.(*Individual)

        if !ok {
            log.Printf("individual %d is not *Individual", i)
        }
    
		concreteIndividuals[i] = concretePop
    }

    // 1. Determine a dynamic threshold based on maximum fitness.
    thresholdPercent := 0.95  // 95% of maximum fitness (adjustable parameter)
    maxFitness := FindMaxFitness(concreteIndividuals)
    threshold := thresholdPercent * maxFitness

    // 2. Filter out high-performing individuals.
    var highPerformingIndividuals []*Individual
    
    for _, ind := range concreteIndividuals {
    
        if ind.GetFitness() >= threshold {
            highPerformingIndividuals = append(highPerformingIndividuals, ind)
        }
    
    }

    // Select an individual with ncRNA (the first in the list, or the best).
    var individualWithARN *Individual

    if len(highPerformingIndividuals) > 0 {
        
        // Take the first high-performing individual (or use FindBestIndividual if necessary).
        individualWithARN = highPerformingIndividuals[0]
    } else {

        // If no high-performing individual is found, take the first 
        // individual from the population.
        if len(concreteIndividuals) > 0 {
            individualWithARN = concreteIndividuals[0]
        } else {
            
            // -- Debug --
            log.Printf("UpdatePatternLibrary: No individuals to extract patterns from")
            
            return
        }

    }    

    // 3. Extract motifs for different block sizes (2, 3, 4 codons).
    for blockLength := 2; blockLength <= 4; blockLength++ {
        patterns := g.ExtractLinguisticPatterns(
            highPerformingIndividuals, 
            threshold, 
            blockLength,
            individualWithARN,  // Individual with ncRNA
        )
    
        for _, pattern := range patterns {
    
            // 4. Check if the pattern already exists.
            existingPattern := g.PatternLibrary.FindPatternByCodons(pattern.CodonBlock)
    
            if existingPattern != nil {
    
                // Update fitness and frequency.
                g.PatternLibrary.UpdatePatternFitness(pattern.CodonBlock, pattern.Fitness)
    
                // Update the semantic tag if necessary.
                if pattern.SemanticTag != "" && existingPattern.SemanticTag == "" {
                    existingPattern.SemanticTag = pattern.SemanticTag
                }
    
            } else {
    
                // Add the new pattern.
                g.PatternLibrary.AddPattern(pattern)
            }
    
        }
    
    }

}

// UpdateReductionChain adds a new reduction chain with a priority.
func (g *Genomizer) UpdateReductionChain(
    symbol string, 
    newChain []IRuleModel, 
    cache *ReductionChainCache, 
    priority bool,
) {
    cache.mutex.Lock()
    defer cache.mutex.Unlock()

    // Add the new string to the existing list.
    if priority {
    
        // Insert at the beginning of the list to prioritize.
        cache.cache[symbol] = append([][]IRuleModel{newChain}, cache.cache[symbol]...)
    } else {
    
        // Add to the end of the list.
        cache.cache[symbol] = append(cache.cache[symbol], newChain)
    }
    
}

// UpdateSuccessfulProductions updates the subset of "successful" productions 
// after evaluation.
func (g *Genomizer) UpdateSuccessfulProductions(individuals []IIndividual) {

    concreteIndividuals := make([]*Individual, len(individuals))
    
	for i, p := range individuals {
        concretePop, ok := p.(*Individual)

        if !ok {
            log.Printf("individual %d is not *Individual", i)
        }
    
		concreteIndividuals[i] = concretePop
    }

    if len(concreteIndividuals) == 0 {

        // -- Warning --
        log.Printf("UpdateSuccessfulProductions: no valid individuals")

        return
    }

    // -- Debug --
    // log.Printf("UpdateSuccessfulProductions: before update, successfulProductions count = %v", len(g.successfulProductions))

    // 1. Determine the fitness threshold for "performing" individuals.
    thresholdPercent := 0.7  // 70% of maximum fitness
    maxFitness := FindMaxFitness(concreteIndividuals)
    threshold := thresholdPercent*maxFitness 

    // -- Debug -- Log to check the fitness threshold.
    // log.Printf("UpdateSuccessfulProductions: maxFitness = %v, threshold = %v", maxFitness, threshold)

    // 2. Update successful productions (existing logic).
    for _, ind := range concreteIndividuals {
        
        // -- Debug -- Check which individuals cross the threshold.
        // log.Printf("UpdateSuccessfulProductions: individual fitness = %v (threshold = %v)", ind.GetFitness(), threshold)

        if ind.GetFitness() >= threshold {

            // -- Debug --
            // log.Printf("UpdateSuccessfulProductions: adding %d productions from individual with fitness %v", 
            //     len(ind.GetProductionHistory()), 
            //     ind.GetFitness(),
            // )
        
            for _, production := range ind.GetProductionHistory() {
                g.AddToSuccessfulProductions(production, ind.GetFitness())
                g.AddToRecentSuccessfulProductions(production, ind.GetFitness())
            }
        
        }
    
    }

    // 3. Update successful genomes.
    for _, ind := range concreteIndividuals {

        if ind.GetFitness() >= threshold {  // Dynamic threshold (0.70 * maxFitness)
            g.AddToSuccessfulGenomes(ind.GetGenome(), ind.GetFitness())
        }

    }

    // -- Debug --
    // log.Printf("UpdateSuccessfulProductions: after update, successfulProductions count = %v", len(g.successfulProductions))    

    // 4. Clean up obsolete data.
    g.CleanSuccessfulProductions()
    g.CleanUpSuccessfulGenomes(100)  // Limit size to 100
}

// ValidatePhenotypeFromProductionSequence verifies that the phenotype 
// generated from productionSequence corresponds to the initial phenotype.
func (g *Genomizer) ValidatePhenotypeFromProductionSequence(
    productionSequence [][]IRuleModel, 
    expectedPhenotype string,
) error {

    // --- Detect if the sequence is a fallback ---
    // A fallback is a minimal sequence such as [[grammar 1] [ε 0]] where 
    // g.startRule is the fallback marker. By convention, Epsilon cancels 
    // the entire derivation context.
    isFallback := false
    
    for _, prod := range productionSequence {
    
        for _, rule := range prod {
    
            if rule.GetText() == g.startRule {
                isFallback = true
                break
            }
    
        }
    
        if isFallback {
            break
        }
    
    }

    if isFallback {
    
        log.Printf("ValidatePhenotypeFromProductionSequence: Fallback sequence detected. Skipping validation.")
    
        return nil
    }

    if len(productionSequence) == 0 {
    
        // -- Warning --
        log.Printf("ValidatePhenotypeFromProductionSequence: Empty production sequence. Using expected phenotype as fallback.")
    
        return nil  // Accept the expected phenotype as valid
    }    

    // Generate the phenotype from productionSequence.
    generatedPhenotype, err := g.GeneratePhenotypeFromProductionSequence(productionSequence)

    // -- Debug --
    // log.Printf("ValidatePhenotypeFromProductionSequence: Phenotype %v from production sequence %v",
    //     generatedPhenotype,
    //     productionSequence,
    // )

    if err != nil {
        return fmt.Errorf("error during phenotype generation: %v", err)
    }

    // Compare the generated phenotype with the expected phenotype.
    if generatedPhenotype != expectedPhenotype {
        return fmt.Errorf("generated phenotype '%s' does not correspond to the expected phenotype '%s'", generatedPhenotype, expectedPhenotype)
    }

    return nil
}
