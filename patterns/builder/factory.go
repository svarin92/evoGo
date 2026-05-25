// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package builder

/* Exports */

// NewBuilder creates and returns a new instance of RuleBuilder.
func NewBuilder() IBuilder {
	return new(RuleBuilder).Create()
}