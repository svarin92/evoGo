# Visitor Pattern

## Description
The Visitor Pattern allows you to add new operations to objects without 
modifying their structure. This pattern is particularly useful for:
- Traversing and transforming complex structures (trees, graphs);
- Adding functionality dynamically without modifying existing classes;
- Decouple algorithms from the data structures on which they operate.
This package implements the pattern in Go with generics, ensuring type safety 
and flexibility.

## Key Components
   Component         | Role                                                    |
 |-------------------|---------------------------------------------------------|
 | `IVisitable[T, U]`| Interface for "visitable" objects.                      |
 | `IVisitor`        | Interface for "visitor" objects.                        |
 | `IVisited[T]`     | Extendable and adaptable to support chaining operations |
 |                   | or lazy evaluation.                                     |
 | `VisitedModel`    | Concrete implementation of a visitable object.          |
 | `ModelVisitor`    | Concrete implementation of a visitor, using a fonction. |

## Example of Use

### 1. Define a concrete visitor
```go
visitor := &ModelVisitor{
    algo: func(data any) {
        fmt.Println("Visiteur : Data processing", data)
    },
}
visited := &VisitedModel{}
visited.Accept(visitor)
```
### 2. Create a visitable object
```go
visited := (&VisitedModel{}).Create()
```
### 3. Accepter le visiteur
```go
visited.Accept(visitor)
```
### 4. Sequence of operations (optional)
```go
visited.Accept(
    visitor,
    func() IVisited[IVisitor] {
        return (&VisitedModel{}).Create()
    },
)
```
## Integration with other patterns

- Notify: The Visitor can trigger notifications after a visit (e.g., logs, 
  updates).
- Builder: Can be combined with the Builder to validate rules during object 
  construction.
