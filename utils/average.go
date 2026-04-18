// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

// Average calculates the average of a list of float64 values. Returns 0.0 if 
// the list is empty. 
// Example:
//   values ​​:= []float64{1.0, 2.0, 3.0}
//   avg := Average(values) // Returns 2.0
func Average(values []float64) float64 {

    if len(values) == 0 {
        return 0.0
    }

    sum := 0.0
    
	for _, v := range values {
        sum += v
    }
    
	return sum / float64(len(values))
}
