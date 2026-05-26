// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

import (
	"strings"

	"evoGo/interfaces"
)

// RuleModelSliceToString converts a list of RuleModel to a string.
func RuleModelSliceToString(rules []interfaces.IRuleModel) string {
	var sb strings.Builder

	for i, rule := range rules {

		if i > 0 {
			sb.WriteString(" ")  // Add a space between symbols
		}

		sb.WriteString(rule.GetText())
	}

	return sb.String()
}