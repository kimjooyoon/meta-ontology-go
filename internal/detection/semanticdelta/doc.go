// Package semanticdelta detects semantic IR changes outside an allowed scope.
//
// The package owns a small, normalized boundary model instead of importing the
// semantic kernel. A future semantic.IR adapter can map nodes and deterministic
// facts into Snapshot through Adapter without coupling this detector to that
// package's concrete types.
//
// The adapter contract projects only stable node IDs, node kinds, and
// deterministic subject/predicate/object facts. Display names, source spans,
// candidate facts, and host-specific evidence stay outside the delta. A
// producer represents removal explicitly; absence in a partial snapshot is
// not interpreted as removal.
//
// Request and Report have deterministic JSON and text encodings. A report with
// any out-of-scope endpoint is a failed check; an unimplemented host remains
// deferred and cannot be represented as a successful self-hosting stage.
package semanticdelta
