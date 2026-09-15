// Package freshness checks whether provenance records and reconstructable
// projections still describe the current authoritative inputs.
//
// The package deliberately owns a small, dependency-free manifest boundary.
// Adapters can map semantic IR, cache metadata, or generated-source manifests
// into Snapshot values without coupling this detector to those packages.
package freshness
