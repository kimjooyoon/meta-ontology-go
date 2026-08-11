// Package performance defines host-neutral performance contracts and evidence.
//
// The initial implementation may be Go-hosted while a future implementation is
// gooo-hosted. Both hosts must report the same contract before their measurements
// are comparable; planned or unavailable evidence is never treated as a pass.
// Experiment results pass only on exact fixture, repetition, operation, and
// allocation matches; counterexamples fail, and unimplemented hosts defer.
package performance
