# Package `model`

## Description
The `model` package provides the **data structures** and **domain models** to 
represent EBNF (Extended Backus-Naur Form) rules and expressions. These models 
are designed for use with the **Visitor**, **Builder**, and **Notifier** 
patterns, and are initialized **late** by the `Serializer` during serialization.

This package plays a central role in **evoGo**, particularly for the **Genomizer**, 
where **EBNF sequences** (like `SequenceModel`) are treated as **codons**: basic
units for building and evolving grammars. These codons can be manipulated, 
recombined, or mutated to simulate **natural selection** within the framework of 
evolutionary grammars.

## Key Features

### 1. Hierarchy of types
- **Type Hierarchy**: `TermModel` acts as a base class for derived types :
  - `LiteralModel` : Represents a literal (ex: "abc"`).
  - `IdentifierModel` : Represents an identifier (e.g., `letter`).
  - `SubExpressionModel` : Represents a sub-expression (e.g., `(a b)`).
- **`NotifiedModel`** is inherited by all models to support the **Notifier Pattern**
  (post-visit notifications).
- **`SequenceModel`** represents a **syntagm** (or codon): a sequence of EBNF rules 
  that can evolve within the **Genomizer** framework.

### 2. Late initialization 
Fields like `lexeme` or `Symbols` are populated **during serialization**, via 
methods like `Serializer.DoHandle*`. This allows for flexible and decoupled 
model construction.

### 3. Pattern Integration
- **Visitor Pattern**: Each model implements DoAccept(visitor IVisitor) to allow 
traversal, validation, or transformation operations.
- **Notifier Pattern**: Thanks to the inheritance from NotifiedModel, models can 
trigger notifications after a visit.

### 4. Validation
Each model exposes an IsValid() method to check its validity. This includes:
- The presence of required fields (e.g., Symbol for RuleModel).
- The validity of substructures (e.g., Symbols for ExpressionModel).

## Main models
   Model                  | Role                                                        | EBNF Example              |
 |------------------------|-------------------------------------------------------------|---------------------------|
 | `EBNFModel`            | Represents a complete EBNF model (entire grammar).          | EBNF = Production*        |
 | `RuleModel`            | Represents a production rule.                               | letter = "a" "b"          |
 | `ExpressionModel`      | Represents an EBNF (sequence alternatives) expression.      | Expression1 | Expression2 |
 | `SequenceModel`        | Represents a sequence of terms (a codon for the Genomizer). | a b c                     |
 | `SubExpressionModel`   | Represents a sub-expression.                                | (a b)                     |
 | `TermModel`            | Base class for terms (literals, identifiers, groups, etc.). | "a", letter, (a)          |
 | `LiteralModel`         | Represents a literal.                                       | "a"                       |
 | `IdentifierModel`      | Represents an identifier.                                   | letter                    |

## Example of Use

### 1. Creating a rule model
```go
ruleModel := &RuleModel{
    Symbol:    "letter",
    SymbolType: NonTerminal,
}
ruleModel.CreateFrom(ebnfNode)  // Initialization from an EBNF node
```
### 2. Validating an expression
```go
if !expressionModel.IsValid() {
    log.Error("Expression invalide")
}
```
### 3. Use with a visitor
```go
visitor := &ParseRuleVisitor{
    Algo: func(data any) {
        rule := data.(*RuleModel)
        fmt.Println("Visiting rule:", rule.Symbol)
    },
}
ruleModel.Accept(visitor)
```
### 4. Integration with the Notifier Pattern
```go
notifiedRule := &RuleModel{
    NotifiedModel: NotifiedModel{},  // Héritage pour les notifications.
    Symbol:        "notified_rule",
}
notifiedRule.Accept(visitor, func() {
    fmt.Println("Notification : visite terminée !")
})
```

## License
This project is distributed under the [MIT](LICENSE).
© Stéphane Varin, 2026.