# Gooo project starters

`gooo init` creates a small project that can be checked and used immediately.
The default `app` starter is a single activity. Choose `library` when the
project's first deliverable is a reusable, source-level contract:

```sh
gooo init --template library boundedint
cd boundedint
gooo check main.gooo
gooo body-codegen --json --activity Clamp main.gooo
```

The library template declares `Clamp(Integer) -> Integer`, names the
`boundedint` namespace, and keeps three input/output examples next to the
activity. Those cases let body generation reject candidates that do not match
the examples. They provide a small regression set, not a proof over all
integers.

The Gooo source is the API contract and the source for generation. When a local
Laya endpoint is configured, the model can rank the routes already declared in
that file. Without Laya, generation follows the same deterministic route. Gooo
still checks the candidate against the declared types and cases before showing
the generated Go code.

The starter deliberately has no dependency manifest or registry behavior yet.
It gives library authors a checkable contract and generation path; package
distribution and dependency resolution remain separate ecosystem work.
