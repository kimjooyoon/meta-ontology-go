# Source change impact

`gooo impact` compares two source files by stable declaration ID and explains
which declarations need another look. The command is part of the development
source. Use a CLI built from the branch containing this command.
From that compiler checkout, `go run ./cmd/gooo impact ...` also uses the checked-out command.

For example, a fee calculation may use `Reference` to obtain an answer while
assembling a body. `Main` then calls that fee calculation. Changing the answer
rule should make both of those relationships visible, even when the activity
names and Go signatures stay the same.

```sh
gooo impact examples/source-change-impact/before.gooo.fixture \
  examples/source-change-impact/after.gooo.fixture

gooo impact examples/source-change-impact/before.gooo.fixture \
  examples/source-change-impact/after.gooo.fixture --format json
```

The example changes `Reference` from `31 - input` to `32 - input`. The report
identifies one changed declaration and three declarations to recheck:
`Reference`, `Fee` through `observe:oracle`, and `Main` through its call to `Fee`.
`Unrelated` uses the same `Integer` type and stays outside that affected set.

## What Gooo carries through the comparison

- Stable IDs connect the two source views. A display rename remains a change to
  the same declaration.
- Source bodies, assembly declarations, field contracts, input order and typed
  dependency relationships are compared separately.
- Direct source calls, observation oracles, input/output types, field types,
  runtime bindings and feedback edges explain potential effects.
- Selection cases, condition cases, held evaluation cases and observation inputs
  keep separate counts before and after the change. These are declared counts;
  they contain no new pass/fail observations.
- Declared attempt limits, explicit fill candidate counts and grammar candidate
  limits retain their own fields. Fill plans report their candidate counts;
  `max_attempts` appears only when the source declares a positive attempt limit.
- Policy and import changes appear in `scope_changes`. Their unresolved consumer
  relationships remain in `unknowns`.

The old and new dependency graphs are followed separately. Each affected
declaration gets one nearest changed origin per revision, with a distance and
the incoming edge. Following those edges explains the chain. This bounds report
size even when a graph contains cycles or many routes to the same declaration.

## Reading the result

The JSON schema is `gooo/source-change-impact/v1`. It binds the two source hashes
and the report hash. `RECHECK_REQUIRED` means a source change may affect a
declaration. Run its construction, evaluation or saved-result replay to learn
the actual result.

`COMPLETE_STATIC` means this single-file inspection recorded no unresolved
dependency surface. `PARTIAL_STATIC` lists unresolved surfaces explicitly.
Neither status measures functional correctness or certifies a saved artifact.
Exit code 0 means the report was produced; automation should also read its status
and `unknowns`. Invalid input, expired context or output failure returns an error.

Calls through local values, external targets and malformed bodies remain
unresolved. Activities declared without a source body retain an explicit
uninspected implementation. Search-generated candidate call sets require materialization before
they can be resolved. Workspace imports and policy consumers require their
respective evaluators. Ordinary body text edits are reported conservatively,
including edits whose behavior might remain equivalent.

Explicit assembly alternatives and fill expressions are also inspected. When
an enclosing body declares the same name as a fragment call, its exact candidate
scope remains unresolved until materialization locates that expression. The
report does not turn that ambiguous name into a global activity relationship.

The command performs zero model calls and zero candidate executions. It does not
change source, regenerate code, invalidate artifacts or run tests. The CLI uses
the existing source-read limit and command deadline; the comparison additionally
bounds declarations to 4,096 and dependency edges to 32,768.

## Why this belongs in the language

A source program records more than a finished function body: it can describe
which choices are permitted, which examples guide assembly, and which reference
answers are consulted. Change impact lets tools follow those declarations. It
acts like a map of the questions that need to be asked again after a rule changes.
Execution supplies the answers, with its own evidence and scope.
