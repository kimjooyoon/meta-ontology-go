# Domain observation example

This example models a small incident-resolution domain in Gooo.

The domain is intentionally concrete:

1. An observed service symptom is declared.
2. A diagnosis is proposed from that observation.
3. A resolution is applied without changing the original observation.
4. The outcome records whether the problem was resolved.

The source is a domain definition, not an implementation claim. The compiler
checks its semantic structure and can generate the corresponding Go artifact.
The activity chain is intentionally declared without runtime bindings because
the current generic generator rejects those bindings fail-closed; executing
the chain is a separate compiler capability to add with its own evidence.
The workflow generates it twice and compares the bytes, so a successful run
proves deterministic generation only. Runtime usefulness, performance, and
operational correctness remain separate observations.

This boundary is deliberate: facts about the domain are preserved as inputs,
while conclusions about the incident stay explicit and inspectable.
