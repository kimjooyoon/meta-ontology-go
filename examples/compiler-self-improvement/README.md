# Compiler self-improvement

This fixture is an executable compiler-improvement loop. The `.gooo` program is
the authority for the semantic graph, the transformation rule, the measurement
stages, and the decision precedence.

The declared improvement folds IR collection canonicalization into the
generator's existing deep-copy pass. The generated candidate removes one
redundant collection-copy pass while preserving canonical empty collections.
The workflow materializes the candidate only under the runner's temporary
directory, compares it with the checked-out implementation, and runs the same
billing fixture through baseline and candidate source trees.

The primary outcome is the deterministic collection-copy count (`2 -> 1`).
Generated output must remain byte-identical and deterministic. Wall time,
peak RSS, build time, and test time are integer observations and guardrails;
they do not replace the primary deterministic count.
