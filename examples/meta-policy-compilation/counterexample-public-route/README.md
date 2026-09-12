# A counterexample through the public Gooo profiles

The accepted counterexample-execution adapter derives an exact revision request
and observes it through the source-owned Gooo operation. This example shows how
that same request reaches the public CLI without hand-selecting a new decision.

Use caller-owned paths. WORK must be outside PROJECT and contain no input source.
The example does not apply a candidate to the project or authorize adoption.

```sh
go run ./examples/meta-policy-compilation/counterexample-execution \
  -policy "$POLICY" -input "$INPUT" \
  -operation internal/meta/policycompilation/revision-operation.gooo \
  > "$WORK/execution.json"

jq -r '.revision_request_json' "$WORK/execution.json" > "$WORK/request.json"

gooo generate "$POLICY" --out "$WORK/proposal" \
  --profile meta-policy-revision-v1 \
  --profile-package metapolicycompilation --profile-namespace metapolicycompilation \
  --profile-project-root "$PROJECT" \
  --profile-source-digest "$(jq -r '.expected_source_digest' "$WORK/request.json")" \
  --profile-condition "$(jq -r '.condition' "$WORK/request.json")" \
  --profile-from-decision "$(jq -r '.from_decision' "$WORK/request.json")" \
  --profile-to-decision "$(jq -r '.to_decision' "$WORK/request.json")" --json

gooo check "$WORK/proposal/candidate.gooo"

gooo generate "$WORK/proposal/candidate.gooo" --out "$WORK/generated" \
  --profile meta-policy-compilation-v3 \
  --profile-package metapolicycompilation --profile-namespace metapolicycompilation \
  --profile-project-root "$WORK/proposal" --json
```

The revision profile emits candidate.gooo and proposal.json; the compilation
profile keeps its four-artifact contract. Both report nonexecution until a
separate observer actually runs the generated judge.

The native CI witness additionally replays both revision artifacts byte-for-byte,
executes the public-generated judge with the three unchanged candidate cases,
and compares every result with the bound execution receipt. It records actual
public commands, process exits and source/semantic/generated-byte identities.
No caller evidence digest or validator expectation is repaired to make it pass.

This closes a public-route integration boundary only when native CI supplies its
evidence. Synthetic cases, independent receipt consistency and fresh execution
are not external utility, persistent automatic adoption or speed improvement.
The existing seven #804 requirements are not a whole-language progress score.
