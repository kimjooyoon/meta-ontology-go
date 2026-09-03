# Compiler self-improvement

This fixture is an executable compiler-improvement loop. The
`operation-envelope.gooo` program is the authority for intent, source binding,
the read-only effect grant, replay identity, the measurement rule, and decision
precedence. Its semantic IR is materialized into six caller-owned artifacts per
fixed scenario and independently verified before the candidate is accepted.

The declared improvement makes repeated LSP refresh requests reuse the exact
document result when the source, support profile, toolchain, and cache contract
digests all match. A changed source or identity digest invalidates reuse and
parses again; corrupt or stale evidence is rejected by the independent
operation-envelope verifier. The workflow materializes the candidate only under
the runner's temporary directory, compares generated output across baseline and
candidate source trees, and measures the same billing fixture through both
revisions.

The primary outcome is the deterministic unchanged-refresh parse-call count
(`1 -> 0`).
Generated output must remain byte-identical and deterministic. Wall time, peak
RSS, build time, test time, allocation count, source lines, and artifact topology
are exact integer observations; performance superiority is not inferred from
noisy timing.
