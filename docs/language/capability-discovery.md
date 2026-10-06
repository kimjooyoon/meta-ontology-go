# Capability discovery

`gooo discover` answers a narrow question: which declared capability in the
deterministic JEV catalog best matches a natural-language query? It binds the
query and observation trail to the supplied `.gooo` source. It does not infer
new language features from the source, call a language model, generate code, or
run a program.

```sh
gooo discover --query "How do I generate a canonical .gooo declaration?" \
  examples/language-operation-catalog/main.gooo
```

Add `--json` to receive a machine-readable trail and the shared completeness
receipt. The receipt records the exact source digest, normalized semantic IR
digest, query digest, and JEV evidence digest. Its `catalog_match` dimension
describes a catalog result only; it is not a measure of whether Gooo implements
the requested behavior.

## Reading the result

- `AVAILABLE` means the query matched a catalog entry. It does not mean that
  generation or runtime behavior was verified.
- `DEFERRED` means the catalog recognizes the capability but an external
  boundary remains unresolved.
- `UNKNOWN` means the query did not resolve to a known catalog entry. The
  compiler preserves this state instead of guessing.
- `PROGRESS` is the receipt decision while required evidence remains open.
  The receipt intentionally has no aggregate completeness score.

For every discovery, generation coverage, independent use-case coverage, and
reverse-observation coverage remain `UNKNOWN`. They need separate evidence:
generated artifacts bound to a declaration; independent inputs with expected
outputs; and runtime observations mapped back to the originating source. A
catalog suggestion or a repeated fixture is not a substitute for those
observations. See [declared completeness receipts](../declared-completeness-receipt.md)
and [body generation](body-codegen.md) for the next stages.
