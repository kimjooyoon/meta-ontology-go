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
	"body-context": bodyContextHelp,
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
