// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package controller

import (
	"evoGo/patterns/algo"
	"evoGo/patterns/builder"
)

/* Exports */

// NewSerializer creates a new instance of Serializer.
func NewSerializer() ISerializer {
	return new(Serializer).Create(
		func() IAlgo { return algo.NewAlgo() },
		func() IBuilder { return builder.NewBuilder() },
	)
}