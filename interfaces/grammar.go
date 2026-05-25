// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The interfaces package defines the contracts (interfaces) for the key 
// components of the project, including evolving grammars, parsers, and 
// individuals. These interfaces allow for strong decoupling between 
// implementations and their use.
package interfaces

// IGrammar defines the contract for an evolutionary grammar used within the 
// framework of the Grammatical Evolution (GE) algorithm.
// A grammar describes the production rules for generating valid phenotypes
// from genomes (codon sequences).
//
// In the context of the Genomizer, a grammar is used to:
// - Define terminal and non-terminal symbols (e.g., "letter", "string").
// - Guide the derivation of the phenotype from an individual's genome.
// - Allow dynamic modifications (e.g., codon updates).
type IGrammar = interface {

	// GetSymbols returns the grammar's codon map, where each key is a symbol
	// (e.g., "letter") and each value is a rule model (IRuleModel) associating
	// the symbol with its possible productions. The codon map can be nil if 
	// the grammar is not initialized.
	GetSymbols() map[string]IRuleModel

	// GetRuleIndex returns the index of a rule in the grammar, given its symbol 
	// or text. Useful for identifying the order of rules during derivation.
	GetRuleIndex(rule string) int

	// GetRules returns the complete list of grammar rules, serialized by 
	// the Serializer. Unlike GetSymbols(), this method returns the rules in 
	// a ready-to-use format for derivation (e.g., with additional metadata).
	GetRules() map[string]IRuleModel

	// GetStartRule returns the starting symbol of the grammar (e.g., 
	// "string"). This symbol is used as the entry point for phenotype 
	// derivation.
	GetStartRule() string

	// Segment parses and serializes the grammar from a given file or source.
	// This method is called only once to initialize the grammar. It combines:
	// 1. Parsing the EBNF file (via IParser).
	// 2. Serializing the rules into an internal format (via ISerializer).
	Segment() error

	// UpdateSymbols updates the grammar's codon map. Used to dynamically 
	// modify production rules (e.g., during evolution).
	UpdateSymbols(newSymbols map[string]IRuleModel)
}

// GrammarFactory is a factory function for creating instances of IGrammar. 
// Used to decouple grammar creation from usage.
type GrammarFactory func() IGrammar