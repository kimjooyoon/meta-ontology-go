// Package bidir contains the parser-neutral core of the .gooo bidirectional
// transformation.
//
// Document, Model, and FactSet are stable boundaries for DSL, syntax, analyzer,
// and generator adapters. The syntax and semantic IR adapters stay at the
// boundary; the lens and reconciliation core operates only on neutral types.
// Business intent is represented by Document, stable semantic identity is
// represented by ID, and a Model is the normalized semantic IR.
//
// There are three fact layers:
//
//   - Syntactic facts describe observations made by a parser or analyzer. They
//     never change the semantic model.
//   - Deterministic facts are source-backed, unambiguous semantic observations
//     and may be reconciled into the model.
//   - Candidate facts are plausible but ambiguous observations. They are
//     retained as evidence, but never promoted by this package.
//
// Reconcile is transactional. A caller can therefore run it in a CI gate and
// reject the complete update when one fact violates identity, locality, or
// provenance policy.
//
// HostingContract records which host phase is actually evidenced. The
// gooo-hosted phase is represented as a planned target until its checks are
// verified; naming a future phase never makes it complete.
//
// ReconciliationFixture is the adapter contract for parser-neutral BX
// experiments. MeasureBXFixture emits stable golden evidence, while the
// package benchmarks exercise the same contract without importing a parser or
// generator implementation.
package bidir
