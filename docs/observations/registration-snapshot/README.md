# Preserving historical syntax scope while registration grows

The value-origin module registers one required source case. Its version 48
snapshot records a syntax target of 80. Later registration operations add new
source cases and increase the current counters in a temporary candidate.

PR #1536 exposed a conflict in a new regression: the historical snapshot test
also required current counters to remain 82 total and 80 language cases. Both
[syntax candidate execution](original-syntax-failure.log) and
[common registration execution](original-native-registration-failure.log) stopped
at that same assertion. Their original heads and runs remain unchanged.

The correction checks version 48 against its retained digest and target 80,
and compares its other boundaries with version 47. The existing complete corpus
conformance still checks the current source inventory and counters. Future
registration can grow while the historical scope keeps its original meaning.

## Fresh observation

[Producer 6863](producer6863.json) is revision
`b87072bef6017b8c04af7c8c4af17adce6361e35` plus three recorded replacements.
The two test files and a deliberately synthetic `model.go.fixture` retain their
observed bytes. That synthetic file changes only observation counters to 83 total
and 81 language cases; it is not part of the submitted compiler.

The [new static test](observation6863.log) read frozen target 80 and current
synthetic target 81 successfully. It executed no historical source program,
registration operation, model or old test cohort. This narrows the cause of the
original failure; the submitted revision still requires its own complete CI.

The final sibling test also rejects modified historical bytes, and avoids
requiring other live boundary counts to equal a historical version. Its final
source differs from the preserved observation producer and has not been locally
replayed. Those final assertions belong to the new revision's CI.

## Activity declaration annotations in saved receipts

The [original repository projection log](original-projection-failure.log) also
reported failures in package construction and saved input observations. Prepared
programs now include `activity_origins`. The release comparison understood absent
historical `condition_origins`, but still required the new declaration annotation
to exist in every old program record.

The comparison now applies the same presence rule to these two named optional
annotations. If a field exists in the recorded program, its full value must match.
Explicit null, an empty list that differs from the prepared declaration, a changed
origin or an unknown member remain mismatches. Source, entry identity, workspace
image and the other program members retain their comparisons.

[Producer 6971](producer6971.json) is `b87072be` plus the two preserved replacement
files. Its [fresh program and ten receipt forms](observation6971.log) passed both
the program comparison and the raw receipt entry point. The source contains one
activity declaration and one condition; the compiler only prepared this source.
There were no native runs, model calls or executions of historical source cohorts.
The final sibling source matches this producer. The submitted revision's own CI
must still check the full construction and replay integration.
