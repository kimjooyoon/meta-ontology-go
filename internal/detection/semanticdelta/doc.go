// Package semanticdelta detects semantic IR changes outside an allowed scope.
//
// The package owns a small, normalized boundary model instead of importing the
// semantic kernel. A future semantic.IR adapter can map nodes and deterministic
// facts into Snapshot through Adapter without coupling this detector to that
// package's concrete types.
package semanticdelta
