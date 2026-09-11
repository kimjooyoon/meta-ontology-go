# Value-witness execution input

This directory owns the v27A execution-input schema and its language-level
registry entry. The artifact is a caller-owned input declaration: it snapshots
the existing value-witness source, binds the `Increment` activity and its
semantic identity, carries the typed input/expected-output corpus, and closes
the evaluator and safety identities.

It does not execute the activity, grant execution, write the repository, run
tests, or produce runtime output. Those responsibilities remain deferred to a
future bounded executor.
