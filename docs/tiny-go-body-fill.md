# Local tiny_go body fill

`gooo body-codegen --tiny-model` opts a finite IR body-fill plan into the local
tiny_go classifier. The existing `--fill-plan` route keeps its current behavior
when `--tiny-model` is absent.

```sh
go run ./cmd/gooo body-codegen --json \
  --fill-plan plan.json --tiny-model model.json \
  --activity ChooseRegion activity.gooo
```

For example, a body can put the hole in a Boolean condition:

```gooo
activity ChooseRegion(Integer) -> Integer computes "if __GOOO_BODY_HOLE_region__ { return input + 1 } else { return input }"
```

Its plan may offer two distinct comparison operations:

```json
{
  "schema": "gooo/body-codegen-ir-fill-plan/v1",
  "intent": "Apply the offset only to inputs below the lower bound.",
  "hole_id": "region",
  "candidates": [
    {"id": "belowzero", "expression": "input < 0"},
    {"id": "loweredge", "expression": "input <= -10"}
  ],
  "test_cases": [
    {"input": -12, "expected": -11},
    {"input": -4, "expected": -4}
  ]
}
```

The activity must have one Integer input and one Integer output, with exactly
one typed expression hole. Every candidate is parsed, inserted into the body,
Go typechecked, and scored on the plan's declared finite test cases before the
model is called. The tiny_go classifier receives only the plan's `intent`
string. It does not receive body state, candidate expressions, test inputs,
expected values, candidate scores, or holdout data as classifier input. The
bodycodegen typed request still binds the plan and candidate evidence in its
request digest for audit.

Each candidate must have a supported binary expression at its root. Parentheses
around that root are ignored. The supported roots are `+`, `-`, `*`, `<`, `<=`,
`==`, `&&`, and `||`; every candidate in one plan must map to a different
operation. Candidate IDs can use any valid declared identifier and do not
affect the operation mapping. Literal, identifier-only, unary, unsupported, or
duplicate-root candidates fail before local inference. Those plans can still
use the existing default or Laya fill path.

The model proposes one operation. The compiler translates that operation back
to the corresponding declared candidate ID. An abstention or operation absent
from the plan selects the deterministic first-candidate fallback. The compiler
may then adjust a lower-scoring proposal to the best observed candidate based
on the same declared finite test cases. Reported functional accuracy is only
the selected body’s score on those cases; it is not a full-domain proof.

The report records `local_model_predictions`, `external_provider_calls`, and
whether the external-call count is known. A successful tiny_go path records one
local prediction and zero external provider choice invocations. The external
provider counter counts choice-inference calls; it excludes provider health or
readiness HTTP requests. `provider_decision_ms` measures every synchronous
resolver call, including validation and result mapping. `laya_decision_ms` is
populated only when the receipt records a successful Laya decision; failed or
unconfigured Laya routes are reported through `provider_decision_ms` and leave
the Laya field at zero. `tiny_decision_ms` measures the synchronous local
Resolve call, including request validation, inference, and choice mapping; these
fields are not model-kernel timings. The CLI also records `tiny_model_load_ms`.
`total_ms` starts after the source and plan have been read and the model bundle
has loaded. Direct API callers that provide an already-loaded model leave
model-load time absent.

The tiny_go receipt binds both `tiny_go_metadata_sha256`, the exact metadata
bytes read by the SDK when it loaded the model, and `tiny_go_weights_sha256`,
the validated packed weight bytes. The metadata digest changes when settings
such as confidence threshold or temperature change, even if the weights remain
the same.

The CLI rejects `--tiny-model` unless it is paired with `--fill-plan`. It also
rejects simultaneous `GOOO_LAYA_URL`, `GOOO_LAYA_API_KEY`, or a plan
`provider_model` selector. The model bundle is read locally; body-codegen emits
the generated projection and does not edit the `.gooo` source or repository.
