// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package cell

import (
	"context"
	"fmt"
	"log"
	"sync"

	"evoGo/interfaces"

	"github.com/vladopajic/go-actor"
)

// CellActor wraps a DigitalCell to work with the go-actor library.
// It implements the actor.Producer interface.
type CellActor struct {
	// cell is the underlying DigitalCell.
	cell *DigitalCell

	// pid is the actor's PID.
	pid *actor.PID

	// ctx is the actor context.
	ctx context.Context

	// cancel is the context cancellation function.
	cancel context.CancelFunc

	// wg is used to wait for goroutines to complete.
	wg sync.WaitGroup

	// neighbors is a map of neighbor PIDs to CellActor references.
	neighbors map[string]*actor.PID

	// mu protects the neighbors map.
	mu sync.RWMutex
}

// NewCellActor creates a new CellActor from a DigitalCell.
func NewCellActor(cell *DigitalCell) *CellActor {
	return &CellActor{
		cell:     cell,
		pid:      nil,
		ctx:      context.Background(),
		cancel:   nil,
		neighbors: make(map[string]*actor.PID),
	}
}

// CreateCellActor creates and registers a new CellActor with the actor system.
func CreateCellActor(cell *DigitalCell, system *actor.System) (*CellActor, error) {
	actor := NewCellActor(cell)
	
	// Create the actor
	pid, err := system.Spawn(actor.produce, "cell")
	if err != nil {
		return nil, fmt.Errorf("failed to spawn cell actor: %w", err)
	}

	actor.pid = pid
	actor.cell.SetPID(pid)
	
	// Set up context with cancellation
	actor.ctx, actor.cancel = context.WithCancel(context.Background())
	
	return actor, nil
}

// produce is the actor's message processing function.
// It implements the actor.Producer interface.
func (ca *CellActor) produce(ctx *actor.Context) {
	ca.ctx = ctx
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("CellActor.produce: Context cancelled for cell %v", ca.pid)
			return
		default:
			// Process incoming messages
			msg := ctx.Receive()
			if msg == nil {
				continue
			}

			// Handle the message
			if err := ca.handleMessage(msg); err != nil {
				log.Printf("CellActor.produce: Error handling message: %v", err)
			}
		}
	}
}

// handleMessage handles an incoming message.
func (ca *CellActor) handleMessage(msg any) error {
	// Convert the actor message to our internal message type
	// The message from go-actor is wrapped in an envelope
	
	// Check if it's a wrapped message
	if envelope, ok := msg.(actor.Message); ok {
		msg = envelope.Payload
	}

	// Handle the message using the cell's Receive method
	return ca.cell.Receive(msg)
}

// Send sends a message to another cell actor.
func (ca *CellActor) Send(to *CellActor, message any) error {
	if to == nil || to.pid == nil {
		return fmt.Errorf("invalid target actor")
	}

	// Send the message via the actor system
	ca.pid.Send(to.pid, message)
	return nil
}

// Broadcast sends a message to all neighboring cells.
func (ca *CellActor) Broadcast(message any) error {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	for _, neighborPID := range ca.neighbors {
		ca.pid.Send(neighborPID, message)
	}
	return nil
}

// AddNeighbor adds a neighbor to the cell's neighbor list.
func (ca *CellActor) AddNeighbor(neighbor *CellActor) error {
	if neighbor == nil || neighbor.pid == nil {
		return fmt.Errorf("invalid neighbor")
	}

	ca.mu.Lock()
	defer ca.mu.Unlock()

	pidStr := neighbor.pid.String()
	ca.neighbors[pidStr] = neighbor.pid
	ca.cell.AddNeighbor(neighbor.pid)
	
	return nil
}

// RemoveNeighbor removes a neighbor from the cell's neighbor list.
func (ca *CellActor) RemoveNeighbor(neighbor *CellActor) error {
	if neighbor == nil || neighbor.pid == nil {
		return fmt.Errorf("invalid neighbor")
	}

	ca.mu.Lock()
	defer ca.mu.Unlock()

	pidStr := neighbor.pid.String()
	delete(ca.neighbors, pidStr)
	ca.cell.RemoveNeighbor(neighbor.pid)
	
	return nil
}

// DiscoverNeighbors returns the PIDs of neighboring cells.
func (ca *CellActor) DiscoverNeighbors() []any {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	neighbors := make([]any, len(ca.neighbors))
	for _, pid := range ca.neighbors {
		neighbors = append(neighbors, pid)
	}
	return neighbors
}

// GetPID returns the cell's PID.
func (ca *CellActor) GetPID() any {
	return ca.pid
}

// GetCell returns the underlying DigitalCell.
func (ca *CellActor) GetCell() *DigitalCell {
	return ca.cell
}

// Stop stops the cell actor.
func (ca *CellActor) Stop() {
	if ca.cancel != nil {
		ca.cancel()
	}
}

// Wait waits for all goroutines to complete.
func (ca *CellActor) Wait() {
	ca.wg.Wait()
}

// CellSystem manages a population of cell actors.
type CellSystem struct {
	// system is the underlying go-actor system.
	system *actor.System

	// cells is a map of cell PIDs to CellActor references.
	cells map[string]*CellActor

	// mu protects the cells map.
	mu sync.RWMutex
}

// NewCellSystem creates a new CellSystem.
func NewCellSystem() *CellSystem {
	return &CellSystem{
		system: actor.NewSystem(actor.NewSystemConfig()),
		cells: make(map[string]*CellActor),
	}
}

// CreateCell creates and registers a new cell in the system.
func (cs *CellSystem) CreateCell(nucleus interfaces.IIndividual, membrane interfaces.IMembrane) (*CellActor, error) {
	// Create the digital cell
	cell := NewDigitalCell(nucleus, membrane)
	
	// Create the cell actor
	actor, err := CreateCellActor(cell, cs.system)
	if err != nil {
		return nil, fmt.Errorf("failed to create cell actor: %w", err)
	}

	// Add to the cells map
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	pidStr := actor.pid.String()
	cs.cells[pidStr] = actor
	
	return actor, nil
}

// CreateCellWithNeighbors creates a cell and connects it to existing neighbors.
func (cs *CellSystem) CreateCellWithNeighbors(
	nucleus interfaces.IIndividual, 
	membrane interfaces.IMembrane,
	neighborPIDs []*actor.PID,
) (*CellActor, error) {
	// Create the cell
	cellActor, err := cs.CreateCell(nucleus, membrane)
	if err != nil {
		return nil, err
	}

	// Add neighbors
	for _, pid := range neighborPIDs {
		// Find the neighbor actor
		cs.mu.RLock()
		neighborActor, exists := cs.cells[pid.String()]
		cs.mu.RUnlock()
		
		if exists {
			cellActor.AddNeighbor(neighborActor)
			// Also add ourselves as a neighbor to the neighbor
			neighborActor.AddNeighbor(cellActor)
		}
	}

	return cellActor, nil
}

// GetCell returns a cell actor by PID.
func (cs *CellSystem) GetCell(pid *actor.PID) (*CellActor, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	actor, exists := cs.cells[pid.String()]
	return actor, exists
}

// GetAllCells returns all cell actors.
func (cs *CellSystem) GetAllCells() []*CellActor {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	cells := make([]*CellActor, len(cs.cells))
	for _, cell := range cs.cells {
		cells = append(cells, cell)
	}
	return cells
}

// BroadcastToAll sends a message to all cells in the system.
func (cs *CellSystem) BroadcastToAll(message any) error {
	cells := cs.GetAllCells()
	for _, cell := range cells {
		cell.Broadcast(message)
	}
	return nil
}

// SendTickToAll sends a tick message to all cells.
func (cs *CellSystem) SendTickToAll(generation int) error {
	msg := NewMsgTick(generation, len(cs.cells))
	return cs.BroadcastToAll(msg)
}

// Shutdown shuts down the cell system.
func (cs *CellSystem) Shutdown() {
	// Stop all cells
	cells := cs.GetAllCells()
	for _, cell := range cells {
		cell.Stop()
	}
	
	// Shutdown the actor system
	cs.system.Shutdown()
}

// ConnectNeighbors connects all cells as neighbors in a grid pattern.
// This creates a 2D grid topology where each cell is connected to its immediate neighbors.
func (cs *CellSystem) ConnectNeighbors(gridSize int) error {
	cells := cs.GetAllCells()
	if len(cells) == 0 {
		return nil
	}

	// For simplicity, connect each cell to a few random neighbors
	// In a real implementation, this would create a proper topology
	for i, cell := range cells {
		// Connect to the next few cells (circular)
		for j := 1; j <= 3; j++ {
			neighborIdx := (i + j) % len(cells)
			if neighborIdx != i {
				cell.AddNeighbor(cells[neighborIdx])
			}
		}
	}

	return nil
}

// ConnectAll connects all cells to each other (fully connected graph).
func (cs *CellSystem) ConnectAll() error {
	cells := cs.GetAllCells()
	for i, cell := range cells {
		for j, other := range cells {
			if i != j {
				cell.AddNeighbor(other)
			}
		}
	}
	return nil
}

// CreateCellGrid creates a grid of cells with the given dimensions.
func (cs *CellSystem) CreateCellGrid(
	rows, cols int,
	nucleusFactory func(row, col int) interfaces.IIndividual,
	membraneFactory func() interfaces.IMembrane,
) ([][]*CellActor, error) {
	grid := make([][]*CellActor, rows)
	
	for i := 0; i < rows; i++ {
		grid[i] = make([]*CellActor, cols)
		for j := 0; j < cols; j++ {
			// Create nucleus
			nucleus := nucleusFactory(i, j)
			
			// Create membrane
			membrane := membraneFactory()
			
			// Create cell
			cell, err := cs.CreateCell(nucleus, membrane)
			if err != nil {
				return nil, fmt.Errorf("failed to create cell at (%d,%d): %w", i, j, err)
			}
			
			grid[i][j] = cell
		}
	}

	// Connect neighbors in the grid
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			cell := grid[i][j]
			
			// Connect to right neighbor
			if j < cols-1 {
				cell.AddNeighbor(grid[i][j+1])
			}
			
			// Connect to bottom neighbor
			if i < rows-1 {
				cell.AddNeighbor(grid[i+1][j])
			}
			
			// Connect to left neighbor
			if j > 0 {
				cell.AddNeighbor(grid[i][j-1])
			}
			
			// Connect to top neighbor
			if i > 0 {
				cell.AddNeighbor(grid[i-1][j])
			}
		}
	}

	return grid, nil
}
