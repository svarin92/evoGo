// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// The Levenshtein distance is the minimum number of operations (insertions, 
// deletions, or substitutions) required to transform s1 into s2.
// Example:
//   distance := LevenshteinDistance("kitten", "sitting") // Returns 3
func LevenshteinDistance(s1, s2 string) int {
    m := len(s1)
    n := len(s2)
    dp := make([][]int, m+1)

    for i := range dp {
        dp[i] = make([]int, n+1)
    }

    for i := 0; i <= m; i++ {
        dp[i][0] = i
    }

    for j := 0; j <= n; j++ {
        dp[0][j] = j
    }

    for i := 1; i <= m; i++ {
    
        for j := 1; j <= n; j++ {
    
            if s1[i-1] == s2[j-1] {
                dp[i][j] = dp[i-1][j-1]
            } else {
                dp[i][j] = 1 + Min(
                    dp[i-1][j],    // delete
                    dp[i][j-1],    // insert
                    dp[i-1][j-1],  // substitution
                )
            }
    
        }
    
    }
    
    return dp[m][n]
}

// Min returns the smallest value from a list of integers. Returns 0 if the 
// list is empty.
// Example:
//   minimum := Min(1, 2, 3) // Returns 1
func Min(values ...int) int {
    if len(values) == 0 {
        return 0
    }

    m := values[0]
    
    for _, v := range values[1:] {
    
        if v < m {
            m = v
        }
    
    }
    
    return m
}
