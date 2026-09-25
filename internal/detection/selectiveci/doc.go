// Package selectiveci plans deterministic, digest-bound selective CI work.
//
// It adapts the existing semantic binding, impact graph, work frontier,
// resource envelope, and provenance path-closure contracts. It never runs a
// command and has no filesystem, environment, clock, randomness, or network
// dependency.
package selectiveci
