// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package cell

import (
	"fmt"
	"log"
	"math/rand"

	"evoGo/interfaces"
	"evoGo/model"
)

// DigitalCell represents a cellular agent with a nucleus (Individual) and a membrane.
// It implements the ICell interface and uses the Actor model for concurrent communication.
type DigitalCell struct {
	// nucleus is the core of the cell, containing the genome and phenotype.
	nucleus interfaces.IIndividual

	// membrane is the cellular interface for communication and filtering.
	membrane interfaces.IMembrane

	// pid is the unique identifier of the cell in the actor system.
	pid any

	// fitnessThreshold is the minimum fitness required for the cell to be active.
	fitnessThreshold float64

	// generation is the current generation of the cell.
	generation int

	// neighbors is a cache of neighboring cell PIDs.
	neighbors []any
}

// NewDigitalCell creates a new DigitalCell with the given nucleus and membrane.
func NewDigitalCell(nucleus interfaces.IIndividual, membrane interfaces.IMembrane) *DigitalCell {
	return &DigitalCell{
		nucleus:        nucleus,
		membrane:       membrane,
		pid:            nil, // Will be set when registered with actor system
		fitnessThreshold: 0.7,
		generation:     0,
		neighbors:       make([]any, 0),
	}
}

// NewDigitalCellWithPID creates a new DigitalCell with a specific PID.
func NewDigitalCellWithPID(nucleus interfaces.IIndividual, membrane interfaces.IMembrane, pid any) *DigitalCell {
	cell := NewDigitalCell(nucleus, membrane)
	cell.pid = pid
	return cell
}

// CreateFromIndividual creates a DigitalCell from an existing Individual.
func CreateFromIndividual(ind *model.Individual, membrane interfaces.IMembrane) *DigitalCell {
	return NewDigitalCell(ind, membrane)
}

// GetPID returns the unique identifier of the cell in the actor system.
func (c *DigitalCell) GetPID() any {
	return c.pid
}

// SetPID sets the unique identifier of the cell.
func (c *DigitalCell) SetPID(pid any) {
	c.pid = pid
}

// GetMembrane returns the cell's membrane.
func (c *DigitalCell) GetMembrane() interfaces.IMembrane {
	return c.membrane
}

// SetMembrane sets the cell's membrane.
func (c *DigitalCell) SetMembrane(membrane interfaces.IMembrane) {
	c.membrane = membrane
}

// GetNucleus returns the cell's nucleus (Individual).
func (c *DigitalCell) GetNucleus() interfaces.IIndividual {
	return c.nucleus
}

// SetNucleus sets the cell's nucleus.
func (c *DigitalCell) SetNucleus(nucleus interfaces.IIndividual) {
	c.nucleus = nucleus
}

// GetFitness returns the cell's fitness.
func (c *DigitalCell) GetFitness() float64 {
	return c.nucleus.GetFitness()
}

// GetPhenotype returns the cell's phenotype.
func (c *DigitalCell) GetPhenotype() any {
	return c.nucleus.GetPhenotype()
}

// GetTemplate returns the cell's template.
func (c *DigitalCell) GetTemplate() any {
	return c.nucleus.GetTemplate()
}

// GetUsedWraps returns the number of used wraps.
func (c *DigitalCell) GetUsedWraps() int {
	return c.nucleus.GetUsedWraps()
}

// SetFitness sets the cell's fitness.
func (c *DigitalCell) SetFitness(value float64) {
	c.nucleus.SetFitness(value)
}

// SetPhenotype sets the cell's phenotype.
func (c *DigitalCell) SetPhenotype(value any) {
	c.nucleus.SetPhenotype(value)
}

// SetTemplate sets the cell's template.
func (c *DigitalCell) SetTemplate(value any) {
	c.nucleus.SetTemplate(value)
}

// Send sends a message to another cell.
func (c *DigitalCell) Send(to interfaces.ICell, message any) error {
	// Filter the outgoing message through our membrane
	filteredMsg := c.membrane.FilterOut(message, c)
	if filteredMsg == nil {
		log.Printf("DigitalCell.Send: Message blocked by membrane filter")
		return fmt.Errorf("message blocked by membrane")
	}

	// TODO: Send via actor system
	log.Printf("DigitalCell.Send: Cell %v sending message to %v", c.pid, to.GetPID())
	
	// For now, directly deliver the message
	return to.Receive(filteredMsg)
}

// Broadcast sends a message to all neighboring cells.
func (c *DigitalCell) Broadcast(message any) error {
	neighbors := c.DiscoverNeighbors()
	for _, neighborPID := range neighbors {
		// TODO: Look up neighbor by PID and send message
		// For now, just log
		log.Printf("DigitalCell.Broadcast: Cell %v broadcasting to %v", c.pid, neighborPID)
	}
	return nil
}

// DiscoverNeighbors returns a list of neighboring cell PIDs.
func (c *DigitalCell) DiscoverNeighbors() []any {
	// Return cached neighbors for now
	// In a real implementation, this would query the actor system
	return c.neighbors
}

// AddNeighbor adds a neighbor PID to the cache.
func (c *DigitalCell) AddNeighbor(pid any) {
	// Check if not already in the list
	for _, existing := range c.neighbors {
		if existing == pid {
			return
		}
	}
	c.neighbors = append(c.neighbors, pid)
}

// RemoveNeighbor removes a neighbor PID from the cache.
func (c *DigitalCell) RemoveNeighbor(pid any) {
	for i, existing := range c.neighbors {
		if existing == pid {
			c.neighbors = append(c.neighbors[:i], c.neighbors[i+1:]...)
			return
		}
	}
}

// GetFitnessThreshold returns the current fitness threshold for membrane permeability.
func (c *DigitalCell) GetFitnessThreshold() float64 {
	return c.fitnessThreshold
}

// SetFitnessThreshold sets the fitness threshold for membrane permeability.
func (c *DigitalCell) SetFitnessThreshold(threshold float64) {
	c.fitnessThreshold = threshold
	// Also update the membrane's threshold
	if mem, ok := c.membrane.(*Membrane); ok {
		mem.fitnessThreshold = threshold
	}
}

// IsPermeable returns whether the membrane is currently permeable to genetic material.
func (c *DigitalCell) IsPermeable() bool {
	if mem, ok := c.membrane.(*Membrane); ok {
		return mem.GetPermeability() > 0.3
	}
	return false
}

// Receive handles incoming messages for the cell.
// This is the main message handler that implements the cellular behavior.
func (c *DigitalCell) Receive(message any) error {
	// Filter the incoming message through our membrane
	filteredMsg := c.membrane.FilterIn(message, c)
	if filteredMsg == nil {
		log.Printf("DigitalCell.Receive: Message blocked by membrane filter")
		return nil // Message was filtered out
	}

	// Process the message based on its type
	switch msg := filteredMsg.(type) {
	case interfaces.MsgTick:
		return c.handleTick(msg)
	case interfaces.MsgGeneticMaterial:
		return c.handleGeneticMaterial(msg)
	case interfaces.MsgSignal:
		return c.handleSignal(msg)
	case interfaces.MsgCrossoverRequest:
		return c.handleCrossoverRequest(msg)
	case interfaces.MsgCrossoverResponse:
		return c.handleCrossoverResponse(msg)
	default:
		log.Printf("DigitalCell.Receive: Unknown message type: %T", msg)
		return nil
	}
}

// handleTick handles metabolism and evolution for the cell.
func (c *DigitalCell) handleTick(msg interfaces.MsgTick) error {
	c.generation = msg.Generation

	// 1. Perform internal metabolism (evolution)
	if err := c.performEvolution(); err != nil {
		return fmt.Errorf("failed to perform evolution: %w", err)
	}

	// 2. Membrane behavior: communicate with neighbors if healthy
	if c.GetFitness() > c.fitnessThreshold {
		neighbors := c.DiscoverNeighbors()
		for _, n := range neighbors {
			// TODO: Send genetic material or signals to neighbors
			// For now, just excrete genetic material
			if c.IsPermeable() {
				// Extract a fragment of our genome
				genome := c.nucleus.GetGenome()
				if len(genome) > 10 {
					fragment := genome[:10] // First 10 codons
					c.GetMembrane().Excrete(c, fragment, "random_fragment")
				}
			}

			// Send quorum sensing signal
			if c.GetFitness() > 0.8 {
				c.GetMembrane().Signal(c, interfaces.SignalStability, interfaces.SignalIntensity(c.GetFitness()))
			} else if c.GetFitness() < 0.4 {
				c.GetMembrane().Signal(c, interfaces.SignalStress, interfaces.SignalIntensity(1.0-c.GetFitness()))
			}
		}
	}

	return nil
}

// handleGeneticMaterial handles incoming genetic material (horizontal gene transfer).
func (c *DigitalCell) handleGeneticMaterial(msg interfaces.MsgGeneticMaterial) error {
	log.Printf("DigitalCell.handleGeneticMaterial: Received genetic material from %v with fitness %.2f",
		msg.SenderPID, msg.FitnessOfSender)

	// Pass to membrane for absorption
	if c.membrane.Absorb(c, msg) {
		// Genetic material accepted by membrane
		// TODO: Integrate the genetic material into our genome
		// This would involve crossover or mutation with the incoming fragment
		
		// For now, just log
		log.Printf("DigitalCell.handleGeneticMaterial: Genetic material accepted, length: %d", len(msg.CodonFragment))
		
		// Optionally: perform crossover with the incoming material
		if rand.Float64() < 0.3 { // 30% chance of integrating the material
			c.integrateGeneticMaterial(msg.CodonFragment)
		}
	}

	return nil
}

// handleSignal handles quorum sensing signals.
func (c *DigitalCell) handleSignal(msg interfaces.MsgSignal) error {
	log.Printf("DigitalCell.handleSignal: Received %v signal with intensity %.2f from %v",
		msg.SignalType, msg.Intensity, msg.SenderPID)

	// Adjust membrane sensitivity based on the signal
	c.membrane.AdjustSensitivity(msg.Intensity)

	// Adjust our own behavior based on the signal
	switch msg.SignalType {
	case interfaces.SignalStress:
		// Neighbor is in stress: increase our mutation rate (exploration)
		log.Printf("DigitalCell.handleSignal: Neighbor in stress, increasing exploration")
		// TODO: Adjust our mutation rate
		
	case interfaces.SignalStability:
		// Neighbor is stable: we can be more conservative
		log.Printf("DigitalCell.handleSignal: Neighbor stable, maintaining stability")
		
	case interfaces.SignalNeutral:
		// No particular action
	}

	return nil
}

// handleCrossoverRequest handles a request for genetic crossover.
func (c *DigitalCell) handleCrossoverRequest(msg interfaces.MsgCrossoverRequest) error {
	log.Printf("DigitalCell.handleCrossoverRequest: Received crossover request from %v", msg.SenderPID)

	// Check if we accept the request
	// For now, accept with 50% probability
	if rand.Float64() < 0.5 {
		// Extract a fragment of our genome
		genome := c.nucleus.GetGenome()
		fragment := make([]int, 0)
		
		// Take a random segment
		if len(genome) > 0 {
			start := rand.Intn(len(genome))
			end := start + rand.Intn(len(genome)-start)
			if end > len(genome) {
				end = len(genome)
			}
			fragment = genome[start:end]
		}

		// Send response
		response := NewMsgCrossoverResponse(c.pid, msg.SenderPID, true, fragment)
		c.Send(c, response) // This will fail - need to look up the sender
		
		// For now, just log
		log.Printf("DigitalCell.handleCrossoverRequest: Accepted, sending fragment of length %d", len(fragment))
	} else {
		// Reject the request
		response := NewMsgCrossoverResponse(c.pid, msg.SenderPID, false, nil)
		log.Printf("DigitalCell.handleCrossoverRequest: Rejected")
		// TODO: Send response
	}

	return nil
}

// handleCrossoverResponse handles a response to a crossover request.
func (c *DigitalCell) handleCrossoverResponse(msg interfaces.MsgCrossoverResponse) error {
	if msg.Accepted {
		log.Printf("DigitalCell.handleCrossoverResponse: Crossover accepted by %v, received fragment of length %d",
			msg.SenderPID, len(msg.GenomeFragment))
		
		// Integrate the received fragment
		c.integrateGeneticMaterial(msg.GenomeFragment)
	} else {
		log.Printf("DigitalCell.handleCrossoverResponse: Crossover rejected by %v", msg.SenderPID)
	}

	return nil
}

// performEvolution performs the cell's internal evolution.
func (c *DigitalCell) performEvolution() error {
	// This would typically involve:
	// 1. Mutation
	// 2. Phenotype generation
	// 3. Fitness evaluation
	
	// For now, just perform a random mutation
	genome := c.nucleus.GetGenome()
	if len(genome) > 0 {
		// Mutate a random codon
		index := rand.Intn(len(genome))
		newValue := rand.Intn(256) // Assuming codons are in range 0-255
		
		// Create a new genome with the mutation
		newGenome := make([]int, len(genome))
		copy(newGenome, genome)
		newGenome[index] = newValue
		
		c.nucleus.SetGenome(newGenome)
		
		log.Printf("DigitalCell.performEvolution: Mutated codon at index %d to %d", index, newValue)
	}

	return nil
}

// integrateGeneticMaterial integrates incoming genetic material into the cell's genome.
func (c *DigitalCell) integrateGeneticMaterial(fragment []int) {
	if len(fragment) == 0 {
		return
	}

	genome := c.nucleus.GetGenome()
	if len(genome) == 0 {
		// If we have no genome, just use the fragment
		c.nucleus.SetGenome(fragment)
		return
	}

	// Perform one-point crossover with the fragment
	// Choose a random crossover point
	crossoverPoint := rand.Intn(len(genome))
	
	// Create new genome
	newGenome := make([]int, 0, len(genome)+len(fragment))
	
	// Add part of our genome before crossover point
	newGenome = append(newGenome, genome[:crossoverPoint]...)
	
	// Add the fragment
	newGenome = append(newGenome, fragment...)
	
	// Add the rest of our genome
	newGenome = append(newGenome, genome[crossoverPoint:]...)
	
	// Set the new genome
	c.nucleus.SetGenome(newGenome)
	
	log.Printf("DigitalCell.integrateGeneticMaterial: Integrated fragment of length %d at position %d",
		len(fragment), crossoverPoint)
}

// String returns a string representation of the cell.
func (c *DigitalCell) String() string {
	return fmt.Sprintf("DigitalCell{PID: %v, Fitness: %.2f, Generation: %d, Neighbors: %d}",
		c.pid, c.GetFitness(), c.generation, len(c.neighbors))
}

// DoAccept implements the IVisitedModel interface for the Visitor pattern.
func (c *DigitalCell) DoAccept(visitor interfaces.IVisitor) {
	// Delegate to the nucleus if it implements the Visitor pattern
	if notified, ok := c.nucleus.(interfaces.INotifiedModel); ok {
		notified.DoAccept(visitor)
	}
}

// SetVisited sets the visited status (part of INotifiedModel).
func (c *DigitalCell) SetVisited(visited bool) {
	if notified, ok := c.nucleus.(interfaces.INotifiedModel); ok {
		notified.SetVisited(visited)
	}
}

// GetVisited returns the visited status (part of INotifiedModel).
func (c *DigitalCell) GetVisited() bool {
	if notified, ok := c.nucleus.(interfaces.INotifiedModel); ok {
		return notified.GetVisited()
	}
	return false
}
