# External ecosystem reference conformance

This example binds an external metaprogramming reference without counting its
features as Gooo implementations.

The pinned source is
[cosmos72/gomacro at cf0d4bf](https://github.com/cosmos72/gomacro/commit/cf0d4bf32da393dbda97e3572f216731013ffa55).
Its official README describes an interpreter, REPL, embedded evaluation,
AST-oriented macros, package imports, generics, and debugging.

The reference denominator has eight entries. The implemented capability count
remains zero. The unrestricted effects available to upstream macros are
recorded as a guardrail contrast, not as a feature target.

CI fetches only the two pinned documents. It never executes gomacro. Missing
documents lower resolution to UNKNOWN; known digest or contract mismatches use
INVARIANT_ONLY. Both routes fail closed.

See `usecases.json` for the exact expected decisions.
