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
// experiments. BXEvidenceFixture is the hard evidence extension: it must
// provide non-empty DSL, IR, Go, source-map, evidence, and provenance base
// artifacts, ordered delta hashes, locality closure JSON, evidence ID/span
// cardinality, accepted-write before/after bytes and lstat snapshots, and an
// observer-owned rejected-write adapter. MeasureBXFixture rejects incomplete
// contracts; missing evidence is never green. A rejected partial observation
// must preserve semantic/source/region/slot/bytes/lstat digests, prove atomic
// no-write through its observer, must not create removals or promote
// candidates, and must retain the filesystem/inode seam as deferred.
//
// Source-order preservation is currently defined for source-backed activity
// input ports. Deterministic facts must reference registered model endpoints.
// Partial observations never imply relation removal, and explicit removals
// remain transactional and idempotent.
//
// Generic gooo:invokes lifting, PROV-O mapping policy, Go-lift/CLI delta
// atomicity, Go-side port inference, and three-way merge remain explicit
// deferred seams until their owning APIs and evidence exist. CLI delta
// atomicity belongs to the CLI ownership boundary and is also deferred here.
package bidir
