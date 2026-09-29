# Experimental activity body code generation

`gooo body-codegen` is the first source-to-body projection experiment. It reads
an activity's existing `computes` string and emits one deterministic Go
function, bound to the activity's stable semantic ID by generated-region
markers:

```sh
go run ./cmd/gooo body-codegen --activity ClampBelowZero examples/body-codegen/main.gooo
```

The v1 body profile accepts one `Integer` or `Boolean` input and one matching
result, local `let` declarations, assignment to an existing local, `if/else`,
and one-value `return`. Conditions and expressions are checked by Go's type
checker after a closed syntax filter. Function calls, imports, loops, multiple
inputs, and external effects fail closed. The generated result is written to
stdout; this command does not mutate the repository.

The JSON report records the source/program/generated digests, number of
lowered statement constructs, typecheck result, deterministic replay digest,
and completeness over the accepted statement set. A `PASS` means every
statement in this closed profile was lowered and typechecked; it is not a
claim that the profile covers every possible Gooo program. The experiment
does not change `gooo generate`'s package projection or claim that the current
runtime executes these bodies. Laya remains outside source generation: a
planner may later select among bounded lowering routes, while this emitter
retains deterministic authority over the resulting Go.

The example intentionally uses a finite, side-effect-free body so later
experiments can compare behavior across codegen routes without granting the
model authority to produce code.
