// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The grammar package implements the evolving grammars for the Genomizer
// project.
// This package contains the concrete implementations of the interfaces
// defined in `interfaces`, notably SpeciesGrammar, which represents a
// grammar common to all individuals of a species.
//
// A grammar in this context is used to:
// - Define the production rules for generating valid phenotypes.
// - Be shared among all individuals of the same species.
// - Allow dynamic modifications (e.g., evolution of the rules).
package grammar

import (
	"io"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/participle/v2/ebnf"
)

// SpeciesGrammar represents a grammar common to all individuals of a species.
// It implements the IGrammar interface and uses a parser (IParser) and a
// serializer (ISerializer) to load and process grammar rules from an EBNF
// file.
//
// Fields:
// - ast ebnf.Node: The abstract syntax tree (AST) of the parsed grammar.
// - symbols map[string]interfaces.IRuleModel: The symbol map :
//   (symbols → rules).
// - context *kong.Context: The CLI context for error and option handling.
// - parserFac interfaces.ParserFactory: Factory for creating an EBNF parser.
// - reader io.Reader: The reader for the grammar's source file.
// - serializer interfaces.ISerializer: The serializer to convert the AST
//   into usable rules.
// - startRule string: The starting symbol of the grammar.
type SpeciesGrammar struct {
	ast 	   ebnf.Node
	symbols    map[string]IRuleModel
	terminals  map[string]bool
	context    *kong.Context
	parserFac  ParserFactory
	reader     io.Reader
	serializer ISerializer
	startRule  string
}

// Create initializes a new instance of SpeciesGrammar with the necessary 
// dependencies. This method configures the base fields but does not yet 
// parse  the grammar.
func (sg *SpeciesGrammar) Create(
	CreateSerializer SerializerFactory, 
	ctx *kong.Context, 
	reader io.Reader,
) *SpeciesGrammar {
	sg.symbols = map[string]IRuleModel{}
	sg.terminals = map[string]bool{}
	sg.context = ctx
	sg.parserFac = func() IParser { return NewParser() }
	sg.reader = reader
	sg.serializer = CreateSerializer()
	return sg
}

// GetSymbols returns the grammar's symbol map.
func (sg *SpeciesGrammar) GetSymbols() map[string]IRuleModel {
	return sg.symbols
}

// GetRuleIndex returns the index of a rule in the grammar, given its symbol 
// or text. Useful for identifying the order of rules during derivation.
func (sg *SpeciesGrammar) GetRuleIndex(rule string) int {
    rules := sg.GetRules()
    
	if rules == nil {
        return -1
    }

    // Browse the rules to find the index.
    index := 0
    
	for symbol, ruleModel := range rules {
    
		if symbol == rule || ruleModel.GetText() == rule {
            return index
        }
    
		index++
    }
    
	return -1
}

// GetRules returns the complete list of grammar rules, serialized by 
// the Serializer. Unlike GetSymbols(), this method returns the rules in 
// a ready-to-use format for derivation (e.g., with additional metadata).
func (sg *SpeciesGrammar) GetRules() map[string]IRuleModel {

    if sg.serializer == nil {
        return nil
    }

    return sg.serializer.GetRules()
}

// GetStartRule returns the starting symbol of the grammar (e.g., "string").
// This symbol is defined during serialization (via Segment()).
func (sg *SpeciesGrammar) GetStartRule() string {
	return sg.startRule
}

func (sg *SpeciesGrammar) GetTerminals() map[string]bool {

    if sg.serializer == nil {
        return nil
    }
    
	return sg.serializer.GetTerminals()
}

func (sg *SpeciesGrammar) HasTerminal(terminal string) bool {
    return sg.terminals[terminal]
}

// Parse uses the Parser to analyze the grammar file and build the AST. This 
// method is automatically called by Segment().
func (sg *SpeciesGrammar) Parse(ctx *kong.Context, reader io.Reader) error {
	parser := sg.parserFac()

	if err := parser.Parse(ctx, reader, &sg.ast); err != nil {
		return err
	}

	return nil
}

// Segment parses and serializes the grammar once :
// 1. Parses the EBNF file using the Parser.
// 2. Serializes the resulting AST using the Serializer to obtain the rules
//    and the start symbol.
// 3. Stores the start symbol in sg.startRule.
func (sg *SpeciesGrammar) Segment() error {

	// Parse the grammar.
	if err := sg.Parse(sg.context, sg.reader); err != nil {
	 	return err
	}

	// Serialize the grammar.
	if err := sg.Serialize(sg.ast); err != nil {
		return err
	}

	sg.startRule = sg.serializer.GetStartRule()
    sg.terminals = sg.serializer.GetTerminals()  // Retrieves the terminals

	// --Debug --
	// log.Printf("Segment: Symbols: %v", sg.symbols)
	// log.Printf("Segment: Terminals: %v", sg.terminals)
	// log.Printf("Segment: Start rule: %s", sg.startRule)
	
	return nil
}

// Serialize uses the Serializer to convert the AST (Abstract Syntax Tree)
// into a **linear string of symbols** (analogous to a DNA sequence). 
// This transformation allows the grammar to be represented in a form usable 
// by the Genomizer for phenotype derivation.
//
// The process delegates the following steps to the Serializer:
// 1. Scans the AST to extract terminal and non-terminal symbols.
// 2. Generates an ordered sequence of symbols (e.g., ["letter", "digit", 
//    ...]).
// 3. Stores the result in sg.symbols for later use.
func (sg *SpeciesGrammar) Serialize(node ebnf.Node) error {

	if err := sg.serializer.Serialize(sg.symbols, node); err != nil {
		return err
	}

	return nil
}

// UpdateSymbols updates the grammar's symbol map. Used to dynamically modify 
// production rules (e.g., during evolution).
func (sg *SpeciesGrammar) UpdateSymbols(newSymbols map[string]IRuleModel) {
	sg.symbols = newSymbols
}