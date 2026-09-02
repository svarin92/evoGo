# Cell Package - Cellular Architecture for evoGo

## Overview

The `cell` package implements a **cellular architecture** for evoGo, hybridizing the existing Grammatical Evolution (GE) engine with the **Actor model** from `go-actor`. This creates a distributed, self-organizing system where individuals (cells) can communicate, exchange genetic material, and collectively evolve towards solutions.

## Key Concepts

### DigitalCell

A `DigitalCell` represents a cellular agent with:
- **Nucleus**: The core containing the genome and phenotype (based on `model.Individual`)
- **Membrane**: The cellular interface for communication and filtering
- **PID**: Unique identifier in the actor system

### Membrane

The membrane implements three critical functions:

1. **Excrétion (Horizontal Gene Transfer)**: Sending genetic material (codon fragments) to other cells
2. **Absorption (Immune Filtering)**: Receiving and validating genetic material through the immune system
3. **Signalisation (Quorum Sensing)**: Broadcasting fitness signals to influence neighbor behavior

### Message Types

| Message Type | Purpose | Direction |
|-------------|---------|-----------|
| `MsgTick` | Metabolism/evolution trigger | System → Cell |
| `MsgGeneticMaterial` | Horizontal gene transfer | Cell → Cell |
| `MsgSignal` | Quorum sensing (stress/stability) | Cell → Neighbors |
| `MsgCrossoverRequest` | Request for genetic crossover | Cell → Cell |
| `MsgCrossoverResponse` | Response to crossover request | Cell → Cell |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CellSystem                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ CellActor 1 │  │ CellActor 2 │  │ CellActor 3 │         │
│  │  ┌─────────┐│  │  ┌─────────┐│  │  ┌─────────┐│         │
│  │  │DigitalCell││  │  │DigitalCell││  │  │DigitalCell││         │
│  │  │ ┌───────┐││  │  │ ┌───────┐││  │  │ ┌───────┐││         │
│  │  │ │Nucleus│││  │  │ │Nucleus│││  │  │ │Nucleus│││         │
│  │  │ └───────┘││  │  │ └───────┘││  │  │ └───────┘││         │
│  │  │Membrane ││  │  │Membrane ││  │  │Membrane ││         │
│  │  └─────────┘│  │  │ └───────┘│  │  │ └───────┘│         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Usage

### Basic Example

```go
package main

import (
    "evoGo/cell"
    "evoGo/model"
)

func main() {
    // Create cell system
    system := cell.NewCellSystem()
    
    // Create a cell with a random genome
    genome := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
    individual := model.NewIndividual(genome)
    membrane := cell.NewMembrane()
    
    // Create cell actor
    cellActor, err := system.CreateCell(individual, membrane)
    if err != nil {
        panic(err)
    }
    
    // Send a tick message to trigger evolution
    tickMsg := cell.NewMsgTick(0, 1)
    cellActor.Send(cellActor, tickMsg)
    
    // Shutdown when done
    system.Shutdown()
}
```

### Creating a Cellular Population

```go
// Create a population of 10 cells
system := cell.NewCellSystem()
for i := 0; i < 10; i++ {
    genome := generateRandomGenome(50)
    individual := model.NewIndividual(genome)
    membrane := cell.NewMembraneWithParams(0.5, 0.7, immuneSys)
    system.CreateCell(individual, membrane)
}

// Connect all cells
system.ConnectAll()

// Run simulation
for gen := 0; gen < 10; gen++ {
    system.SendTickToAll(gen)
    time.Sleep(100 * time.Millisecond)
}
```

## Membrane Configuration

The membrane can be configured with different parameters:

```go
// Create a highly permeable membrane (for exploration)
membrane := cell.NewMembraneWithParams(
    0.8,  // High permeability
    0.5,  // Low fitness threshold
    immuneSys,
)

// Create a restrictive membrane (for exploitation)
membrane := cell.NewMembraneWithParams(
    0.2,  // Low permeability
    0.9,  // High fitness threshold
    immuneSys,
)
```

## Quorum Sensing

Cells automatically broadcast signals based on their fitness:

- **Stress Signal** (fitness < 0.4): Neighbors increase mutation rate
- **Stability Signal** (fitness > 0.8): Neighbors reduce mutation rate
- **Neutral** (0.4 ≤ fitness ≤ 0.8): No change

The membrane adjusts its permeability and sensitivity based on received signals.

## Integration with evoGo

To integrate with the existing evoGo evolutionary loop:

1. Replace `Population` with `CellSystem`
2. Replace `Individual` with `DigitalCell`
3. Use `SendTickToAll()` instead of the traditional generation loop
4. The membrane handles communication between cells

## Files

| File | Purpose |
|------|---------|
| `cell.go` | DigitalCell implementation |
| `membrane.go` | Membrane implementation with filtering |
| `messages.go` | Message type definitions |
| `actor.go` | go-actor integration |
| `factory.go` | Factory functions for creating cells |
| `examples/` | Example usage |

## Dependencies

- `github.com/vladopajic/go-actor` - Actor model library
- `evoGo/interfaces` - Core interfaces
- `evoGo/model` - Individual model
- `evoGo/controller` - Genomizer and immune system

## Future Enhancements

- [ ] Dynamic topology (grid, torus, small-world networks)
- [ ] Adaptive membrane permeability based on population diversity
- [ ] Specialized cell types (stem cells, immune cells, etc.)
- [ ] Chemical gradient-based communication
- [ ] Spatial organization and neighborhood radii
