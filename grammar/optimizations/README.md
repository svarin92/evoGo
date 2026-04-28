# Package `optimizations`

**Go package for optimizing evolving grammars.**

This package provides tools to **simplify, factorize, and optimize** grammar rules within the **Genomizer** project. It is designed to be integrated into evolving grammar systems, where rules can be modified dynamically.

---

## **Features**
   Fonction                     | Description                                                         |
 |------------------------------|---------------------------------------------------------------------|
 | `ExpandRepetition`           | Develop the repetitions (`*`, `+`, `?`) in the productions.         |
 | `SimplifyRepetition`         | Replace non-terminals with their productions and remove duplicates. |
 | `LeftFactorize`              | Apply left factorization to eliminate redundancies in the rules.    |
 | `DirectLeftRecurse`          | Eliminates forward left recursion in a rule.                        |
 | `IndirectLeftRecurse`        | Detects and resolves indirect left recursiveness.                   |
 | `RemoveDuplicates`           | Removes duplicate productions from a production list.               |
 | `FindLongestCommonPrefix`    | Find the longest common prefix among a set of productions.          |

---

## **Installation**

Add this package to your Go project by importing the appropriate path:
```go
import "evoGo/grammar/optimizations"
```

## **Usage**

### 1. Simplifying repetitions
Replace non-terminals with their productions and remove duplicates.
```go
rules := map[string]model.IRuleModel{
    "X": model.NewRuleModel("X", model.NonTerminal, [][]model.IRuleModel{
        {model.NewRuleModel("a", model.Terminal, nil)},
        {model.NewRuleModel("b", model.Terminal, nil)},
    }),
}

rhs := [][]model.IRuleModel{
    {model.NewRuleModel("X", model.NonTerminal, nil)},
    {model.NewRuleModel("c", model.Terminal, nil)},
}

simplified := optimizations.SimplifyRepetition(rhs, rules)
// Résultat : [[a], [b], [c]]
```

### 2. Left factorization
Apply left factorization to eliminate redundancies.
```go
rule := model.NewRuleModel("A", model.NonTerminal, [][]model.IRuleModel{
    {model.NewRuleModel("a", model.Terminal, nil), model.NewRuleModel("b", model.Terminal, nil)},
    {model.NewRuleModel("a", model.Terminal, nil), model.NewRuleModel("c", model.Terminal, nil)},
})

rules := map[string]model.IRuleModel{"A": rule}
factorizedRule := optimizations.LeftFactorize(rule, rules)
// Result: A → aA_tail, A_tail → b | c
```

### 3. Development of repetitions
Develop the repetitions (*, +, ?) in the productions.
```go
repetition := "*"
rhs := [][]model.IRuleModel{
    {model.NewRuleModel("a", model.Terminal, nil)},
}

expanded := optimizations.ExpandRepetition(repetition, rhs, rules)
// Result: Develops repetition according to EBNF semantics.
```

## **Examples**

### Exemple 1 : Simplification
```go
// Grammar: X → a | b
rules := map[string]model.IRuleModel{
    "X": model.NewRuleModel("X", model.NonTerminal, [][]model.IRuleModel{
        {model.NewRuleModel("a", model.Terminal, nil)},
        {model.NewRuleModel("b", model.Terminal, nil)},
    }),
}

// Production to be simplified: [X, c]
rhs := [][]model.IRuleModel{
    {model.NewRuleModel("X", model.NonTerminal, nil), model.NewRuleModel("c", model.Terminal, nil)},
}

simplified := optimizations.SimplifyRepetition(rhs, rules)
// Résult : [[a, c], [b, c]]
```

### Exemple 2 : Factorization
```go
// Initial rule: A → ab | ac
rule := model.NewRuleModel("A", model.NonTerminal, [][]model.IRuleModel{
    {model.NewRuleModel("a", model.Terminal, nil), model.NewRuleModel("b", model.Terminal, nil)},
    {model.NewRuleModel("a", model.Terminal, nil), model.NewRuleModel("c", model.Terminal, nil)},
})

rules := map[string]model.IRuleModel{"A": rule}
factorizedRule := optimizations.LeftFactorize(rule, rules)
// Résult: A → aA_tail, A_tail → b | c
```

## **Edge cases and best practices**

### 1. Managing circular dependencies
Use IndirectLeftRecurse to detect and resolve circular dependencies between non-terminals.
```go
optimizations.IndirectLeftRecurse(rules)
```
### 2. Avoid duplicates
RemoveDuplicates uses a signature key including the symbol and its type to avoid false positives.

### 3. Unit tests
Always test your optimizations with simple and complex cases.
Example test for SimplifyRepetition:
```go
func TestSimplifyRepetition(t *testing.T) {
    rules := map[string]model.IRuleModel{
        "X": model.NewRuleModel("X", model.NonTerminal, [][]model.IRuleModel{
            {model.NewRuleModel("a", model.Terminal, nil)},
            {model.NewRuleModel("b", model.Terminal, nil)},
        }),
    }

    rhs := [][]model.IRuleModel{
        {model.NewRuleModel("X", model.NonTerminal, nil)},
        {model.NewRuleModel("c", model.Terminal, nil)},
    }

    simplified := optimizations.SimplifyRepetition(rhs, rules)

    if len(simplified) != 3 {
        t.Errorf("Expected 3 productions, got %d", len(simplified))
    }

}
```

## **Architecture and integration**

### 1. Integration with Genomizer
This package is designed for use with the Genomizer project, where grammars are dynamically modified by evolutionary mechanisms.

### 2. Dependencies
* evoGo/model: Provides data structures for rules and symbols.
* golang.org/x/exp/slices: Used for slice operations.

## **Contributing**

### 1. Report a bug
Open an issue on the GitHub repository with a clear description of the problem.

### 2. Suggest an improvement
* Fork the repository.
* Create a branch for your feature.
* Submit a Pull Request.

## **Licence**
This project is distributed under the [MIT](https://opensource.org/licenses/MIT).
© Stéphane Varin, 2026.

## **Acknowledgments**
* Thanks to the Go community for the open-source tools and libraries.
* Inspiration drawn from work on evolutionary grammars and factorization algorithms.


