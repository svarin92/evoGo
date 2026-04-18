// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

// ToFloat64 converts a value to float64. Returns (value, true) if the 
// conversion is successful, (0, false) otherwise. The supported types are: 
// int, int32, int64, float32, float64.
// Example:
//   val, ok := ToFloat64(42) // Returns (42.0, true)
//   val, ok := ToFloat64("abc") // Returns (0, false)
func ToFloat64(v any) (float64, bool) {

    switch t := v.(type) {
    case int: return float64(t), true
    case int32: return float64(t), true
    case int64: return float64(t), true
    case float32: return float64(t), true
    case float64: return t, true
    default: return 0, false
    }
	
}