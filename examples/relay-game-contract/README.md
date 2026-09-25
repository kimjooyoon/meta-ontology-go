# Relay request graph contract

This is the unchanged four-field request declaration used by the Relay Lab
consumer in the public gooo-site repository. It is a record-forward contract,
not an implementation of the game's movement rules.

Public `gooo check --json main.gooo` and `gooo graph dump main.gooo` should both
accept the explicit field-V1 declaration. Graph output must retain four field
identities, parent links, populated type references, required presence and one
cardinality. Generic inspector seams remain deferred rather than gaining
implicit support. Graph inspection grants no write or execution authority.

The CLI regression includes three invalid variants: unknown type, duplicate
field identity and malformed cardinality. Validation runs in GitHub Actions;
the existence of this fixture is not evidence that those tests have passed.

Consumer evidence: https://github.com/kimjooyoon/meta-ontology-go/issues/763

The declaration is registered as a meta source in the syntax inventory, not as
an extra language completeness case. The existing 60-case denominator stays
unchanged. Its graph behavior is covered separately by the CLI regressions.
