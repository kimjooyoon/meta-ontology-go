// Package semanticdelta detects semantic IR changes outside an allowed scope.
//
// The package owns a small, normalized boundary model instead of importing the
// semantic kernel. A future semantic.IR adapter can map nodes and deterministic
// facts into Snapshot through Adapter without coupling this detector to that
// package's concrete types.
//
// Integration contract:
//   - AST, Go analysis, BX, and codegen adapters produce Snapshot or Delta.
//   - CI supplies a deterministic Request and consumes a Report.
//   - Cache keys may use EncodeJSON(Request); provenance uses stable IDs only.
//   - Host comparison requires equal semantic reports; an unimplemented host
//     remains deferred and cannot be reported as a successful stage.
package semanticdelta
