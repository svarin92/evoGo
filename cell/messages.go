// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package cell

import "evoGo/interfaces"

// SignalType constants for quorum sensing.
const (
	SignalStress     = interfaces.SignalStress
	SignalStability  = interfaces.SignalStability
	SignalNeutral    = interfaces.SignalNeutral
)

// SignalIntensity type alias.
type SignalIntensity = interfaces.SignalIntensity

// MsgGeneticMaterial represents a horizontal gene transfer message.
type MsgGeneticMaterial = interfaces.MsgGeneticMaterial

// MsgSignal represents a quorum sensing signal between cells.
type MsgSignal = interfaces.MsgSignal

// MsgTick represents a metabolism/clock tick message for the cell.
type MsgTick = interfaces.MsgTick

// MsgCrossoverRequest represents a request for genetic crossover.
type MsgCrossoverRequest = interfaces.MsgCrossoverRequest

// MsgCrossoverResponse represents a response to a crossover request.
type MsgCrossoverResponse = interfaces.MsgCrossoverResponse

// NewMsgGeneticMaterial creates a new genetic material message.
func NewMsgGeneticMaterial(senderPID any, codonFragment []int, fitnessOfSender float64, semanticTag string) MsgGeneticMaterial {
	return MsgGeneticMaterial{
		SenderPID:      senderPID,
		CodonFragment:  codonFragment,
		FitnessOfSender: fitnessOfSender,
		SemanticTag:    semanticTag,
	}
}

// NewMsgSignal creates a new quorum sensing signal message.
func NewMsgSignal(senderPID any, signalType interfaces.SignalType, intensity SignalIntensity, fitnessOfSender float64) MsgSignal {
	return MsgSignal{
		SenderPID:      senderPID,
		SignalType:     signalType,
		Intensity:      intensity,
		FitnessOfSender: fitnessOfSender,
	}
}

// NewMsgTick creates a new tick message.
func NewMsgTick(generation int, populationSize int) MsgTick {
	return MsgTick{
		Generation:     generation,
		PopulationSize: populationSize,
	}
}

// NewMsgCrossoverRequest creates a new crossover request message.
func NewMsgCrossoverRequest(senderPID any, partnerPID any) MsgCrossoverRequest {
	return MsgCrossoverRequest{
		SenderPID: senderPID,
		PartnerPID: partnerPID,
	}
}

// NewMsgCrossoverResponse creates a new crossover response message.
func NewMsgCrossoverResponse(senderPID any, requesterPID any, accepted bool, genomeFragment []int) MsgCrossoverResponse {
	return MsgCrossoverResponse{
		SenderPID:     senderPID,
		RequesterPID:  requesterPID,
		Accepted:      accepted,
		GenomeFragment: genomeFragment,
	}
}
