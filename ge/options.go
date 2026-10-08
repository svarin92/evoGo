// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package ge

type SearchOption func(*SearchConfig)

type SearchConfig struct {
    hook IHybridizationHook
}

// WithHybridizationHook attaches an external module (e.g., evocell) to the 
// population created by SearchLoop — the ONLY injection point, since the 
// population is created inside the loop.
func WithHybridizationHook(hook IHybridizationHook) SearchOption {
    return func(c *SearchConfig) { c.hook = hook }
}