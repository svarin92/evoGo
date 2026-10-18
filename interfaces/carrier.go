// Copyright 2026 Stéphane Varin. All rights reserved.
// Use of this source code is governed by the MIT license.
// See the LICENSE file for details.
package interfaces

import "errors"

// GenomeCarrier is the contract through which a GE individual accepts a
// deposit of genetic material from an external module (e.g., cellular
// tissue from an evoCell). It acts as a symmetrical counterpart to the
// HybridizationHook: while the hook notifies the outside world of
// synchronization points, the carrier allows the outside world to deposit
// codons INTO an individual, without exposing the internal structure of
// the genome.
//
// SEMANTICS OF THE DEPOSIT — three guarantees of implementation:
//
// 1. IMPOTENCE WITH REGARD TO FITNESS. Deposit does not modify an
// individual's current fitness: evaluation remains the sole
// responsibility of the engine. An individual that has undergone a
// graft is never considered "already evaluated."
//
// 2. PHENOTYPE STATUS. Following a 'Deposit', the phenotype is
// obsolete until the next derivation (orchestrated by the
// Genomizer—never by the carrier).
//
// 3. OPTIONAL CARRIER. Callers obtain the contract via type 
// assertion: an entity that does not implement it is silently 
// ignored by the attachment mechanism.
//
// POLICY: the contract encodes NO selection policy (elitism included) 
// — the targeting of grafts is up to the caller.
type IGenomeCarrier interface {

    // Deposit replaces the slice [anchor, anchor+len(codons)) of the
    // genome with the provided codons. This is an ATOMIC operation: 
    // either the entire slice is replaced, or nothing changes—a failed 
    // deposit never leaves the genome in a partially modified state.
    //
    // nil: deposit completed, phenotype obsolete; 
    // ErrAnchorOutOfRange: anchor outside the genome, nothing has changed; 
    // ErrEmptyDeposit: empty codons, nothing has changed.
    Deposit(anchor int, codons []int) error
}

// Sentinels — comparable to errors.Is.
var (
    ErrAnchorOutOfRange = errors.New("interfaces: deposit anchor out of range")
    ErrEmptyDeposit = errors.New("interfaces: empty deposit")
)