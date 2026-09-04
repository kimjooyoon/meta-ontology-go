# Public generate self-observation

This small `.gooo` project is an opt-in example for the public compiler
self-observation ledger.

```text
first generate + empty ledger  -> UNKNOWN: one comparable observation
second generate + same ledger  -> CLOSED: stable non-executing candidate
third generate + same ledger   -> CLOSED: deterministic candidate replay
```

The normal `gooo generate` path is unchanged unless the caller supplies
`--observation-ledger DIRECTORY`. The directory is caller-owned and receives a
digest-only append-only ledger, an invocation receipt, and machine/human
reports. The generated source and manifest remain in the ordinary `--out`
directory.

The candidate is evidence only: it contains no executable patch, has
`execution_allowed: false`, and requires the repository's existing explicit
proposal, authorization, and certificate controls before any future use. No
repository source is modified by this observation loop.

The continuity handoff is explicit and caller-owned:

```text
candidate -> gooo authorize-discovery --decision accept|reject
          -> gooo certify-discovery (accepted only)
          -> gooo generate --continuity-certificate CERTIFICATE
```

The decision receipt and certificate carry the raw candidate digest and every
source, semantic, toolchain, contract, evaluator, output, and manifest digest.
Certification replays generation and records a typed conversion because a
public-discovery candidate and a semantic-retention certificate represent
different operations. Rejection is a terminal human decision; missing,
tampered, stale-source, stale-policy, and mismatched-toolchain evidence remain
distinct fail-closed outcomes. The continuity verifier requires four equal
candidate-digest edges and zero manual transformations.

The incompatible-source case is derived into caller-owned temporary output from
`project.gooo`; no second tracked language source is needed.

Corpus accounting is explicit: the fixed language corpus moves from 51 cases,
56 sources, and 997 `.gooo` lines to 53 cases, 58 sources, and 1,011 lines.
