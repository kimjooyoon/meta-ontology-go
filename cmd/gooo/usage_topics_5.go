package main

var topicHelpGroup5 = map[string]string{
	"body-context": `Inspect source input and model compatibility

` + bodyContextUsage + `

From the compiler source checkout:
  gooo body-context --activity Describe \
    --model examples/scalar-identity/model/model.json examples/scalar-identity/source.gooo.fixture

The local record-field model's verified metadata selects its input feature.
READY_FOR_RANKING means the source input can be encoded for that model.
DECLINED_TO_DETERMINISTIC retains the reason, such as THREE_FIELD_CHOICES_REQUIRED.
Inspection makes zero predictions and candidate tests. It prepares and typechecks
source alternatives; assembly checks their finite cases. A recorded representation
decline is a successful inspection result. Invalid files or source return failure.

--value-flow adds the source provenance graph. --include-plan adds the separate
finite cases. --model currently applies to source-owned record field assembly;
an explicit --feature-version must match the loaded model.
See docs/record-model-preflight.md and gooo help body-compose.
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

The source owns alternatives and finite assembly cases. --cases supplies the
native execution inputs. Assembly field/case counts and native results are
reported separately. Omitting --model uses deterministic order. The runtime
model_calls count is zero during saved replay; the saved assembly report retains
its original call count. Native compilation requires the project's Go toolchain.
See docs/native-body-composition.md and gooo help body-context.
`,
}
