package cell

import (
	"testing"

	"evoGo/model"
)

// TestNewDigitalCell tests the creation of a new DigitalCell.
func TestNewDigitalCell(t *testing.T) {
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	membrane := NewMembrane()

	cell := NewDigitalCell(individual, membrane)

	if cell == nil {
		t.Fatal("Expected non-nil cell")
	}

	if cell.GetNucleus() == nil {
		t.Fatal("Expected non-nil nucleus")
	}

	if cell.GetMembrane() == nil {
		t.Fatal("Expected non-nil membrane")
	}
}

// TestDigitalCellGetters tests the getter methods of DigitalCell.
func TestDigitalCellGetters(t *testing.T) {
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	individual.SetFitness(0.8)
	individual.SetPhenotype("test")

	membrane := NewMembrane()
	cell := NewDigitalCell(individual, membrane)

	// Test GetFitness
	if cell.GetFitness() != 0.8 {
		t.Errorf("Expected fitness 0.8, got %f", cell.GetFitness())
	}

	// Test GetPhenotype
	if cell.GetPhenotype() != "test" {
		t.Errorf("Expected phenotype 'test', got %v", cell.GetPhenotype())
	}
}

// TestDigitalCellSetters tests the setter methods of DigitalCell.
func TestDigitalCellSetters(t *testing.T) {
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	membrane := NewMembrane()
	cell := NewDigitalCell(individual, membrane)

	// Test SetFitness
	cell.SetFitness(0.9)
	if cell.GetFitness() != 0.9 {
		t.Errorf("Expected fitness 0.9, got %f", cell.GetFitness())
	}

	// Test SetPhenotype
	cell.SetPhenotype("new_phenotype")
	if cell.GetPhenotype() != "new_phenotype" {
		t.Errorf("Expected phenotype 'new_phenotype', got %v", cell.GetPhenotype())
	}
}

// TestMembraneFilterIn tests the membrane's incoming message filtering.
func TestMembraneFilterIn(t *testing.T) {
	membrane := NewMembraneWithParams(0.5, 0.7, nil)

	// Test with genetic material message
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	cell := NewDigitalCell(individual, membrane)

	msg := MsgGeneticMaterial{
		SenderPID:      "sender1",
		CodonFragment:  []int{5, 6, 7},
		FitnessOfSender: 0.8,
		SemanticTag:    "test",
	}

	// Filter should accept this message
	filtered := membrane.FilterIn(msg, cell)
	if filtered == nil {
		t.Error("Expected message to pass through filter")
	}

	// Test with low fitness sender
	msgLowFitness := MsgGeneticMaterial{
		SenderPID:      "sender2",
		CodonFragment:  []int{5, 6, 7},
		FitnessOfSender: 0.2,
		SemanticTag:    "test",
	}

	filtered = membrane.FilterIn(msgLowFitness, cell)
	if filtered != nil {
		t.Error("Expected message from low-fitness sender to be filtered out")
	}
}

// TestMembraneFilterOut tests the membrane's outgoing message filtering.
func TestMembraneFilterOut(t *testing.T) {
	membrane := NewMembraneWithParams(0.5, 0.7, nil)

	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	individual.SetFitness(0.8)
	cell := NewDigitalCell(individual, membrane)

	// Test with genetic material message
	msg := MsgGeneticMaterial{
		SenderPID:      "sender1",
		CodonFragment:  []int{5, 6, 7},
		FitnessOfSender: 0.8,
		SemanticTag:    "test",
	}

	// Filter should accept this message
	filtered := membrane.FilterOut(msg, cell)
	if filtered == nil {
		t.Error("Expected message to pass through filter")
	}

	// Test with low fitness cell
	individualLow := model.NewIndividual([]int{0, 1, 2})
	individualLow.SetFitness(0.2)
	cellLow := NewDigitalCell(individualLow, membrane)

	filtered = membrane.FilterOut(msg, cellLow)
	if filtered != nil {
		t.Error("Expected message from low-fitness cell to be filtered out")
	}
}

// TestMembranePermeability tests the membrane's permeability settings.
func TestMembranePermeability(t *testing.T) {
	membrane := NewMembrane()

	// Default permeability should be 0.5
	if membrane.GetPermeability() != 0.5 {
		t.Errorf("Expected default permeability 0.5, got %f", membrane.GetPermeability())
	}

	// Test setting permeability
	membrane.SetPermeability(0.8)
	if membrane.GetPermeability() != 0.8 {
		t.Errorf("Expected permeability 0.8, got %f", membrane.GetPermeability())
	}

	// Test clamping
	membrane.SetPermeability(1.5)
	if membrane.GetPermeability() != 1.0 {
		t.Errorf("Expected clamped permeability 1.0, got %f", membrane.GetPermeability())
	}

	membrane.SetPermeability(-0.1)
	if membrane.GetPermeability() != 0.0 {
		t.Errorf("Expected clamped permeability 0.0, got %f", membrane.GetPermeability())
	}
}

// TestMembraneNeighbors tests the membrane's neighbor management.
func TestMembraneNeighbors(t *testing.T) {
	membrane := NewMembrane()

	// Test adding neighbors
	membrane.AddNeighbor("pid1")
	membrane.AddNeighbor("pid2")
	membrane.AddNeighbor("pid3")

	neighbors := membrane.GetNeighbors()
	if len(neighbors) != 3 {
		t.Errorf("Expected 3 neighbors, got %d", len(neighbors))
	}

	// Test removing a neighbor
	membrane.RemoveNeighbor("pid2")
	neighbors = membrane.GetNeighbors()
	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors after removal, got %d", len(neighbors))
	}

	// Test adding duplicate neighbor
	membrane.AddNeighbor("pid1")
	neighbors = membrane.GetNeighbors()
	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors after duplicate add, got %d", len(neighbors))
	}
}

// TestMembraneBlockList tests the membrane's block list functionality.
func TestMembraneBlockList(t *testing.T) {
	membrane := NewMembrane()

	// Test blocking a sender
	membrane.BlockSender("bad_pid")

	// Test filtering a message from blocked sender
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	cell := NewDigitalCell(individual, membrane)

	msg := MsgGeneticMaterial{
		SenderPID:      "bad_pid",
		CodonFragment:  []int{5, 6, 7},
		FitnessOfSender: 0.8,
		SemanticTag:    "test",
	}

	filtered := membrane.FilterIn(msg, cell)
	if filtered != nil {
		t.Error("Expected message from blocked sender to be filtered out")
	}

	// Test unblocking
	membrane.UnblockSender("bad_pid")
	filtered = membrane.FilterIn(msg, cell)
	if filtered == nil {
		t.Error("Expected message from unblocked sender to pass through")
	}
}

// TestDigitalCellNeighbors tests the DigitalCell's neighbor management.
func TestDigitalCellNeighbors(t *testing.T) {
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	membrane := NewMembrane()
	cell := NewDigitalCell(individual, membrane)

	// Test adding neighbors
	cell.AddNeighbor("pid1")
	cell.AddNeighbor("pid2")

	neighbors := cell.DiscoverNeighbors()
	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors, got %d", len(neighbors))
	}

	// Test removing a neighbor
	cell.RemoveNeighbor("pid1")
	neighbors = cell.DiscoverNeighbors()
	if len(neighbors) != 1 {
		t.Errorf("Expected 1 neighbor after removal, got %d", len(neighbors))
	}
}

// TestDigitalCellPermeability tests the DigitalCell's permeability check.
func TestDigitalCellPermeability(t *testing.T) {
	genome := []int{0, 1, 2, 3, 4}
	individual := model.NewIndividual(genome)
	
	// Test with permeable membrane
	membrane := NewMembraneWithParams(0.6, 0.7, nil)
	cell := NewDigitalCell(individual, membrane)
	if !cell.IsPermeable() {
		t.Error("Expected cell to be permeable with permeability 0.6")
	}

	// Test with impermeable membrane
	membrane2 := NewMembraneWithParams(0.2, 0.7, nil)
	cell2 := NewDigitalCell(individual, membrane2)
	if cell2.IsPermeable() {
		t.Error("Expected cell to be impermeable with permeability 0.2")
	}
}

// TestMessages tests the message factory functions.
func TestMessages(t *testing.T) {
	// Test MsgGeneticMaterial
	msg := NewMsgGeneticMaterial("sender", []int{1, 2, 3}, 0.8, "test")
	if msg.SenderPID != "sender" {
		t.Errorf("Expected sender 'sender', got %v", msg.SenderPID)
	}
	if msg.FitnessOfSender != 0.8 {
		t.Errorf("Expected fitness 0.8, got %f", msg.FitnessOfSender)
	}

	// Test MsgSignal
	signal := NewMsgSignal("sender", SignalStability, 0.8, 0.9)
	if signal.SenderPID != "sender" {
		t.Errorf("Expected sender 'sender', got %v", signal.SenderPID)
	}
	if signal.SignalType != SignalStability {
		t.Errorf("Expected signal type SignalStability, got %v", signal.SignalType)
	}

	// Test MsgTick
	tick := NewMsgTick(5, 10)
	if tick.Generation != 5 {
		t.Errorf("Expected generation 5, got %d", tick.Generation)
	}
	if tick.PopulationSize != 10 {
		t.Errorf("Expected population size 10, got %d", tick.PopulationSize)
	}
}
