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

`project-alt.gooo` is a second small project used to demonstrate that an
incompatible source/input group remains UNKNOWN rather than reaching quorum.
