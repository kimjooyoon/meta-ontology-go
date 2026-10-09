package main

var topicHelpGroup5 = map[string]string{
	"body-plan": `Read caller inputs before assembling a program

` + bodyPlanUsage + `

From the compiler source checkout:
  gooo body-plan --source examples/body-codegen/native-input-joins.gooo.fixture --entry Label
  gooo body-plan --source examples/body-codegen/native-input-joins.gooo.fixture \
    --entry Label --inputs-template > inputs.json
  gooo body-compose --source examples/body-codegen/native-input-joins.gooo.fixture \
    --entry Label --inputs inputs.json --out out/observed

Edit the template's zeros, false values and empty strings to supply your inputs.
Bound ports receive their producer's result and are omitted from the template.
Optional record fields are omitted; add them explicitly when present.
--json exports source and contract digests, stable IDs, the structural plan,
root input keys and called-body assembly order. Inspection performs zero model
calls, candidate tests or native executions. Body checks, model compatibility
and correctness are separate. No expected outputs are created.
See docs/composition-plan-inspection.md and gooo help body-compose.
`,
	"body-context": `Inspect source input and model compatibility

` + bodyContextUsage + `

From the compiler source checkout:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture

The verified local model's metadata selects its input feature.
READY_FOR_RANKING means the source input can be encoded for that model.
DECLINED_TO_DETERMINISTIC retains the reason, such as THREE_FIELD_CHOICES_REQUIRED
for a record with a different number of selected fields.
Inspection makes zero predictions and candidate tests. It prepares and typechecks
source alternatives; assembly checks their finite cases. A recorded representation
decline is a successful inspection result. Invalid files or source return failure.

Typed branches and variables:
  gooo body-context --activity Choose --include-plan \
    --model examples/scalar-identity/model/model.json examples/caller-typed-paths/unary.gooo.fixture

The current record model reports THREE_DECISION_COUNT_UNSUPPORTED for this one
branch choice. Deterministic assembly remains available. --include-plan shows
source choices, legal options, fallbacks and the separate finite cases. Joint
models also export their complete framed input when compatible.
--value-flow applies to records. An explicit --feature-version must match the
loaded model. Whole-candidate judges and source IR search have separate contracts.
See docs/typed-model-preflight.md, docs/record-model-preflight.md and gooo help body-compose.
`,
	"body-compose": `Assemble, compile and execute a typed Gooo program

` + bodyComposeUsage + `

From the compiler source checkout, using a new output directory:
  gooo body-compose --source examples/scalar-identity/source.gooo.fixture \
    --cases examples/scalar-identity/cases.json --entry Describe \
    --model examples/scalar-identity/model/model.json --out out/scalar-model

Saved replay:
  gooo body-compose --source out/scalar-model/original.gooo \
    --cases examples/scalar-identity/cases.json --composition out/scalar-model/composition.json

Observe values before writing runtime expectations:
  gooo body-compose --source examples/scalar-identity/source.gooo.fixture \
    --inputs examples/composition-inputs/record.json --entry Describe --out out/scalar-observed

--inputs uses gooo/body-composition-inputs/v1 and saves inputs.json. It reports
actual root and intermediate values with finite_total=0, finite_passed=0 and
NO_RUNTIME_EXPECTATIONS (UNKNOWN correctness). Supply --cases on saved replay
to score caller expectations later. --inputs, --cases and --case-series are exclusive.

The source owns alternatives and finite assembly cases. --cases supplies the
native execution inputs and expected outputs. Assembly field/case counts and native results are
reported separately. Omitting --model uses deterministic order. The runtime
model_calls count is zero during saved replay; the saved assembly report retains
its original call count. Native compilation requires the project's Go toolchain.
See docs/native-body-composition.md and gooo help body-context.
`,
}
