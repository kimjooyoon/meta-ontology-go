# Native plan/input regression observations

The 20 `plan-inputs-*.json.gz` files retain the exact CLI stdout from four actual
journeys with clean compiler source `a15b413412ad7b465a0567fb1db4de5706a8c2a8`,
Go 1.27.2 and the unchanged scalar own-model fixture. Capture occurred on
2026-10-10 KST before the 0.6.23 version bump. Gzip has no timestamp or source name.
The source/producer identities inside the original observations remain unchanged.

Each journey includes plan, template, input-only generation with two requests,
input-only saved replay and later scored saved replay. Input-only results remain
0/0 UNKNOWN. Later named expectations are 4, 4, 49 and 2 in their separate programs.
The Korean record uses one generation prediction; all runtime and saved replay
reports have zero new predictions. The other three generation routes use fixed order.

These are unit regression fixtures. Release readiness separately builds and
executes the candidate on all four native platforms at its own exact clean head.
