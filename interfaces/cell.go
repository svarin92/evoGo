// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

// SignalType defines the type of cellular signal for quorum sensing.
type SignalType int

const (
	// SignalStress indicates a cell is under stress (low fitness).
	// Neighbors should increase mutation rate to explore solutions.
	SignalStress SignalType = iota

	// SignalStability indicates a cell is stable (high fitness).
	// Neighbors should reduce mutation rate to exploit current solutions.
	SignalStability

	// SignalNeutral indicates no particular state.
	SignalNeutral
)

// SignalIntensity represents the strength of a cellular signal.
type SignalIntensity float64

// MsgGeneticMaterial represents a horizontal gene transfer message.
// It contains a fragment of codons to be integrated into the recipient's genome.
type MsgGeneticMaterial struct {
	// SenderPID is the identifier of the sending cell (actor PID).
	SenderPID any
	// CodonFragment is the sequence of codons being transferred.
	CodonFragment []int
	// FitnessOfSender is the fitness of the sending cell (for filtering).
	FitnessOfSender float64
	// SemanticTag describes the type of genetic material (e.g., "vowel_transition", "consonant_block").
	SemanticTag string
}

// MsgSignal represents a quorum sensing signal between cells.
type MsgSignal struct {
	// SenderPID is the identifier of the sending cell (actor PID).
	SenderPID any
	// SignalType indicates the type of signal (stress, stability, neutral).
	SignalType SignalType
	// Intensity represents the strength of the signal (0.0 to 1.0).
	Intensity SignalIntensity
	// FitnessOfSender is the fitness of the sending cell.
	FitnessOfSender float64
}

// MsgTick represents a metabolism/clock tick message for the cell.
type MsgTick struct {
	// Generation is the current generation number.
	Generation int
	// PopulationSize is the total number of cells in the population.
	PopulationSize int
}

// MsgCrossoverRequest represents a request for genetic crossover.
type MsgCrossoverRequest struct {
	// SenderPID is the identifier of the requesting cell.
	SenderPID any
	// PartnerPID is the identifier of the proposed partner.
	PartnerPID any
}

// MsgCrossoverResponse represents a response to a crossover request.
type MsgCrossoverResponse struct {
	// SenderPID is the identifier of the responding cell.
	SenderPID any
	// RequesterPID is the identifier of the requesting cell.
	RequesterPID any
	// Accepted indicates whether the crossover is accepted.
	Accepted bool
	// GenomeFragment is the genetic material offered for crossover.
	GenomeFragment []int
}

// ICell defines the interface for a cellular agent with a membrane.
// It extends IOrganism to include communication capabilities via the Actor model.
type ICell interface {
	IOrganism

	// GetPID returns the unique identifier of the cell in the actor system.
	GetPID() any

	// GetMembrane returns the cell's membrane.
	GetMembrane() IMembrane

	// Send sends a message to another cell.
	Send(to ICell, message any) error

	// Broadcast sends a message to all neighboring cells.
	Broadcast(message any) error

	// DiscoverNeighbors returns a list of neighboring cell PIDs.
	// This is used for quorum sensing and local interactions.
	DiscoverNeighbors() []any

	// GetFitnessThreshold returns the current fitness threshold for membrane permeability.
	GetFitnessThreshold() float64

	// SetFitnessThreshold sets the fitness threshold for membrane permeability.
	SetFitnessThreshold(threshold float64)

	// IsPermeable returns whether the membrane is currently permeable to genetic material.
	IsPermeable() bool
}

// IMembrane defines the interface for a cell membrane.
// The membrane manages all communication and filtering for the cell.
type IMembrane interface {
	// FilterIn filters an incoming message before it reaches the cell.
	// Returns the filtered message or nil if the message is rejected.
	FilterIn(message any, cell ICell) any

	// FilterOut filters an outgoing message before it is sent.
	// Returns the filtered message or nil if the message should not be sent.
	FilterOut(message any, cell ICell) any

	// Excrete sends genetic material to another cell (horizontal gene transfer).
	Excrete(cell ICell, codonFragment []int, semanticTag string) error

	// Absorb receives and processes incoming genetic material.
	// Returns true if the material was successfully integrated.
	Absorb(cell ICell, geneticMaterial MsgGeneticMaterial) bool

	// Signal sends a quorum sensing signal to neighboring cells.
	Signal(cell ICell, signalType SignalType, intensity SignalIntensity) error

	// AdjustSensitivity adjusts the membrane's sensitivity based on environmental signals.
	AdjustSensitivity(intensity SignalIntensity)

	// GetPermeability returns the current permeability level (0.0 to 1.0).
	// 0.0 = completely impermeable, 1.0 = completely permeable.
	GetPermeability() float64

	// SetPermeability sets the permeability level.
	SetPermeability(level float64)

	// GetImmuneSystem returns the immune system associated with this membrane.
	GetImmuneSystem() IImmune

	// SetImmuneSystem sets the immune system for this membrane.
	SetImmuneSystem(immuneSys IImmune)
}

// CellFactory is a factory function for creating instances of ICell.
type CellFactory func(organism IOrganism, membrane IMembrane) ICell
