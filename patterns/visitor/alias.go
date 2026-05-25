// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package visitor

import "evoGo/interfaces"

// Aliases for functional types.
type (
	VisitorFunc = interfaces.VisitorFunc
)

// Interfaces imported to ensure architectural consistency.
type (
	IVisitor = interfaces.IVisitor
	IVisited = interfaces.IVisited[IVisitor]
)
