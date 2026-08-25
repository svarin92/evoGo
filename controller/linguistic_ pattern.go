// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import "fmt"

/* LinguisticPattern */

// A linguistic pattern can be defined as: A sequence of codons
//
//      (e.g.: [10, 20, 30])
//
// associated with a sequence of grammatical productions
//
//      (e.g. string → letter → consonant → 'b'),
//
// an average fitness (calculated on the individuals where the pattern
// appears), a frequency of occurrence in high-performing individuals,
// an emerging semantics (e.g. “valid word beginning”, “transition to a
// vowel”).
type LinguisticPattern struct {
    CodonBlock      []int           // A block of recurrent codons in the genome
    ProductionBlock [][]IRuleModel  // A sequence of production rules associated with this block
    Fitness         float64         // The average fitness score of individuals using this pattern
    Frequency       int             // The number of times this pattern was observed
    SemanticTag     string          // A semantic label (e.g., "vowel_transition", "word_start")
}

/* LinguisticPatternLibrary */

// Pattern library indexed by codon sequence (for fast searching) and 
// semantics (for searching by grammatical function).
type LinguisticPatternLibrary struct {
    PatternsByCodons    map[string]LinguisticPattern    // Key: fmt.Sprintf("%v", block)
    PatternsBySemantics map[string][]LinguisticPattern  // Key: semantic tag
}

// Adds a pattern to the library.
func (lib *LinguisticPatternLibrary) AddPattern(pattern LinguisticPattern) {
    
    // Indexing by codon sequence.
    blockKey := fmt.Sprintf("%v", pattern.CodonBlock)
    lib.PatternsByCodons[blockKey] = pattern

    // Indexing by semantic tag.
    if pattern.SemanticTag != "" {
        lib.PatternsBySemantics[pattern.SemanticTag] = append(lib.PatternsBySemantics[pattern.SemanticTag], pattern)
    }

}

// Search for a motif by its codon sequence and return a pointer to the motif 
// if it exists.
func (lib *LinguisticPatternLibrary) FindPatternByCodons(block []int) *LinguisticPattern {
    blockKey := fmt.Sprintf("%v", block)

    if pattern, exists := lib.PatternsByCodons[blockKey]; exists {
        return &pattern
    }

    return nil
}

// Search for all patterns associated with a given semantic tag.
func (lib *LinguisticPatternLibrary) FindPatternsBySemantics(tag string) []LinguisticPattern {
    return lib.PatternsBySemantics[tag]
}

// Update the fitness of an existing pattern, identified by its codon sequence.
func (lib *LinguisticPatternLibrary) UpdatePatternFitness(block []int, newFitness float64) bool {
    blockKey := fmt.Sprintf("%v", block)

    if pattern, exists := lib.PatternsByCodons[blockKey]; exists {

        // Fitness update (frequency-weighted average).
        pattern.Fitness = (pattern.Fitness*float64(pattern.Frequency) + newFitness) / float64(pattern.Frequency+1)
        pattern.Frequency++
        lib.PatternsByCodons[blockKey] = pattern
        return true
    }
    
    return false
}