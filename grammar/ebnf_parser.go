// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
//
// The grammar package provides implementations for parsing grammar files in 
// EBNF (Extended Backus-Naur Form) format. This package is designed for use
// with the Genomizer project's evolving grammar system, where grammars define 
// the production rules for generating phenotypes from genomes.
package grammar

import (
	"bytes"
	"fmt"
	"io"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/participle/v2/ebnf"
)

/* EBNFParser */

// EBNFParser is a concrete implementation of IParser for EBNF files. It 
// uses the particle/ebnf library to parse grammar files and build an 
// abstract syntax tree (AST).
//
// This parser is designed to be:
//   - **Decoupled**: It depends only on the IParser interface and not on 
//     other components.
//   - **Reusable**: Can be used by any component that needs to parse EBNF 
//     grammars.
//   - **Robust**: Explicitly handles read and parsing errors.
type EBNFParser struct {

}

// Create initializes a new instance of EBNFParser. This method is 
// primarily useful for testing or custom initialization. In most 
// cases, use the NewEBNFParser() factory instead.
func (p *EBNFParser) Create() *EBNFParser {
    return p
}

// DoParse allows separating context-related logic (e.g., configuration, 
// options) from pure parsing logic (ebnf.Parse). Currently, this method does 
// nothing, but it is being kept for a future extension (e.g., validation of 
// parsing options).
func (p *EBNFParser) DoParse(ctx *kong.Context) error {
	return nil
}

// Parse analyzes an EBNF file and builds an AST (Abstract Syntax Tree) in 
// the variable `ast`. This method is the core of the parser and follows these 
// steps:
// 1. Checks that the context (ctx) and the reader are valid.
// 2. Handles non-seekable readers (e.g., network streams) by loading them 
//    into memory.
// 3. Resets the reader if possible (to allow multiple reads).
// 4. Calls ebnf.Parse() to build the AST.
// 5. Stores the result in `ast`.
//
// Parameters:
// - ctx *kong.Context: CLI context for error handling and options.
// - reader io.Reader: Source of the EBNF file to parse (e.g., file, stream).
// - ast *ebnf.Node: Pointer to the AST to be populated. If nil, a new AST is created.
//
// Returns:
// - error: An error if:
// - The reader is nil.
// - Stream reading fails (for non-seekable readers).
// - EBNF parsing fails (e.g., invalid syntax).
func (p *EBNFParser) Parse(
	ctx *kong.Context, 
	reader io.Reader, 
	ast *ebnf.Node,
) error {
	if err := p.DoParse(ctx); err != nil {
		return err
	}

	// Reader check.
	if reader == nil {
        return fmt.Errorf("reader is nil")
    }

	// If the reader is not seekable, read all of its contents into memory.
    if _, ok := reader.(io.Seeker); !ok {
        data, err := io.ReadAll(reader)

        if err != nil {
            return fmt.Errorf("failed to read non-seekable reader: %v", err)
        }

		// Creates a seekable reader in memory.
        reader = bytes.NewReader(data)
    } else {
        
		// Rewind the reader if possible.
        if _, err := reader.(io.Seeker).Seek(0, io.SeekStart); err != nil {
            return fmt.Errorf("failed to rewind reader: %v", err)
        }

    }	

	// Parse the EBNF file.
	newAST, err := ebnf.Parse(reader)
	
	if err != nil {
        return fmt.Errorf("EBNF grammar parsing failed: %v", err)
    }
	
	// Stores the result in `ast`.
	if ast != nil {
		*ast = newAST
	}

	return nil
}
