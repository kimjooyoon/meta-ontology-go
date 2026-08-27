# Semantic delta receipt

This experiment makes three change layers explicit for a pair of Gooo source
files:

| Layer | What is measured | What it cannot prove alone |
| --- | --- | --- |
| `textual_delta` | raw bytes, byte digests, and changed byte count | that meaning changed |
| `structural_delta` | canonical nodes and directed semantic facts | that every claim interpretation is valid |
| `semantic_claim_delta` | stable claim IDs, changed claim values, and transitions | behavior outside the bounded Gooo grammar |

The fixed denominator is three cases: `equivalent`, `semantic-change`, and
`indeterminate`. The first changes text while preserving the graph and claims;
the second changes one small output signature and therefore changes a semantic
claim; the third contains a source construct outside this experiment's grammar
and must remain `FAIL_CLOSED / LOWER_RESOLUTION` at
`project-source/parse-lower/SEMANTIC_TRANSLATION_VALIDATION_UNAVAILABLE`.

The producer records the three layers. The consumer is an independent
adjudicator in a separate package: it rereads both raw sources and recomputes the expected layers
before accepting the receipt. It rejects a tampered receipt and performs zero
repository writes. The meta-operation is
`separate-text-structural-semantic-deltas`.

The raw decision and semantic decision are separate receipt fields. Claim
transitions persist `OPEN`, `DISCHARGED`, or `REFUTED`; an unknown subject is
reported as `FAIL_CLOSED / LOWER_RESOLUTION` with the exact stage `bind-subject`, step
`resolve-subject`, and reason `SEMANTIC_DELTA_SUBJECT_UNKNOWN`.

Object propositions are not preservation propositions. A changed or removed
before object claim refutes its separate preservation row; an after-only object
claim is a canonical source observation and is discharged. Proposition and
preservation digests stabilize claim IDs, and old rows remain in the receipt.

## Research decisions

The experiment adopts the following principles from primary sources:

1. Semantic diffing is a separate question from syntactic diffing. Microsoft
   Research describes semantic diffing as comparing program abstractions and
   explaining behavioral differences, rather than only changed lines. See
   [Abstract Semantic Diffing of Evolving Concurrent Programs](https://www.microsoft.com/en-us/research/publication/abstract-semantic-diffing-evolving-concurrent-programs/).
2. Validation is per translation/change, not a claim that every future
   transformation is correct. LLVM's translation-validation description says
   each compiler run is followed by a validation pass for the produced target.
   See [Program Analysis for Compiler Validation](https://llvm.org/pubs/2008-11-PASTE-CompilerValidation.html).
3. Semantic preservation must be stated over observable meaning. CompCert's
   [formal semantic-preservation theorem](https://compcert.org/man/manual001.html)
   is the boundary we echo: a successful translation may be credited only when
   the resulting meaning is related to the source meaning.

We reject raw line-count equality as a semantic proof, text equality as the
definition of equivalence, and approximation as an `EXACT` result. An
unparseable or unsupported source is evidence that the validator is unable to
decide, not evidence that the source is equivalent.

## Meta-operation contract

The operation binds `Producer`, `Consumer`, `Stage`, `Step`, `Reason`, and a
proof choice to every receipt indicator. `FOUNDATION` binds bytes and the
canonical graph, `COHERENCE` checks the graph/claim/adjudication relationship,
and `REGRESSION` checks the no-write invariant. The suite reports a fixed
`3/3 = 10000` basis points; its observed case counts are `textual=3`,
`structural=3`, `semantic_preserved=1`, `semantic_changed=1`, and
`indeterminate=1`.

For a known pair:

```text
textual_changed && structural_delta == empty && semantic_claim_delta == empty
  => SEMANTIC_PRESERVED / FIXED_POINT / EXACT
textual_changed && (structural_delta != empty || semantic_claim_delta != empty)
  => SEMANTIC_CHANGED / DELTA_OBSERVED / EXACT
```

Any failed source projection or receipt replay is `INDETERMINATE` and
`FAIL_CLOSED`; projection failure uses `LOWER_RESOLUTION` at
`project-source/parse-lower`. No activation or promotion is implied. The
suite's `FIXED_POINT` is only fixed three-case contract reproduction; subject
semantic equivalence is separately `NOT_ASSERTED`.

## Falsification

The claim is falsifiable. Reordering stable IDs, changing the output entity,
adding a supported relation, or changing a claim value must move the receipt
from `SEMANTIC_PRESERVED` to `SEMANTIC_CHANGED`. Adding a grammar feature that
the bounded projector cannot parse must remain `INDETERMINATE`. Mutating any
receipt layer without regenerating its digest must be rejected by the
independent adjudicator. The result does not claim whole-program behavioral
equivalence, compiler correctness, or coverage outside the fixed grammar.
