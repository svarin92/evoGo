// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package cell

import (
	"log"
	"math/rand"

	"evoGo/interfaces"
)

// Membrane implements the IMembrane interface.
// It acts as the cellular interface for communication, filtering, and immune response.
type Membrane struct {
	// permeability is the current permeability level (0.0 to 1.0).
	// 0.0 = completely impermeable, 1.0 = completely permeable.
	permeability float64

	// fitnessThreshold is the minimum fitness required for a cell to be considered healthy.
	fitnessThreshold float64

	// immuneSys is the immune system associated with this membrane.
	immuneSys interfaces.IImmune

	// neighborPIDs is a list of neighboring cell PIDs for quorum sensing.
	neighborPIDs []any

	// blockList contains PIDs of cells that have been blocked due to malicious behavior.
	blockList map[any]bool

	// allowList contains PIDs of trusted cells (whitelist).
	allowList map[any]bool

	// mutationRateAdjustment is the current adjustment to mutation rate based on signals.
	mutationRateAdjustment float64
}

// NewMembrane creates a new membrane with default settings.
func NewMembrane() *Membrane {
	return &Membrane{
		permeability:          0.5, // Default: moderately permeable
		fitnessThreshold:      0.7, // Default: 70% fitness required
		immuneSys:             nil,
		neighborPIDs:          make([]any, 0),
		blockList:             make(map[any]bool),
		allowList:             make(map[any]bool),
		mutationRateAdjustment: 0.0,
	}
}

// NewMembraneWithParams creates a new membrane with custom parameters.
func NewMembraneWithParams(permeability, fitnessThreshold float64, immuneSys interfaces.IImmune) *Membrane {
	return &Membrane{
		permeability:     permeability,
		fitnessThreshold:  fitnessThreshold,
		immuneSys:        immuneSys,
		neighborPIDs:      make([]any, 0),
		blockList:        make(map[any]bool),
		allowList:        make(map[any]bool),
		mutationRateAdjustment: 0.0,
	}
}

// FilterIn filters an incoming message before it reaches the cell.
// Returns the filtered message or nil if the message is rejected.
func (m *Membrane) FilterIn(message any, cell interfaces.ICell) any {
	// Check if the message is from a blocked sender
	senderPID := m.extractSenderPID(message)
	if senderPID != nil && m.blockList[senderPID] {
		log.Printf("Membrane.FilterIn: Message from blocked sender %v rejected", senderPID)
		return nil
	}

	// Check if the message is from an allowed sender (whitelist)
	if senderPID != nil && m.allowList[senderPID] {
		return message // Always accept from trusted sources
	}

	// Handle different message types
	switch msg := message.(type) {
	case interfaces.MsgGeneticMaterial:
		return m.filterGeneticMaterial(msg, cell)
	case interfaces.MsgSignal:
		return m.filterSignal(msg, cell)
	case interfaces.MsgTick:
		return message // Always accept tick messages
	case interfaces.MsgCrossoverRequest:
		return m.filterCrossoverRequest(msg, cell)
	case interfaces.MsgCrossoverResponse:
		return message // Always accept crossover responses
	default:
		// Unknown message type: reject if permeability is low
		if m.permeability < 0.3 {
			return nil
		}
		return message
	}
}

// FilterOut filters an outgoing message before it is sent.
// Returns the filtered message or nil if the message should not be sent.
func (m *Membrane) FilterOut(message any, cell interfaces.ICell) any {
	// Check if the cell's fitness is below the threshold
	cellFitness := cell.GetFitness()
	if cellFitness < m.fitnessThreshold {
		// Cell is not healthy enough to send messages
		log.Printf("Membrane.FilterOut: Cell fitness %.2f below threshold %.2f, message blocked", cellFitness, m.fitnessThreshold)
		return nil
	}

	// Handle different message types
	switch msg := message.(type) {
	case interfaces.MsgGeneticMaterial:
		// Only allow sending genetic material if permeability is sufficient
		if m.permeability < 0.4 {
			return nil
		}
		return message
	case interfaces.MsgSignal:
		// Always allow signals (they're important for quorum sensing)
		return message
	case interfaces.MsgTick:
		return message
	case interfaces.MsgCrossoverRequest:
		// Only allow crossover requests if permeability is sufficient
		if m.permeability < 0.4 {
			return nil
		}
		return message
	case interfaces.MsgCrossoverResponse:
		return message
	default:
		return message
	}
}

// Excrete sends genetic material to another cell (horizontal gene transfer).
func (m *Membrane) Excrete(cell interfaces.ICell, codonFragment []int, semanticTag string) error {
	// Check if the cell has sufficient fitness
	if cell.GetFitness() < m.fitnessThreshold {
		return nil // Don't excrete if not healthy
	}

	// Check if permeability allows excretion
	if m.permeability < 0.4 {
		return nil // Don't excrete if membrane is not permeable
	}

	// Create and send the genetic material message
	msg := NewMsgGeneticMaterial(cell.GetPID(), codonFragment, cell.GetFitness(), semanticTag)
	
	// Get neighbors
	neighbors := cell.DiscoverNeighbors()
	if len(neighbors) == 0 {
		return nil
	}

	// Send to a random neighbor
	neighborPID := neighbors[rand.Intn(len(neighbors))]
	
	// TODO: Send via actor system
	log.Printf("Membrane.Excrete: Cell %v sending genetic material to %v", cell.GetPID(), neighborPID)
	
	return nil
}

// Absorb receives and processes incoming genetic material.
// Returns true if the material was successfully integrated.
func (m *Membrane) Absorb(cell interfaces.ICell, geneticMaterial interfaces.MsgGeneticMaterial) bool {
	// Check if the sender is blocked
	if m.blockList[geneticMaterial.SenderPID] {
		log.Printf("Membrane.Absorb: Genetic material from blocked sender %v rejected", geneticMaterial.SenderPID)
		return false
	}

	// Check if the sender's fitness is below our threshold
	if geneticMaterial.FitnessOfSender < m.fitnessThreshold {
		log.Printf("Membrane.Absorb: Genetic material from low-fitness sender (%.2f) rejected", geneticMaterial.FitnessOfSender)
		return false
	}

	// Check if permeability allows absorption
	if m.permeability < 0.3 {
		log.Printf("Membrane.Absorb: Membrane permeability too low (%.2f) for absorption", m.permeability)
		return false
	}

	// Pass to immune system for validation
	if m.immuneSys != nil {
		// Create a temporary individual with the genetic material
		// TODO: This needs proper implementation with the actual individual type
		log.Printf("Membrane.Absorb: Genetic material accepted, passing to immune system")
		// For now, we accept the material
		return true
	}

	// If no immune system, accept by default
	return true
}

// Signal sends a quorum sensing signal to neighboring cells.
func (m *Membrane) Signal(cell interfaces.ICell, signalType interfaces.SignalType, intensity interfaces.SignalIntensity) error {
	// Create the signal message
	msg := NewMsgSignal(cell.GetPID(), signalType, intensity, cell.GetFitness())
	
	// Send to all neighbors
	neighbors := cell.DiscoverNeighbors()
	for _, neighborPID := range neighbors {
		// TODO: Send via actor system
		log.Printf("Membrane.Signal: Cell %v sending %v signal (intensity: %.2f) to %v", 
			cell.GetPID(), signalType, intensity, neighborPID)
	}
	
	return nil
}

// AdjustSensitivity adjusts the membrane's sensitivity based on environmental signals.
func (m *Membrane) AdjustSensitivity(intensity interfaces.SignalIntensity) {
	// Adjust permeability based on signal intensity
	// Higher intensity signals (from healthy cells) increase permeability
	// Lower intensity signals (from stressed cells) decrease permeability
	if intensity > 0.7 {
		// Strong positive signal: increase permeability
		m.permeability = minFloat(m.permeability+0.1, 1.0)
		m.fitnessThreshold = minFloat(m.fitnessThreshold+0.05, 1.0)
	} else if intensity < 0.3 {
		// Strong negative signal: decrease permeability
		m.permeability = maxFloat(m.permeability-0.1, 0.0)
		m.fitnessThreshold = maxFloat(m.fitnessThreshold-0.05, 0.0)
	}

	// Adjust mutation rate based on signals
	// Stress signals increase mutation rate (exploration)
	// Stability signals decrease mutation rate (exploitation)
	if intensity < 0.5 {
		// Stress signal: increase exploration
		m.mutationRateAdjustment = minFloat(m.mutationRateAdjustment+0.05, 0.5)
	} else {
		// Stability signal: decrease exploration
		m.mutationRateAdjustment = maxFloat(m.mutationRateAdjustment-0.05, -0.5)
	}

	log.Printf("Membrane.AdjustSensitivity: Permeability=%.2f, FitnessThreshold=%.2f, MutationAdjustment=%.2f",
		m.permeability, m.fitnessThreshold, m.mutationRateAdjustment)
}

// GetPermeability returns the current permeability level.
func (m *Membrane) GetPermeability() float64 {
	return m.permeability
}

// SetPermeability sets the permeability level.
func (m *Membrane) SetPermeability(level float64) {
	m.permeability = clampFloat(level, 0.0, 1.0)
}

// GetImmuneSystem returns the immune system associated with this membrane.
func (m *Membrane) GetImmuneSystem() interfaces.IImmune {
	return m.immuneSys
}

// SetImmuneSystem sets the immune system for this membrane.
func (m *Membrane) SetImmuneSystem(immuneSys interfaces.IImmune) {
	m.immuneSys = immuneSys
}

// AddNeighbor adds a neighbor PID to the list.
func (m *Membrane) AddNeighbor(pid any) {
	// Check if not already in the list
	for _, existing := range m.neighborPIDs {
		if existing == pid {
			return
		}
	}
	m.neighborPIDs = append(m.neighborPIDs, pid)
}

// RemoveNeighbor removes a neighbor PID from the list.
func (m *Membrane) RemoveNeighbor(pid any) {
	for i, existing := range m.neighborPIDs {
		if existing == pid {
			m.neighborPIDs = append(m.neighborPIDs[:i], m.neighborPIDs[i+1:]...)
			return
		}
	}
}

// BlockSender adds a sender PID to the block list.
func (m *Membrane) BlockSender(pid any) {
	m.blockList[pid] = true
	log.Printf("Membrane.BlockSender: Blocked PID %v", pid)
}

// UnblockSender removes a sender PID from the block list.
func (m *Membrane) UnblockSender(pid any) {
	delete(m.blockList, pid)
	log.Printf("Membrane.UnblockSender: Unblocked PID %v", pid)
}

// AddToAllowList adds a sender PID to the allow list (whitelist).
func (m *Membrane) AddToAllowList(pid any) {
	m.allowList[pid] = true
	log.Printf("Membrane.AddToAllowList: Allowed PID %v", pid)
}

// RemoveFromAllowList removes a sender PID from the allow list.
func (m *Membrane) RemoveFromAllowList(pid any) {
	delete(m.allowList, pid)
	log.Printf("Membrane.RemoveFromAllowList: Removed PID %v from allow list", pid)
}

// GetMutationRateAdjustment returns the current mutation rate adjustment.
func (m *Membrane) GetMutationRateAdjustment() float64 {
	return m.mutationRateAdjustment
}

// GetNeighbors returns the list of neighbor PIDs.
func (m *Membrane) GetNeighbors() []any {
	return m.neighborPIDs
}

// extractSenderPID extracts the sender PID from a message.
func (m *Membrane) extractSenderPID(message any) any {
	switch msg := message.(type) {
	case interfaces.MsgGeneticMaterial:
		return msg.SenderPID
	case interfaces.MsgSignal:
		return msg.SenderPID
	case interfaces.MsgCrossoverRequest:
		return msg.SenderPID
	case interfaces.MsgCrossoverResponse:
		return msg.SenderPID
	default:
		return nil
	}
}

// filterGeneticMaterial filters incoming genetic material.
func (m *Membrane) filterGeneticMaterial(msg interfaces.MsgGeneticMaterial, cell interfaces.ICell) any {
	// Check if the sender's fitness is below our threshold
	if msg.FitnessOfSender < m.fitnessThreshold {
		log.Printf("Membrane.filterGeneticMaterial: Rejected material from low-fitness sender (%.2f)", msg.FitnessOfSender)
		// Block the sender if their fitness is very low
		if msg.FitnessOfSender < 0.3 {
			m.BlockSender(msg.SenderPID)
		}
		return nil
	}

	// Check if permeability allows absorption
	if m.permeability < 0.3 {
		return nil
	}

	// Accept the message
	return msg
}

// filterSignal filters incoming quorum sensing signals.
func (m *Membrane) filterSignal(msg interfaces.MsgSignal, cell interfaces.ICell) any {
	// Always accept signals - they're important for quorum sensing
	// But we can adjust our sensitivity based on the signal
	m.AdjustSensitivity(msg.Intensity)
	return msg
}

// filterCrossoverRequest filters incoming crossover requests.
func (m *Membrane) filterCrossoverRequest(msg interfaces.MsgCrossoverRequest, cell interfaces.ICell) any {
	// Check if the requester is blocked
	if m.blockList[msg.SenderPID] {
		return nil
	}

	// Check if our cell has sufficient fitness to participate in crossover
	if cell.GetFitness() < m.fitnessThreshold {
		return nil
	}

	// Check if permeability allows crossover
	if m.permeability < 0.4 {
		return nil
	}

	// Accept the request
	return msg
}

// Helper functions for float clamping
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
