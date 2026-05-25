// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package algo

/* Exports */

// NewAlgo creates a new instance of IAlgo via AlgoMaker.
func NewAlgo() IAlgo {
	return new(AlgoMaker).Create()
}