# Package Algo

## Description
The `algo` package provides an **algorithm factory** for creating visitors specialized in parsing EBNF models. It extends the **Visitor Pattern** to offer configurable and reusable algorithms for traversing, validating, or transforming EBNF rule structures.

This package is designed to be used with the models defined in the `model` package (e.g., `RuleModel`, `ExpressionModel`) and can be integrated with other patterns like **Builder** or **Notifier** for advanced processing.

## Key Components
   Component               | Role                                              |
 |-------------------------|---------------------------------------------------|
 | `IAlgo`                 | Interface for the algorithm factory.              |
 | `AlgoMaker`             | Concrete implementation of the factory.           |
 | `IParseEBNF`            | Specialist visitor for complete EBNF models.      |
 | `ParseEBNF`             | Concrete implementation for EBNF models.          |
 | `IParseExpression`      | Specialized visitor for EBNF expressions.         |
 | `ParseExpression`       | Concrete implementation for expressions.          |
 | `IParseRule`            | Specialist visitor for EBNF rules.                |
 | `ParseRule`             | Concrete implementation for the rules.            |
 | `IParseSequence`        | Specialized visitor for EBNF sequences.           |
 | `ParseSequence`         | Concrete implementation for the sequences.        |
 | `IParseSubExpression`   | Specialized visitor for EBNF sub-expressions.     |
 | `ParseSubExpression`    | Concrete implementation for sub-expressions.      |
 | `IParseTerm`            | Specialist visitor for EBNF terms.                |
 | `ParseTerm`             | Concrete implementation for the terms.            |

## Example of Use

### 1. Creating a parsing algorithm for expressions
```go
import (
    "evoGo/patterns/algo"
    "evoGo/patterns/visitor"
    "evoGo/model"
)

func main() {

    // Create an algorithm factory.
    maker := algo.NewAlgo()

    // Create a parsing algorithm for expressions.
    parseExpr := maker.MakeExpressionCase(
        visitor.VisitorFunc(func(data any) {
            fmt.Printf("Visiting expression: %v\n", data)
        }),
    )

    // Create rule models for the expression.
    ruleA := &model.RuleModel{Symbol: "a", SymbolType: interfaces.Terminal}
    ruleB := &model.RuleModel{Symbol: "b", SymbolType: interfaces.Terminal}

    // Create an expression model.
    exprModel := &model.ExpressionModel{}
    exprModel.Symbols = [][]interfaces.IRuleModel{
        {ruleA},
        {ruleB},
    }

	// Apply the algorithm to the model.
	parseExpr.Visit(exprModel)
}
```
### 2. Creating a parsing algorithm for the rules
```go
func main() {
    maker := algo.NewAlgo()

    // Create a parsing algorithm for the rules.
    parseRule := maker.MakeRuleCase(
        visitor.VisitorFunc(func(data any) {
            rule := data.(*model.RuleModel)
            fmt.Printf("Visiting rule: %s (Type: %d)\n", rule.Symbol, rule.SymbolType)
        }),
    )

    // Use the algorithm to visit a rule pattern.
    ruleModel := &model.RuleModel{
        Symbol:  "example_rule",
        SymbolType: model.NonTerminal,
    }
    parseRule.Visit(ruleModel)
}
```
### 3. Integration with the Builder pattern
```go
import (
    "evoGo/patterns/builder"
    "evoGo/patterns/algo"
)

func main() {

    // Create a builder and a parsing algorithm.
    ruleBuilder := builder.NewBuilder()
    maker := algo.NewAlgo()

    // Create an algorithm to validate the rules.
    validateRule := maker.MakeRuleCase(
        visitor.VisitorFunc(func(data any) {
            rule := data.(*model.RuleModel)
            
            if rule.Symbol == "" {
                panic("Rule symbol cannot be empty!")
            }

        }),
    )

    // Construct a rule with validation.
    node := &ebnf.Production{Production: "valid_rule"}
    rule, err := ruleBuilder.BuildRule(node, validateRule)
    
    if err != nil {
        fmt.Println("Error building rule:", err)
    }

}
```
### 4. Integration with the Notifier pattern
```go
import (
    "evoGo/patterns/notifier"
    "evoGo/patterns/algo"
)

func main() {
    maker := algo.NewAlgo()

    // Create an algorithm with notification.
    logRuleVisit := maker.MakeRuleCase(
        visitor.VisitorFunc(func(data any) {
            fmt.Println("Rule visited:", data)
        }),
    )

    // Create a notifiable model.
    notifiedRule := &model.RuleModel{
        NotifiedModel: notifier.NotifiedModel{},
        Symbol:        "notified_rule",
    }

    // Apply the algorithm with notification.
    notifiedRule.Accept(logRuleVisit, func() notifier.IVisited[visitor.IVisitor] {
        fmt.Println("Notification: Rule visit completed!")
        return notifiedRule
    })
}
```

## **Licence**
This project is distributed under the [MIT](https://opensource.org/licenses/MIT). 
© Stéphane Varin, 2026.