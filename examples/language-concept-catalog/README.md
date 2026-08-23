# Language concept catalog

This catalog records useful concepts discovered while building Gooo. It does not claim that an ingredient has never appeared in another language. The comparison level is `UNCOMMON_COMBINATION`: familiar ideas become unusual when they are joined into one deterministic, evidence-carrying development loop.

| Concept | Problem addressed | Positive effect | Stage |
| --- | --- | --- | --- |
| metric meta-program | a number does not explain an action | producer, consumer, and meta-operation travel with the metric | operating |
| executable actionability | advice may have no implementation | every blocking operation resolves to a canonical executor | operating |
| effect-bounded observation | analyzers can alter their subject | repository writes are a zero-valued guardrail | conformed |
| monotone semantic resolution | uncertainty can look like success | unknown meaning descends toward invariant-only reasoning | conformed |
| causal feedback chain | any green run can be mistaken for the cause | one exact predecessor receipt becomes explicit state | conformed |
| CI-selected refactoring | cleanup can be subjective or destructive | metrics choose bounded AST rewrites and demand replay | operating |
| concept-governed refactoring | a selected operation may lack semantic authority | every candidate operation is digest-bound to a language concept | operating |
| toolchain executable use cases | examples can exist without executing rejection paths | one canonical path and two tamper paths emit an exact replayed receipt | operating |
| language syntax round-trip | syntax examples may not prove preservation | the complete registered corpus replays AST, canonical bytes, semantics, and lens laws | operating |

## Executable meaning

The source of truth is `internal/meta/languageconcept.Catalog`. CI evaluates every entry against the repository filesystem. A concept is accepted only when all of these bindings exist:

- a concrete problem and positive effect
- a named meta-operation
- one or more existing code paths
- one or more metric IDs
- an executable failure or success use case
- a maturity stage
- no unverified novelty claim

Missing meta code, missing use cases, missing metrics, unknown maturity, or novelty overclaim produces `FAIL_CLOSED`.
The toolchain executable-use-case concept additionally requires its dynamic
`3/3` receipt; a catalog row alone receives no readiness credit.
The language syntax round-trip concept likewise requires its dynamic `12/12`
receipt; static parser and formatter bindings alone receive no readiness credit.

## Current interpretation

`OPERATING` means the mechanism already participates in the active CI cycle. `CONFORMED` means its pure contract and regression cases pass, but a later integration step is still required before the cycle consumes it directly.

The project root remains an exceptional subject. Root metrics are observed, while root README and root-topology requirements remain `NOT_APPLICABLE` rather than being silently counted as failures.
