// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// Serializes a grammar defined using the Extended Backus-Naur Form (EBNF)
// into a set of rules and symbols. The nodes produced by the EBNF parser
// are traversed before being written back into the symbol table.
package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/alecthomas/participle/v2/ebnf"

	"evoGo/model"
	"evoGo/patterns/algo"
	"evoGo/patterns/builder"
	"evoGo/grammar/optimizations"
)

// Interfaces imported to ensure architectural consistency.
type (
	IBuilder = builder.IBuilder

	IAlgo               = algo.IAlgo
	IParseEBNF          = algo.IParseEBNF
	IParseExpression    = algo.IParseExpression
	IParseRule          = algo.IParseRule
	IParseSequence      = algo.IParseSequence
	IParseSubExpression = algo.IParseSubExpression
	IParseTerm          = algo.IParseTerm

	IRuleModel         = model.IRuleModel
	EBNFModel          = model.EBNFModel
	ExpressionModel    = model.ExpressionModel
	IdentifierModel    = model.IdentifierModel
	LiteralModel       = model.LiteralModel
	RuleModel          = model.RuleModel
	SubExpressionModel = model.SubExpressionModel
	SequenceModel      = model.SequenceModel
	SymbolType         = model.SymbolType
	TermModel          = model.TermModel
)

// Constants for symbol types.
const (
	Terminal    SymbolType = model.Terminal
	NonTerminal SymbolType = model.NonTerminal
)

/* Serializer */

type (
	ISerializer interface {
		GetRules() map[string]IRuleModel
		GetStartRule() string
		Serialize(rules map[string]IRuleModel, node ebnf.Node) error
		ToString() (out string)
	}

	// Serializer is responsible for serializing an EBNF grammar into a set 
	// of rules and symbols.
	Serializer struct {
		rules     map[string]IRuleModel
		lhs       string
		startRule string

		algo              IAlgo
		builder           IBuilder
		expressionCase    IParseExpression
		identifierCase    IParseTerm
		literalCase       IParseTerm
		ruleCase          IParseRule
		rulesCase         IParseEBNF
		sequenceCase      IParseSequence
		subExpressionCase IParseSubExpression
		termCase          IParseTerm
	}
)

// Create initializes the Serializer with the necessary dependencies.
func (s *Serializer) Create(CreateAlgo algo.AlgoFactory, CreateBuilder builder.BuilderFactory) *Serializer {
	s.algo = CreateAlgo()
	s.builder = CreateBuilder()
	return s
}

// MakeAlgo initializes the algorithms and the processing cases.
func (s *Serializer) MakeAlgo() {
	s.expressionCase = s.algo.MakeExpressionCase(s.HandleExpressionFunc())
	s.identifierCase = s.algo.MakeTermCase(s.HandleIdentifierFunc())
	s.literalCase = s.algo.MakeTermCase(s.HandleLiteralFunc())
	s.ruleCase = s.algo.MakeRuleCase(s.HandleRuleFunc())
	s.rulesCase = s.algo.MakeRulesCase(s.HandleRulesFunc())
	s.sequenceCase = s.algo.MakeSequenceCase(s.HandleSequenceFunc())
	s.subExpressionCase = s.algo.MakeSubExpressionCase(s.HandleSubExpressionFunc())
	s.termCase = s.algo.MakeTermCase(s.HandleTermFunc())
}

// DoHandleExpression processes an expression model by constructing sequences
// for each alternative and updating the model symbols.
func (s *Serializer) DoHandleExpression(model *ExpressionModel) {

	// Iterates over each alternative in the expression.
	for i, n := range model.GetExpression().Alternatives {

		// Builds sequence for this alternative.
		if rhsSeq, err := s.builder.BuildSequence(n, s.sequenceCase); err == nil && rhsSeq != nil {

			if len(rhsSeq) > 0 {
				model.Symbols = append(model.Symbols, rhsSeq)
			} else {
				altGroup := s.builder.GetTerms()
				model.Symbols = append(model.Symbols, altGroup...)
			}

		} else {
			log.Printf("Expression failed to build alternative at index %d: %v \n", i, err)
		}

	}

	// Checks if the model is valid after processing.
	if !model.IsValid() {
		log.Printf("Expression model is invalid after processing")
	}

}

// DoHandleIdentifier handles an identifier model.
func (s *Serializer) DoHandleIdentifier(model *IdentifierModel) {

	// Initializes the lexeme.
	model.InitializeLexeme()

	// Checks if the identifier is valid after proceeding.
	if !model.IsValid() {
		log.Printf("Invalid Identifier : %v", model)
		return
	}

}

// DoHandleLiteral handles a literal model.
func (s *Serializer) DoHandleLiteral(model *LiteralModel) {

	// Initializes the lexeme.
	model.InitializeLexeme()

	// Checks if the literal is valid after proceeding.
	if !model.IsValid() {
		log.Printf("Invalid literal : %v", model)
		return
	}

}

// DoHandleRule processes a rule model.
func (s *Serializer) DoHandleRule(model *RuleModel) {
	n := model.Production
	s.lhs = n.Production

	if s.rules == nil {

		// Initializes the rules map if not already done.
		s.rules = map[string]IRuleModel{}
	}

	// Checks if a rule already exists for this symbol.
	if existingRule := s.rules[s.lhs]; existingRule != nil {

		// Handles the case where a rule already exists for this symbol.
		log.Printf("Rule %s already exists. Consider replacing or merging.", s.lhs)
		return
	}

	model.Symbol = model.GetIdentifier()
	model.SymbolType = NonTerminal

	// Checks if the rule is valid after proceeding.
	if !model.IsValid() {
		log.Printf("Invalid rule: %v", model)
		return
	}

	// Constructs the expression for the rule.
	if rhsExpr, err := s.builder.BuildExpression(n, s.expressionCase); err == nil && rhsExpr != nil {
		model.SetSymbols(rhsExpr)
	} else {
		log.Printf("Error building expression for rule %s: %v", s.lhs, err)
	}

	// Applies the left factorization to the rule.
	factorizedModel := optimizations.LeftFactorize(model, s.rules)

	if factorizedModel, ok := factorizedModel.(*RuleModel); ok {
    	model = factorizedModel  // Updates model with the factored rule
	} else {
    	log.Printf("Error: LeftFactorize returned an unexpected type (type: %T)", factorizedModel)
	}
	
	s.rules[s.lhs] = model
}

// DoHandleRules handles an EBNF model.
func (s *Serializer) DoHandleRules(model *EBNFModel) {

	for i, n := range model.GetEBNF().Productions {

		if lhs, err := s.builder.BuildRule(n, s.ruleCase); err == nil && lhs != "" {

			// Checks if it's the start rule.
			if s.startRule == "" {
				s.startRule = lhs
			}

			// Resolves indirect circular dependencies between non-terminals.
			optimizations.IndirectLeftRecurse(s.rules)

		} else {
			log.Printf("Grammar failed to build production at index %d: %v \n", i, err)
		}

	}

}

// DoHandleSequence processes a sequence model.
func (s *Serializer) DoHandleSequence(model *SequenceModel) {

	// Iterates over each term in the sequence.
	for i, n := range model.GetSequence().Terms {

		// Builds term for this sequence.
		if rhsTerm, err := s.builder.BuildTerm(n, s.termCase); err == nil && rhsTerm != nil {
			rhsSymbols := rhsTerm.GetSymbols()

			if rhsSymbols != nil && len(rhsSymbols) > 0 {

				// Case of sequence term groups.
				if len(rhsSymbols) == 1 {
					model.Symbols = append(model.Symbols, rhsSymbols[0]...)

					// Case of alternative term groups.
				} else {
					s.builder.ModifyTerms(func(terms [][]IRuleModel) [][]IRuleModel {
						terms = append(terms, rhsSymbols...)
						return terms
					})
				}

			} else {
				model.Symbols = append(model.Symbols, rhsTerm)
			}

		} else {
			log.Printf("Sequence failed to build term at index %d in sequence: %v \n", i, err)
		}

	}

	// Performs any necessary validation after processing all terms.
	if !model.IsValid() {
		log.Printf("Sequence model is invalid after processing")
	}

}

// DoHandleSubExpression handles a subexpression model.
func (s *Serializer) DoHandleSubExpression(model *SubExpressionModel) {

	// Iterates over each alternative in the group of subexpressions.
	for i, n := range model.GetSubExpression().Expr.Alternatives {

		// Builds sequence for this alternative.
		if rhsSeq, err := s.builder.BuildSequence(n, s.sequenceCase); err == nil && rhsSeq != nil {
			model.Symbols = append(model.Symbols, rhsSeq)
		} else {
			log.Printf("Expression failed to build alternative at index %d: %v \n", i, err)
		}

	}

	// Checks if the model is valid after processing.
	if !model.IsValid() {
		log.Printf("Sub-expression model is invalid after processing")
	}

}

// DoHandleTerm processes a term model.
func (s *Serializer) DoHandleTerm(model *TermModel) {
	n := model.Term

	switch {
	case n.Name != "":

		if rhsIdent, err := s.builder.BuildIdentifier(n, s.identifierCase); err == nil && rhsIdent != nil {
			model.SetLexeme(rhsIdent)
		} else {
			log.Printf("Failed to create identifier: %v", err)
		}

	case n.Literal != "":

		if rhsText, err := s.builder.BuildLiteral(n, s.literalCase); err == nil && rhsText != nil {
			model.SetLexeme(rhsText)
		} else {
			log.Printf("Failed to create literal: %v", err)
		}

	case n.Group != nil:

		if rhsGroup, err := s.builder.BuildSubExpression(n, s.subExpressionCase); err == nil && rhsGroup != nil {

			// Applies the optimizations to rhsGroup.
        	optimizedRHS := optimizations.ExpandAndSimplifyRepetition(n.Repetition, rhsGroup, s.rules)

        	// Defines the RHS of the lexeme with the optimized result.
			if concreteLexeme, ok := model.GetLexeme().(*RuleModel); ok {
    			concreteLexeme.SetSymbols(optimizedRHS)
			} else {
    			log.Printf("Error: lexeme is not a *RuleModel (type : %T)", model.GetLexeme())
			}
		
		} else {
			log.Printf("Failed to create sub-expression '%s' for node %s: %s", n.Group.Expr, n, err)
		}

	}

}

// GetStartRule returns the starting rule of the grammar.
func (s *Serializer) GetStartRule() string {
	return s.startRule
}

// GetRules returns the list of grammar rules.
func (s *Serializer) GetRules() map[string]IRuleModel {
	return s.rules
}

// HandleExpressionFunc returns a function that handles expression models.
func (s *Serializer) HandleExpressionFunc() func(any) {
	return func(data any) {
		model, ok := data.(*ExpressionModel)

		if !ok {
			log.Println(`Error: model is not an expression model`)
		} else {
			s.DoHandleExpression(model)
		}

	}
}

// HandleIdentifierFunc returns a function that handles identifier models.
func (s *Serializer) HandleIdentifierFunc() func(any) {
	return func(data any) {
		model, ok := data.(*IdentifierModel)

		if !ok {
			log.Println(`Error: model is not an identifier model`)
		} else {
			s.DoHandleIdentifier(model)
		}

	}
}

// HandleLiteralFunc returns a function that handles literal models.
func (s *Serializer) HandleLiteralFunc() func(any) {
	return func(data any) {
		model, ok := data.(*LiteralModel)

		if !ok {
			log.Println(`Error: model is not a literal model`)
		} else {
			s.DoHandleLiteral(model)
		}

	}
}

// HandleRuleFunc returns a function that manages rule models.
func (s *Serializer) HandleRuleFunc() func(any) {
	return func(data any) {
		model, ok := data.(*RuleModel)

		if !ok {
			log.Println(`Error: model is not a rule model`)
		} else {
			s.DoHandleRule(model)
		}

	}
}

// HandleRulesFunc returns a function that handles EBNF models.
func (s *Serializer) HandleRulesFunc() func(any) {
	return func(data any) {
		model, ok := data.(*EBNFModel)

		if !ok {
			log.Println(`Error: model is not a ebnf model`)
		} else {
			s.DoHandleRules(model)
		}

	}
}

// HandleSequenceFunc returns a function that handles sequence models.
func (s *Serializer) HandleSequenceFunc() func(any) {
	return func(data any) {
		model, ok := data.(*SequenceModel)

		if !ok {
			log.Println(`Error: model is not a sequence model`)
		} else {
			s.DoHandleSequence(model)
		}

	}
}

// HandleSubExpressionFunc returns a function that handles subexpression 
// models.
func (s *Serializer) HandleSubExpressionFunc() func(any) {
	return func(data any) {
		model, ok := data.(*SubExpressionModel)

		if !ok {
			log.Println(`Error: model is not a sub-expression model`)
		} else {
			s.DoHandleSubExpression(model)
		}

	}
}

// HandleTermFunc returns a function that handles term models.
func (s *Serializer) HandleTermFunc() func(any) {
	return func(data any) {
		model, ok := data.(*TermModel)

		if !ok {
			log.Println(`Error: model is not a term model`)
		} else {
			s.DoHandleTerm(model)
		}

	}
}

// ToString displays the grammar rules table.
func (s *Serializer) ToString() string {
	out := ""

	for symbol, rule := range s.rules {
		out += fmt.Sprintf(
			"Rule %s →  %s %v := %v\n",
			symbol, rule.GetText(), rule.GetSymbolType(), rule.GetSymbols(),
		)
	}

	return out
}

// Serialize serializes an EBNF grammar into a set of rules and symbols.
func (s *Serializer) Serialize(rules map[string]IRuleModel, node ebnf.Node) error {
	
	// Assigns the provided rules.
	s.rules = rules

	// Initialize internal algorithms and data structures.
	s.MakeAlgo()

	// Attempts to build an EBNF production using the provided node.
	if err := s.builder.BuildEBNF(node, s.rulesCase); err != nil {
		return fmt.Errorf("error building production for grammar: %v", err)
	}

	// Sorting of s.rules keys alphabetically.
	sortedKeys := make([]string, 0, len(s.rules))

	for key := range s.rules {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys) // Alphabetical sort

	// Rebuild s.rules in sorted order.
	newRules := make(map[string]IRuleModel)

	for _, key := range sortedKeys {
		newRules[key] = s.rules[key]
	}

	s.rules = newRules

	return nil
}

/* Exports */

// SerializerFactory is a function that creates an instance of ISerializer.
type SerializerFactory func() ISerializer

// NewSerializer creates a new instance of Serializer.
func NewSerializer() ISerializer {
	return new(Serializer).Create(
		func() IAlgo { return algo.NewAlgo() },
		func() IBuilder { return builder.NewBuilder() },
	)
}
