// Package nativeobservation shares Gooo's existing native process wire types
// and original-build/check binding calculations with Go consumers. It uses only
// the Go standard library and does not import the compiler or invoke a model.
//
// These are recorded observations, not authentication or execution authority.
// Callers still validate their declared receipt/profile, source identity,
// successful original processes and current execution/reuse rules. Gooo-owned
// field roles are described by internal/meta/runtimeidentity/fields.gooo and the
// optional version --build --json --describe-runtime query.
package nativeobservation
