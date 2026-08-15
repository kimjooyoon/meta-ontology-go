// Package resourceenvelope evaluates supplied resource observations.
//
// The evaluator is deliberately independent of benchmark execution and wall
// clock discovery. Callers provide one warmup observation followed by the
// configured measured observations; the package returns a sealed integer-only
// result.
package resourceenvelope
