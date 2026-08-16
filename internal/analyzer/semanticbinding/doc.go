// Package semanticbinding extracts the explicit Go-side semantic bindings.
//
// The package recognizes only the bind and obligation directives documented in
// this package. A directive is valid only when parser.ParseComments attaches
// its line-comment group directly to exactly one FuncDecl, TypeSpec, or GenDecl.
// The record span is the half-open AST span of that declaration, not the
// comment span. A GenDecl comment is valid only when the declaration contains
// exactly one named object. This makes grouped declarations deterministic and
// rejects an annotation whose target cannot be identified without guessing.
//
// Semantic IDs come only from directive fields. Package paths are explicit
// input; package, file, and declaration names are never promoted to IDs.
package semanticbinding
