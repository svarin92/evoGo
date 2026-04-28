# Notifier Pattern

## Description
The **Notifier Pattern** extends the **Visitor Pattern** to add a **notification mechanism after a visit**. It is useful for:
- **Triggering actions after a visit** (e.g., logging, updates).
- **Decoupling notification logic** from visit logic.
- **Chain visitors** with notification callbacks.

This pattern is implemented in Go and integrates with the visitor package to provide a flexible and extensible solution.

## Key Components

| Component         | Role                                                     |
|-------------------|----------------------------------------------------------|
| `INotifiedModel`  | Interface for notifiable objects.                        |
| `NotifiedModel`   | Concrete implementation of a notifiable object.          |

## Example of Use
```go
notified := &NotifiedModel{}
visitor := &ModelVisitor{
    algo: func(data any) {
        fmt.Println("Visitor: Data Processing", data)
    },
}
notified.Accept(
    visitor,
    func() IVisited[IVisitor] {
        fmt.Println("Notification : Visit complete!")
        return &VisitedModel{}
    },
)
```
## Integration with the Visitor Pattern
The **Notifier Pattern** relies on the **Visitor Pattern** to:
- Reuse existing visit logic (VisitedModel, ModelVisitor).
- Add a notification layer **without modifying** the visited objects.
- Allow for complex sequences of visitors and notifications.

## **Licence**
This project is distributed under the [MIT](https://opensource.org/licenses/MIT).
© Stéphane Varin, 2026.