# Symbolic invocation schema

This Gooo-only project declares one symbolic activity. The compiler projects
the activity name and ordered entity IDs as a JSON Schema Draft 2020-12
document:

```sh
gooo emit --kind symbolic-invocation-schema --entry Checkout examples/symbolic-invocation-schema
```

The schema also contains one invocation example generated from the activity
name and ordered entity IDs. GitHub Actions compares that generated example to
an independent golden, then asks an independent validator to accept it and
reject one counterexample. Generated files stay in runner temporary storage.

The schema validates symbolic identity and input order only. It does not claim
value-level types, domain correctness, production readiness, or performance
beyond the recorded runner samples.
