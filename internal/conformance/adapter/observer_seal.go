package adapter

import (
	"crypto/sha256"
	"encoding/json"
)

type observerSealInput struct {
	Binding ObservationBinding `json:"binding"`
	Paths   ObserverPaths      `json:"paths"`
	Reason  RejectionKind      `json:"rejection_reason,omitempty"`
	Before  FilesystemState    `json:"before"`
	After   FilesystemState    `json:"after"`
}

func observationSeal(observation NoWriteObservation) [sha256.Size]byte {
	input := observerSealInput{
		Binding: observation.Binding,
		Paths:   observation.Paths,
		Reason:  observation.Reason,
		Before:  observation.Before,
		After:   observation.After,
	}
	payload, err := json.Marshal(input)
	if err != nil {
		panic("observer seal input is not JSON-marshalable")
	}
	return sha256.Sum256(payload)
}
