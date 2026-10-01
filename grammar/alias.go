// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package grammar

import "github.com/svarin92/evoGo/interfaces"

// Aliases for functional types.
type (
	ParserFactory     = interfaces.ParserFactory
	SerializerFactory = interfaces.SerializerFactory
)

// Interfaces imported to ensure architectural consistency.
type (
	IGrammar          = interfaces.IGrammar
	IParser           = interfaces.IParser
	IRuleModel        = interfaces.IRuleModel
	ISerializer       = interfaces.ISerializer
)