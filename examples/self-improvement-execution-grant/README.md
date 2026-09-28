# Self-improvement execution grant v26

This is the bounded grant layer between the v24 system-derived candidate
eligibility receipt and the v25 pre-execution contract. It derives a one-use
grant only when the exact v24 request and resolution, v25 contract, source
artifact, candidate scope, and safety limits all agree.

The grant is not execution. A successful receipt is `CLOSED/GRANTED_UNCONSUMED`
with one remaining use, zero consumed uses, zero executions, zero repository
writes, and `one_use_enforced=false`. The next executor must independently
verify the receipt and consume it exactly once. This v26 layer neither consumes
the grant nor runs the candidate, produces output, compares results, or adopts
changes.

The policy has nine canonical cases: three successful system derivations, three
`UNKNOWN` cases for missing or incomplete upstream evidence, and three
`REFUTED` contradictions. Live output records system-derived grants separately
from executions and keeps performance `UNKNOWN` until comparable runtime
measurements exist.
