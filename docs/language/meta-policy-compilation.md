# Meaning-preserving compilation of a Gooo meta-policy

`examples/meta-policy-compilation/policy.gooo` owns the eight obligations, the
eight reduction rows, every decision coordinate, and the eight claim predicate
coordinates. The compiler lowers this raw source through the repository
semantic IR and checks only its fixed safety envelope. It does not maintain a
second Go list of policy meaning.

The producer uses the public `gooo` CLI path (`check --semantic` and
`generate`) and also compiles the raw file for its receipt. The independent
consumer reads and parses the same raw file separately, derives its own local
policy view, and has a CI-checked import boundary of `0` producer imports.

The canonical matrix is three source executions, three generated executions,
and three independent reconstructions. The valid SHA-256 contradiction is a
known refutation and is ordered before UNKNOWN conditions. Empty or malformed
evidence is lower-resolution UNKNOWN and retains `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`. An unrecognized upper
decision is fail-closed with reason
`FEEDBACK_COVERAGE_DECISION_UNKNOWN`; it cannot be accepted as `FIXED_POINT`.

The receipt contains eight distinct predicate observations per case and a
two-transition append-only lifecycle for each predicate. Its synthetic case
evidence is kept apart from a `CURRENT_EVIDENCE` runner observation. The
write-set claim compares exact sorted file snapshots (path, mode, size, and
content digest) at the start and end of the producer run; it claims only a
net repository change of zero, not that no system call wrote a file.

