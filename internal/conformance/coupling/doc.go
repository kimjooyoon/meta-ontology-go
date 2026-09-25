// Package coupling contains a research-only, producer-independent oracle for
// the code-to-semantic coupling contract.
//
// The package intentionally owns its input codec, normalization, predicates,
// and baseline. It does not import a production detector. Expected fixture
// labels are test data and are not part of oracle inputs or result digests.
package coupling
