// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The interfaces package defines the contracts (interfaces) for the key 
// components of the project. This package allows for strong decoupling 
// between implementations and their use, facilitating testing, maintenance, 
// and extensibility.
package interfaces

import (
	"io"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/participle/v2/ebnf"
)

/* Parser */

// IParser defines the contract for a grammar parser. A parser is responsible 
// for parsing a grammar file (e.g., EBNF) and constructing an abstract syntax 
// tree (AST). In the context of the Genomizer, IParser is used to:
// - Load grammars from files (e.g., letter.bnf).
// - Build an AST that will then be serialized into production rules.
// - Enable parser exchange (e.g., converting from EBNF to JSON).
type IParser interface {	

	// Parse analyzes a grammar file and builds an AST. This method is the 
	// core of the IParser contract.
	//
	// Parameters:
	// - ctx *kong.Context: CLI context for error handling and options.
	// - reader io.Reader: Source of the file to parse.
	// - ast *ebnf.Node: Pointer to the AST to populate.
	//
	// Returns:
	// - error: An error if the parsing fails.
	Parse(ctx *kong.Context, reader io.Reader, ast *ebnf.Node) error
}

// ParserFactory is a factory function for creating IParser instances. Used to 
// decouple parser creation from usage.
type ParserFactory func() IParser