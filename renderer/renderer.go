// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package renderer

import (
	"fmt"
	"math"
	"os"
	"strings"

	"evoGo/utils"
)

// Renderer handles the export and display of grammatical derivations. It 
// depends solely on the interfaces defined in interfaces/provider.go.
type Renderer struct {
    grammarProvider IGrammarProvider
}

// Create initializes the Renderer with a function that provides
// IGrammarProvider.
func (r *Renderer) Create(grammarProvider GrammarProviderFactory) *Renderer {
    r.grammarProvider = grammarProvider()
    return r
}

// DisplayIndividualDetails displays the exact generation where the best 
// individual was found.
func (r *Renderer) DisplayIndividualDetails(provider IProductionHistoryProvider, generation int) {
    fmt.Printf("\n=== Best Individual (Generation %d) ===\n", generation)
    fmt.Printf("Phenotype: %v\n", provider.GetPhenotype()) 
    fmt.Printf("Fitness: %.2f\n", provider.GetFitness())   
    fmt.Printf("Genome: %v, Length: %v\n", provider.GetGenome(), len(provider.GetGenome()))
}

// Generate a DOT file representing the grammatical derivation of an 
// individual. Terminal nodes are shown as ellipses (green), non-terminal
// nodes as boxes (blue). Errors are marked in red and corrections in 
// orange.
func (r *Renderer) ExportToDOT(provider IProductionHistoryProvider, filename string) error {
    file, err := os.Create(filename)
    
    if err != nil {
        return fmt.Errorf("failed to create DOT file: %v", err)
    }
    
    defer file.Close()

    // Write the DOT header.
    _, err = fmt.Fprintf(file, `digraph Derivation {
                rankdir=LR;
                node [shape=box, style=filled, fillcolor=lightblue];`,
            )
    
    if err != nil {
        return err
    }

    // Generate the nodes and edges.
    nodeID := 0
    nodeMap := make(map[int]int)  // Map: production index → ​​DOT node ID

    // Retrieve the data from the provider.
    productionHistory := provider.GetProductionHistory()

    for i, production := range productionHistory {
    
        if len(production) == 0 {
            continue
        }
    
        nodeID++
        nodeMap[i] = nodeID

        // Determine the style of the knot based on the first symbol.
        firstSymbol := production[0].GetText()
        isNonTerminal := true

        // Check if the symbol exists in r.grammarProvider.GetSymbols().
        if _, exists := r.grammarProvider.GetSymbols()[firstSymbol]; !exists {
            isNonTerminal = false  // Terminal
        }        

        // Special cases for special terminals (ε, _, or prefix ').
        if firstSymbol == "ε" || firstSymbol == "_" || strings.HasPrefix(firstSymbol, "'") {
            isNonTerminal = false
        }

        // Define the color and shape of the node.
        nodeStyle := `shape=box, style=filled, fillcolor=lightblue` // Non-terminal by default

        if !isNonTerminal {
            nodeStyle = `shape=ellipse, style=filled, fillcolor=lightgreen` // Terminal
        }
        
        // Calculate the fitness of the production.
        fitness := r.grammarProvider.GetAverageFitness(production)
        
        // Define the color based on fitness.
        if fitness < 0.5 {
            nodeStyle += `, fillcolor=orange`  // Correction needed
        } else if fitness < 0.8 {
            nodeStyle += `, fillcolor=yellow`  // Average fitness
        } else {
            nodeStyle += `, fillcolor=lightblue`  // Good fitness
        }

        // Write the node to the DOT file.
        label := fmt.Sprintf("%s", firstSymbol)

        if !isNonTerminal {
            label = fmt.Sprintf("'%s'", firstSymbol)  // Add quotation marks for terminals.
        }
        
        _, err = fmt.Fprintf(file, "  %d [label=\"%s\", %s];\n", nodeID, label, nodeStyle)
        
        if err != nil {
            return err
        }

    }

    // Write the edges between the nodes.
    for i := 0; i < len(productionHistory)-1; i++ {

        if len(productionHistory[i]) == 0 || len(productionHistory[i+1]) == 0 {
            continue
        }
        
        fromNode := nodeMap[i]
        toNode := nodeMap[i+1]
        _, err = fmt.Fprintf(file, "  %d -> %d;\n", fromNode, toNode)
        
        if err != nil {
            return err
        }

    }

    // Close the DOT graph.
    _, err = fmt.Fprintf(file, "}\n")
    
    if err != nil {
        return err
    }

    return nil        
}

// Display a detailed report of potential or completed corrections to the 
// genome. If `afterCorrection` is true, displays the status after correction.
func (r *Renderer) PrintGenomeCorrection(
	provider IProductionHistoryProvider,
	fitnessThreshold float64, 
	afterCorrection bool,
    averageFitness float64,
) {
    const (
        colorReset  = "\033[0m"
        colorRed    = "\033[31m"
        colorGreen  = "\033[32m"
        colorYellow = "\033[33m"
    )

    var title string

	if afterCorrection {
        title = "\n=== Genome Correction Report (AFTER) ==="
    } else {
        title = "\n=== Genome Correction Report (BEFORE) ==="
    }

	fmt.Println(title)

    // 1. Display phenotype and fitness.
    fmt.Printf("Phenotype: %s (Fitness: %.2f)\n", provider.GetPhenotype(), provider.GetFitness())

    // 2. Display production history.
    fmt.Println("\n--- Production History ---")

    for i, prod := range provider.GetProductionHistory() {
        avgFitness := r.grammarProvider.GetAverageFitness(prod)
        line := fmt.Sprintf("%2d. %s (Average fitness: %.2f)", i, utils.RuleModelSliceToString(prod), avgFitness)

        // Mark the productions to be corrected.
        if !afterCorrection && avgFitness < fitnessThreshold {
            line += fmt.Sprintf(" %s**TO CORRECT**%s", colorRed, colorReset)
        }

        // Mark corrected productions.
        if afterCorrection && avgFitness >= fitnessThreshold {
            line += fmt.Sprintf(" %s**CORRECTED**%s", colorGreen, colorReset)
        }

        fmt.Println(line)
    }

    // 3. If before correction, display similar productions available.
    if !afterCorrection {
        fmt.Println("\n--- Available Similar Productions for Correction ---")

        for i, prod := range provider.GetProductionHistory() {
            avgFitness := r.grammarProvider.GetAverageFitness(prod)
        
			if avgFitness < fitnessThreshold {
                similarProds := r.grammarProvider.FindSimilarProductions(prod, averageFitness)
        
				if len(similarProds) > 0 {
                    fmt.Printf("  For production %d (%s):\n", i, utils.RuleModelSliceToString(prod))
        
					for _, simProd := range similarProds {
                        fmt.Printf("    - %s (avg fitness: %.2f)\n",
                            utils.RuleModelSliceToString(simProd),
                            r.grammarProvider.GetAverageFitness(simProd))
                    }
        
				} else {
                    fmt.Printf("  For production %d (%s): %s**NO SIMILAR PRODUCTION**%s\n",
                        i, utils.RuleModelSliceToString(prod), colorYellow, colorReset)
                }
        
			}
        
		}
    
	}

    // 4. Summary of corrections (if after correction).
    if afterCorrection {
        fmt.Println("\n--- Correction Summary ---")
        oldFitness := 0.0 // Replace with the value before correction if available
        fmt.Printf("Fitness change: %.2f → %.2f\n", oldFitness, provider.GetFitness())
    }
	
}

func (r *Renderer) PrintGrammaticalDerivation(provider IProductionHistoryProvider) {
    
    const (
        colorReset  = "\033[0m"
        colorBlue   = "\033[34m"  // Non-terminals
        colorGreen  = "\033[32m"  // Terminals
        colorYellow = "\033[33m"  // Corrections
        colorRed    = "\033[31m"  // Errors
    )

    // Logging line for debugging.
    fmt.Println("\n--- Grammatical Path of the Phenotype ---")
    fmt.Printf("Phenotype: %v (Fitness: %.2f)\n",
        provider.GetPhenotype(),
        provider.GetFitness(),
    )
    fmt.Printf("PrintGrammaticalDerivation: history %v\n", provider.GetProductionHistory())

    productionHistory := provider.GetProductionHistory()
    
    for i, production := range productionHistory {

        if len(production) == 0 {
            continue
        }

        // Construct the output line.
        var symbols []string
        
        for _, rule := range production {
            symbol := rule.GetText()

            // Check if the symbol is a terminal (e.g., 'a', 'b', 'ε', '_').
            if _, isTerm := r.grammarProvider.GetSymbols()[symbol]; !isTerm || symbol == "ε" || symbol == "_" || strings.HasPrefix(symbol, "'") {
                
                // Terminal: in green with quotation marks.
                symbols = append(symbols, colorGreen+"'"+symbol+"'"+colorReset)
            } else {

                // Non-terminal: in blue.
                symbols = append(symbols, colorBlue+symbol+colorReset)
            }
        
        }

        line := fmt.Sprintf("%2d. %s", i+1, strings.Join(symbols, " "))
        fitness := r.grammarProvider.GetAverageFitness(production)
        line += fmt.Sprintf(" (fitness: %.2f)", fitness)

        // Compare current fitness with fitness before correction.
        key := fmt.Sprintf("%v", production)
        
        if oldFitness, exists := provider.GetOldProductionFitness(key); exists {
        
            if fitness > oldFitness {
                line += colorYellow + " *" + colorReset // Correction marker
            }
        
        }

        fmt.Println(line)
    }

    fmt.Println("----------------------------------")
}

// Retrieve the original rule from g.symbols using the first symbol of each 
// production. Display the LHS followed by the possible RHS.
func (r *Renderer) PrintOriginalRules() {
	fmt.Println("\n--- Original Rules ---")

    for symbol, rule := range r.grammarProvider.GetSymbols() {
        fmt.Printf("%s → ", symbol)

        for i, rhs := range rule.GetSymbols() {
        
			if i > 0 {
                fmt.Print(" | ")
            }
        
			var symbols []string
        
			for _, s := range rhs {
                symbols = append(symbols, s.GetText())
            }
        
			fmt.Printf("%s", strings.Join(symbols, " "))
        }
        
		fmt.Println()
    }

}

// Print the statistics for the generation and individuals.
func (r *Renderer) PrintStats(
    provider IPopulationStatsProvider, 
    generation int,
    generationSize int,
) {
    
    // Internal function to calculate the average.
	average := func(values []float64) float64 {
		sum := 0.0

		for _, v := range values {
			sum += v
		}

		return sum / float64(len(values))
	}

	// Internal function to calculate the standard deviation.
	stdDev := func(values []float64, mean float64) float64 {
		sum := 0.0

		for _, v := range values {
			sum += math.Pow(v-mean, 2)
		}

		return math.Sqrt(sum / float64(len(values)))
	}

	averageFloatToInt := func(values []int) int {
		sum := 0

		for _, v := range values {
			sum += v
		}

		return sum / len(values)
	}

	stdDevFloatToInt := func(values []int, mean int) int {
		sum := 0

		for _, v := range values {
			sum += (v - mean) * (v - mean)
		}

		return int(math.Sqrt(float64(sum) / float64(len(values))))
	}    
    
    individuals := provider.GetIndividuals() 

    // Filtering of invalid individuals (phenotype != nil).
    var validInds []IProductionHistoryProvider
    
    for _, ind := range individuals {
    
        if ind.GetPhenotype() != nil { 
            validInds = append(validInds, ind)
        }
    
    }

    if len(validInds) == 0 {
        fmt.Printf("Generation %d: No valid solution\n", generation)
        return
    }

    // Calculate the statistics.
    fitnessVals := make([]float64, len(validInds))
    usedCodonsVals := make([]int, len(validInds))

    for i, ind := range validInds {
        fitnessVals[i] = ind.GetFitness()       
        usedCodonsVals[i] = ind.GetUsedCodons() 
    }

    aveFit := average(fitnessVals)
    stdFit := stdDev(fitnessVals, aveFit)
    aveUsedCodons := averageFloatToInt(usedCodonsVals)
    stdUsedCodons := stdDevFloatToInt(usedCodonsVals, aveUsedCodons)

  	// Displays the generation value (the current generation number), the 
    // cumulative number of individuals evaluated since the start of the 
    // algorithm, the average fitness (aveFit) of able-bodied individuals 
    // of the generation with 2 digits after the decimal point and a margin 
    // of error (+-) given by standard deviation of the fitness (stdFit) of 
    // valid individuals, the average number of codons used (aveUsedCodons) to 
    // generate the phenotypes of valid individuals and a text representation 
    // of the best valid individual (validInds[0]).
    fmt.Printf("Gen: %d Evals: %d Ave: %.2f +- %.3f AveUsedCodons: %.2f +- %.3f %v\n",
        generation,                     // Generation number
        (generationSize * generation),  // The size of the number of individuals generated for each generation
        aveFit,                         // Average fitness of individuals in this generation
        stdFit,                         // Standard deviation of fitness values
        float64(aveUsedCodons),         // Average number of used codons (genes) per individual
        float64(stdUsedCodons),         // Standard deviation of used codon counts
        validInds[0],                   // Displays the best-suited individual
    )
}

func (r *Renderer) PrintSuccessfulProductions() {

    for _, sp := range r.grammarProvider.GetSuccessfulProductions() {
        fmt.Printf("Successful production: %v, Fitness: %.2f, Frequency: %d\n",
            utils.RuleModelSliceToString(sp.GetProduction()), sp.GetFitness(), sp.GetFrequency())
    }

}