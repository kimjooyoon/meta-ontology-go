# Gooo project starters

`gooo init` creates a small project that can be checked and used immediately.
The default `app` starter is a single activity. Choose `library` when the
project's first deliverable is a reusable, source-level contract:

```sh
gooo init --template library boundedint
cd boundedint
gooo check main.gooo
gooo body-codegen --json --fill-plan body-fill-plan.json --activity Clamp main.gooo
```

The library template declares `Clamp(Integer) -> Integer`, names the
`boundedint` namespace, and marks two expression holes in the activity body.
`body-fill-plan.json` lists complete assignments for both holes and five
input/output examples. Gooo typechecks and scores every assignment before
emission. The examples provide a small regression set, not a proof over all
integers.

The Gooo source is the API contract and the body skeleton. When a local Laya
endpoint is configured, the model can choose among the complete assignments
already declared in the plan. It cannot invent another assignment. Without
Laya, generation starts from the declared-order fallback and adjusts to the
best-scoring assignment when the fallback scores lower. Gooo still checks the
selected body against the declared types and cases before showing the generated
Go code.

The starter deliberately has no dependency manifest or registry behavior yet.
It gives library authors a checkable contract and generation path; package
distribution and dependency resolution remain separate ecosystem work.
