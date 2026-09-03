# Compiler self-improvement

This fixture is an executable compiler-improvement loop. The
`operation-envelope.gooo` program is the authority for intent, source binding,
the read-only effect grant, replay identity, the measurement rule, and decision
precedence. Its semantic IR is materialized into six caller-owned artifacts per
fixed scenario and independently verified before the candidate is accepted.

The declared improvement removes the redundant duplicate-key map and key
materialization in import normalization. Duplicate import paths are already
rejected before that second check, so the candidate preserves validation order
and errors while reducing the deterministic materialization count from `1` to
`0`. The workflow materializes the candidate only under the runner's temporary
directory, compares it with the checked-out implementation, and runs the same
fixture through baseline and candidate source trees.

The primary outcome is the deterministic import duplicate-key materialization
count (`1 -> 0`).
Generated output must remain byte-identical and deterministic. Wall time, peak
RSS, build time, test time, allocation count, source lines, and artifact topology
are exact integer observations; performance superiority is not inferred from
noisy timing.
