package main

var libraryStarterFiles = map[string]string{
	"core.gooo": `package core
namespace boundedint_core
entity Integer id "boundedint://integer"

activity Normalize(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"
`,
	"app.gooo": `package app
namespace boundedint_app
import core "boundedint/core"

activity Main(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__"

bind core.Normalize.result -> Main.input
`,
	"body-fill-plans.json": `{
  "schema": "gooo/workspace-body-fill-plans/v1",
  "activities": [
    {
      "package_path": "boundedint/core",
      "activity": "Normalize",
      "plan": {
        "schema": "gooo/body-codegen-ir-fill-plan/v1",
        "intent": "Add one to the input. 입력에 1을 더한다.",
        "hole_id": "value",
        "candidates": [
          {"id": "increment", "expression": "input + 1"},
          {"id": "identity", "expression": "input"}
        ],
        "test_cases": [
          {"input": -4, "expected": -3},
          {"input": 0, "expected": 1},
          {"input": 7, "expected": 8}
        ]
      }
    },
    {
      "package_path": "boundedint/app",
      "activity": "Main",
      "plan": {
        "schema": "gooo/body-codegen-ir-fill-plan/v1",
        "intent": "Return the value produced by the imported normalization activity.",
        "hole_id": "value",
        "candidates": [
          {"id": "identity", "expression": "input"},
          {"id": "zero", "expression": "0"}
        ],
        "test_cases": [
          {"input": -4, "expected": -4},
          {"input": 0, "expected": 0},
          {"input": 7, "expected": 7}
        ]
      }
    }
  ]
}
`,
	"gooo.workspace.json": `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "boundedint/app", "activity": "Main"},
  "packages": [
    {"path": "boundedint/app", "name": "app", "imports": ["boundedint/core"], "sources": ["app.gooo"]},
    {"path": "boundedint/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}
`,
	"cases.json": `{
  "schema": "gooo/body-composition-cases/v1",
  "cases": [
    {
      "inputs": {"boundedint/core:Normalize": 7},
      "expected": {
        "boundedint/core:Normalize": 8,
        "boundedint/app:Main": 8
      }
    }
  ]
}
`,
	"README.md": `# boundedint

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
`,
}
