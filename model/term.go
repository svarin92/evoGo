// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package model

import (
	"strings"

	"github.com/alecthomas/participle/v2/ebnf"
)

/* IdentifierModel */

// IdentifierModel represents an identifier in an EBNF expression.
type IdentifierModel struct {
	TermModel
}

// CreateFrom initializes an IdentifierModel from an EBNF node.
func (im *IdentifierModel) CreateFrom(node ebnf.Node) IIdentifierModel {
	im.TermModel.CreateFrom(node)
	return im
}

func (im *IdentifierModel) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case IParseTerm:
		v.Visit(im)
	default:
		im.VisitedModel.DoAccept(v)
	}

}

// GetIdentifier returns the identifier.
func (im *IdentifierModel) GetIdentifier() string {
	return im.Term.Name
}

func (im *IdentifierModel) GetText() string {
	return im.GetIdentifier()
}

func (im *IdentifierModel) GetLexeme() IRuleModel {
	return im.TermModel.GetLexeme()
}

func (im *IdentifierModel) InitializeLexeme() {
    im.lexeme.Symbol = im.GetIdentifier()
    im.lexeme.SymbolType = NonTerminal
}

// Delegate to the lexemeAn identifier is valid if its lexeme is valid and 
// its name is not empty.
func (im *IdentifierModel) IsValid() bool {

    if !im.TermModel.IsValid() {  // Call the parent method
        return false
    }
    
	return im.GetIdentifier() != ""  // Checks that the identifier is not empty
}

/* LiteralModel */

// LiteralModel represents a literal in an EBNF expression.
type LiteralModel struct {
	TermModel
}

// CreateFrom initializes a LiteralModel from an EBNF node.
func (lm *LiteralModel) CreateFrom(node ebnf.Node) ILiteralModel {
	lm.TermModel.CreateFrom(node)
	return lm
}

func (lm *LiteralModel) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case IParseTerm:
		v.Visit(lm)
	default:
		lm.VisitedModel.DoAccept(v)
	}

}

func (lm *LiteralModel) GetLexeme() IRuleModel {
	return lm.TermModel.GetLexeme()
}

// GetText returns the text of the literal.
func (lm *LiteralModel) GetText() string {

	if lm.Term != nil && lm.Term.Literal != "" {
    	parts := strings.Split(lm.Term.Literal, `"`)
    
		if len(parts) >= 3 {  // Check that there are at least 3 parts for "abc": ["", "abc", ""]
    	    return parts[1]
    	}

	}

	return ""
}

// InitializeLexeme initializes the lexeme of the literal.
func (lm *LiteralModel) InitializeLexeme() {
    lm.lexeme.Symbol = lm.GetText()
    lm.lexeme.SymbolType = Terminal
}

// A literal is valid if its lexeme is valid and its text is not empty.
func (lm *LiteralModel) IsValid() bool {

    if !lm.TermModel.IsValid() {  // Call the parent method
        return false
    }

    return lm.GetText() != ""  // Checks that the text is not empty
}

/* TermModel */

// TermModel represents a term in an EBNF expression.
type TermModel struct {
	NotifiedModel
	*ebnf.Term
	lexeme *RuleModel
}

// CreateFrom initializes a TermModel from an EBNF node.
func (tm *TermModel) CreateFrom(node ebnf.Node) ITermModel {
	termNode, ok := node.(*ebnf.Term)

	if !ok {
		return nil
	}

	tm.Term = termNode
	tm.lexeme = &RuleModel{}
	return tm
}

func (tm *TermModel) DoAccept(visitor IVisitor) {

	switch v := visitor.(type) {
	case IParseTerm:
		v.Visit(tm)
	default:
		tm.VisitedModel.DoAccept(v)
	}

}

// GetLexeme returns the lexeme of the term.
func (tm *TermModel) GetLexeme() IRuleModel {
	return tm.lexeme  // Implicit cast from *RuleModel to IRuleModel
}

// GetText implements interfaces.TextProvider.GetText.
func (tm *TermModel) GetText() string {

	if tm.Term != nil {
		return tm.Term.String()
	}

	return ""
}

// This method is conceptually abstract, but must be implemented to satisfy the 
// interface. Descendants can override it.
func (tm *TermModel) IsValid() bool {
    
	if tm == nil || tm.lexeme == nil {
        return false
    }
    
	return tm.lexeme.IsValid()
}

// SetLexeme defines the lexeme of the term.
func (tm *TermModel) SetLexeme(lexeme IRuleModel) {
    tm.lexeme = lexeme.(*RuleModel)  // Explicit cast to concrete type
}