# Architecture

```text
source.gooo ──parse──> AST ──lower──> Semantic IR ──project──> Go
    ▲                     │                 │                  │
    │                     ├──> PROV facts   ├──> query/search   │
    │                     ├──> constraints  ├──> docs/evidence │
    │                     └──> cache key    └──> CI scope      │
    └────────── Go analysis / semantic delta ───────────────────┘
```

## Package boundaries

- `internal/syntax`: tokens, parser, AST, spans, diagnostics;
- `internal/semantic`: identities, PROV vocabulary, graph facts, IR and normalization;
- `internal/bidir`: lowering, lifting, delta reconciliation, round-trip equivalence;
- `internal/generator`: deterministic Go projection and semantic source maps;
- `internal/analyzer`: Go AST/type facts limited to registered semantic symbols;
- `internal/lsp`: JSON-RPC/LSP transport and language features;
- `internal/cache`: content-addressed, versioned, atomic incremental artifacts;
- `cmd/gooo`: the user-facing CLI.

The kernel is standard-library-only. Interfaces are preferred at projection boundaries so
the DSL can evolve without coupling storage, search, or editor transport to the parser.

## Verification loop

Every change should be able to answer:

1. Which authoritative view changed?
2. What semantic delta was produced?
3. Is the delta within the allowed scope?
4. Are generated outputs and caches fresh for the same input hash?
5. What evidence proves parse, type, round-trip, and build invariants?

The GitHub Actions workflow runs formatting, vet, tests, race checks, static analysis,
CLI smoke tests, generated-output checks, and a cache/round-trip conformance job.
