// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package notifier

import "evoGo/interfaces"
import "evoGo/patterns/visitor"

// Interfaces imported to ensure architectural consistency.
type (
	IVisitor = interfaces.IVisitor
	IVisited = interfaces.IVisited[IVisitor]
)

// Aliases for concrete types.
type (
	VisitedModel = visitor.VisitedModel
)

