# Minimal causal explanation

This experiment explains the verdict of the real Gooo value-witness compiler
receipt at `examples/language-value-witness/main.gooo`. The receipt is read as
raw observation input; the experiment does not reproduce its prose.

The authority boundary is explicit:

- `main.gooo` declares evidence entity IDs, PROV activity relations, the
  decision predicate value, the prior claim state, six meta-operation values,
  and the fixed indicator denominator.
- The producer and independent consumer both run
  `syntax.ParseFile` → `bidir.Lower` → canonical semantic IR. Neither accepts
  a Go-side canonical graph as the source of meaning.
- The path set is authoritative. `explanation_text` is `INCIDENTAL` and is
  ignored by the independent consumer.
- The observed evidence is the actual compiler receipt. The audit noise added
  to the overlong path and every removal experiment are marked `SYNTHETIC`.

## Formal choices

[Halpern and Pearl, *Causes and Explanations: A Structural-Model Approach,
Part II: Explanations*](https://arxiv.org/abs/cs/0208034) motivates actual-cause
and structural-model counterfactual reasoning. The implementation therefore
executes one removal intervention for every evidence member of each sufficient
path.

[Lynce and Marques-Silva, *On Computing Minimum Unsatisfiable
Cores*](https://www.satisfiability.org/SAT04/programme/110.pdf) distinguishes
minimal from minimum: deletion of every single member proves only
`SUBSET_MINIMAL`/irredundancy; a global cardinality claim requires enumerating
all smaller subsets in a fixed finite corpus. This experiment enumerates all
smaller subsets, so it reports both statuses separately. An insufficient path
reports `UNKNOWN` for cardinality minimum.

[Beckers, *Causal Explanations and XAI*](https://proceedings.mlr.press/v177/beckers22a.html)
is recorded as context for sufficient versus counterfactual explanations.
Natural-language generation, audience ranking, and probabilistic ranking are
not proof inputs.

## Fixed contract

The fixed corpus has three observed evidence records and one synthetic audit
record. It contains three paths:

| case | path | result |
| --- | --- | --- |
| `minimal` | source parsed → semantic IR bound → compiler receipt proven | sufficient, `SUBSET_MINIMAL`, `CARDINALITY_MINIMUM` |
| `overlong` | the same three plus synthetic audit noise | sufficient, `NOT_SUBSET_MINIMAL`, `NOT_CARDINALITY_MINIMUM` |
| `insufficient` | source parsed → compiler receipt proven | `FAIL_CLOSED`, cardinality `UNKNOWN` |

There are 7 single-removal counterfactual executions: 3/3 change the minimal
path decision, and 3/4 change the overlong path decision. Removing audit noise
does not change the decision. The finite corpus search executes 11 smaller
subsets for the three-member path and 15 for the four-member path, for 26
combination executions total.

The receipt contains six claims and 12 append-only transitions. A claim reaches
`DISCHARGED` only when its bound evidence passes; an unobserved claim remains
`OPEN`, and a counterexample becomes `REFUTED`. A regression fixture changes
the compiler receipt to `FAIL_CLOSED` and records that the old unconditional
discharge would be wrong.

Two semantic interventions change the result: changing the predicate value or
breaking the evidence relation makes the result `FAIL_CLOSED` and changes the
path/minimality evidence. A comment-only intervention changes the source
digest but preserves the canonical semantic digest and result.

CI binds repository before/after observations. The experiment is read-only:
workspace writes are 0 and promotion authority is false.
