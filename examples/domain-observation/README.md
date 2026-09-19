# Domain observation example

This example models a small incident-resolution domain in Gooo.

The domain is intentionally concrete:

1. An observed service symptom is declared.
2. A diagnosis is proposed from that observation.
3. A resolution is applied without changing the original observation.
4. The outcome records whether the problem was resolved.

The source is a domain definition, not an implementation claim. The compiler
checks its semantic structure. Its explicit `bind` edges are carried through
the canonical syntax and bidirectional model into typed semantic IR.

`CompileTypedPlan` uses that same model to order activities; it does not execute
them. The workflow runs the syntax and lowering test suites and checks this
fixture, retaining the exact source identity and test events even on failure.

The current Go generator rejects runtime bindings as unsupported. This example
does not claim generated execution, deterministic generation, runtime
usefulness, performance improvement, or operational correctness. A compiler
check passing is not evidence for any of those claims.

This boundary is deliberate: facts about the domain are preserved as inputs,
while conclusions about the incident stay explicit and inspectable.

See [blocker scopes](../../docs/development/blocker-scopes.md) for the distinction
between continuing compiler development, obtaining evidence, and merging.
