# boundedint

This starter is a small Gooo package graph. The core package declares
Normalize(Integer) -> Integer; the app package imports it and binds its result
to Main. Each activity has one typed body hole. The plan lists the candidate
expressions and finite examples that Gooo uses to score them.

Install the Gooo CLI and run these commands from this directory:

~~~sh
go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@dev
gooo package resolve gooo.workspace.json
gooo package execute --json --cases cases.json --body-plans body-fill-plans.json gooo.workspace.json
~~~

The body plans also accept a local compact arithmetic model:

~~~sh
gooo package execute --json --cases cases.json --body-plans body-fill-plans.json \
  --tiny-model path/to/model.json gooo.workspace.json
~~~

`--tiny-model` loads one local model for the sequential activity fills. It cannot
be combined with `GOOO_LAYA_URL` or `GOOO_LAYA_API_KEY`. Without it, Laya is
used when configured; otherwise the declared candidate order is selected
deterministically. The local model currently supports the starter's one-hole
integer plans. Unsupported plan shapes or operators produce a diagnostic.

The execute command follows the declared binding, fills both bodies, compiles
the generated Go, and runs it against the named case. Gooo scores each
candidate before emission. Set GOOO_LAYA_URL to a local Laya /v1/systemone
endpoint to let the model choose only among listed candidates. Without a model,
candidate selection is deterministic. Gooo typechecks the completed bodies and
reports finite observed accuracy; these examples do not prove behavior for all
integer inputs.

The workspace manifest records package imports and the public entry. Package
resolve prints the deterministic graph receipt; package execute adds generated
native execution and a replayable receipt for the chosen body fills.
