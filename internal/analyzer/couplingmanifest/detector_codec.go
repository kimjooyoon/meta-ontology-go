package couplingmanifest

import detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"

type DetectorInput = detector.Input

// DecodeInput delegates strict JSON decoding to the detector package.
func DecodeInput(data []byte) (detector.Input, error) { return detector.DecodeInput(data) }

// DecodeJSON is retained as an adapter vocabulary alias for detector input.
func DecodeJSON(data []byte) (detector.Input, error) { return detector.DecodeInput(data) }

// Decode is the concise detector-input spelling.
func Decode(data []byte) (detector.Input, error) { return detector.DecodeInput(data) }

// EncodeInput delegates canonical JSON encoding to the detector package.
func EncodeInput(input detector.Input) ([]byte, error) { return detector.EncodeInput(input) }

// EncodeJSON is the adapter vocabulary alias for detector input encoding.
func EncodeJSON(input detector.Input) ([]byte, error) { return detector.EncodeInput(input) }

// DecodeResult delegates strict result decoding without reinterpreting the
// detector result algebra.
func DecodeResult(data []byte) (detector.Result, error) { return detector.DecodeResult(data) }

// EncodeResult delegates result encoding without rewriting detector bytes.
func EncodeResult(result detector.Result) ([]byte, error) { return detector.EncodeResult(result) }
