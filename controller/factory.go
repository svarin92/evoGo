// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"sync"
	
	"evoGo/patterns/algo"
	"evoGo/patterns/builder"
)

/* Exports */

func NewGenomizer(grammar IGrammar) IGenomizer {
	return new(Genomizer).Create(grammar)
}

// NewGenomizerImmuneAdapter creates an adapter for the Genomizer.
func NewGenomizerImmuneAdapter(genomizer IGenomizer) *GenomizerImmuneAdapter {

	if genomizer == nil {
		panic("genomizer cannot be nil")
	}

    return &GenomizerImmuneAdapter{genomizer: genomizer}
}

func NewLinguisticPatternLibrary() *LinguisticPatternLibrary {
    return &LinguisticPatternLibrary{
        PatternsByCodons:   make(map[string]LinguisticPattern),
        PatternsBySemantics: make(map[string][]LinguisticPattern),
    }
}

// NewSerializer creates a new instance of Serializer.
func NewSerializer() ISerializer {
	return new(Serializer).Create(
		func() IAlgo { return algo.NewAlgo() },
		func() IBuilder { return builder.NewBuilder() },
	)
}

// Create and initialize a new cache for reduction strings.
func NewReductionChainCache() *ReductionChainCache {
    return &ReductionChainCache{
        cache: make(map[string][][]IRuleModel),  // List of reduction strings by symbol.
        mutex: sync.Mutex{},
    }
}