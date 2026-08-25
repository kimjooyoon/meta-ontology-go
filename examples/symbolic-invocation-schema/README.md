# Symbolic invocation schema

This Gooo-only project declares one symbolic activity. The compiler projects
the activity name and ordered entity IDs as a JSON Schema Draft 2020-12
document:

```sh
gooo emit --kind symbolic-invocation-schema --entry Checkout examples/symbolic-invocation-schema
```

An independent validator in GitHub Actions checks one accepted instance and
one rejected counterexample. Generated files stay in runner temporary storage.

The schema validates symbolic identity and input order only. It does not claim
value-level types, domain correctness, production readiness, or performance
beyond the recorded runner samples.
