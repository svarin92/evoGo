// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package renderer

/* Exports */

// New Renderer creates a new Renderer with a grammar provider.
func NewRenderer(grammarProvider IGrammarProvider) IRenderer {
	return new(Renderer).Create(
		func() IGrammarProvider { return grammarProvider },
	)
}