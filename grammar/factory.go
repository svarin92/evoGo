// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package grammar

import (
	"fmt"
	"io"

	"github.com/alecthomas/kong"

	"evoGo/controller"
)

/* Exports */

func NewParser() IParser {
	return new(EBNFParser).Create()
}

func NewSpeciesGrammar(ctx *kong.Context, reader io.Reader) (IGrammar, error) {
	sg := new(SpeciesGrammar).Create(
		func() ISerializer { return controller.NewSerializer() },
		ctx,
		reader,
	)

	 if err := sg.Segment(); err != nil {
        return nil, fmt.Errorf("failed to segment grammar: %w", err)
    }

	return sg, nil
}