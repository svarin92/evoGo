# Builder Pattern

## Description
The **Builder Pattern** allows you to build **EBNF** rule models from an **AST** 
(Abstract Syntax Tree). This pattern is designed to be used with **Visitor** and 
**Notifier Patterns**, to enable transformations, validations, or notifications 
during rule construction.

Within the **evoGo** project, the Builder plays a key role in:
- **Building evolving grammars** by assembling rules, expressions, and sequences
  (codons).
- **Integrating visitors** to apply validation or transformation algorithms during 
  construction.
- **Managing notifications** to trigger actions after rule construction.

## Key components

| Component         | Role                                                                        |
|-------------------|-----------------------------------------------------------------------------|
| `IBuilder`        | Interface defining the methods for building EBNF rule models.               |
| `RuleBuilder`     | Concrete implementation of the Builder. Handles term groups and constructs. |

## Fonctionnalités Principales

### 1. Model Building
The Builder allows you to construct different types of EBNF models:
- **Rules** (`BuildRule`): Example: `letter = "a" | "b"`.
- **Expressions** (`BuildExpression`): Example: `"a" | "b"`.
- **Sequences** (`BuildSequence`): Example: `"a" "b" "c"` (codon).
- **Subexpressions** (`BuildSubExpression`): Example: `("a" | "b")`.
- **Terms** (`BuildTerm`): Example: `"a"` or `letter`.

### 2. Modifying Terms
The Builder offers two methods for modifying term groups:
- **`ModifyTerms`**: For public use (via `IRuleModel` interfaces).
- **`ModifyConcreteTerms`**: For internal use (with concrete types `*RuleModel`
  and filtering of invalid groups).

### 3. Validation
Each constructed model can be validated via the `IsValid()` method (e.g., 
checking for non-empty groups and valid rules).

## Example of Use
### 1. Création d'un Builder
```go
// Create a new RuleBuilder instance.
builder := NewBuilder()
```
### 2. Constructing an Expression
```go
// Define a visitor to apply transformations.
visitor := &ModelVisitor{
    Algo: func(data any) {
        fmt.Println("Visitor : Data processing", data)
    },
}

// Construct an expression from an EBNF node.
exprSymbols, err := builder.BuildExpression(ebnfNode, visitor)

if err != nil {
    log.Fatal("Error constructing expression:", err)
}

// exprSymbols contains the expression rule groups.
for i, group := range exprSymbols {
    fmt.Printf("Group %d : %v\n", i, group)
}
```
### 3. Construction of a Rule
```go
// Construct a rule from an EBNF node.
ruleSymbol, err := builder.BuildRule(ebnfNode, visitor)

if err != nil {
    log.Fatal("Error during the construction of the rule:", err)
}

fmt.Println("Rule symbol:", ruleSymbol)
```
### 4. Modification of Terms
```go
// Add a group of terms.
rule1 := &model.RuleModel{Symbol: "rule1"}
rule2 := &model.RuleModel{Symbol: "rule2"}
builder.AddTermGroup([]*model.RuleModel{rule1})

// Modify the terms by adding a new group.
builder.ModifyTerms(func(terms [][]interfaces.IRuleModel) [][]interfaces.IRuleModel {
    newGroup := []interfaces.IRuleModel{rule2}
    return append(terms, newGroup)
})

// Retrieve the modified terms.
terms := builder.GetTerms()

for i, group := range terms {
    fmt.Printf("Groupe %d : ", i)
    
    for _, rule := range group {
        fmt.Printf("%s ", rule.GetIdentifier())
    }
    
    fmt.Println()
}
```
### 5. Modification of Specific Terms (with filtering)
```go
// Modify the concrete terms with filtering of invalid groups.
builder.ModifyConcreteTerms(func(terms [][]*model.RuleModel) [][]*model.RuleModel {
    
    // Remove empty or invalid groups.
    var validTerms [][]*model.RuleModel

    for _, group := range terms {

        if builder.IsValidTermGroup(group) {
            validTerms = append(validTerms, group)
        }

    }

    return validTerms
})
```

## Licence
This project is distributed under the [MIT](https://opensource.org/licenses/MIT).
© Stéphane Varin, 2026.
