// Package bindingcoverage observes deterministic coverage of declared binding
// partitions. It is a pure metric foundation, not a catalog metric or CI gate.
// Work units count every endpoint reference, including repeated IDs shared by
// multiple bindings; a valid binding always contributes two references.
// Precedence ranks define the immutable order of registered stage/reason pairs.
package bindingcoverage
