// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package utils

import "evoGo/interfaces"

// DeepCopyDynamicRules creates a deep copy of DynamicRules.
func DeepCopyDynamicRules(rules map[string]interfaces.IRuleModel) map[string]interfaces.IRuleModel {
    copiedRules := make(map[string]interfaces.IRuleModel, len(rules))

    for k, v := range rules {
        copiedRules[k] = v.Clone()
    }
    
    return copiedRules
}

func DeepCopyMap(original map[string]interfaces.IRuleModel) map[string]interfaces.IRuleModel {
    copy := make(map[string]interfaces.IRuleModel, len(original))

    for k, v := range original {
        copy[k] = v  // Assume that IRuleModel is copyable by value.
    }
    
    return copy
}

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

func DeepCopyStringSlice(original []string) []string {
    copiedSlice := make([]string, len(original))
    copy(copiedSlice, original) 
    return copiedSlice
}