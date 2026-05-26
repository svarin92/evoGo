// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

import "evoGo/interfaces"

// DeepCopyProductionHistory performs a deep copy of production history.
func DeepCopyProductionHistory(history [][]interfaces.IRuleModel) [][]interfaces.IRuleModel {

    if history == nil {
        return nil
    }

    newHistory := make([][]interfaces.IRuleModel, len(history))
    
    for i, prod := range history {
        newHistory[i] = make([]interfaces.IRuleModel, len(prod))
    
        for j, rule := range prod {
            newHistory[i][j] = rule.Clone()
        }
    
    }
    
    return newHistory
}