# A small Gooo strategy case

This example is a source-owned decision about a proposed CI improvement. The
numbers in the cases are **synthetic**. It makes no claim about this repository's
measured CI speed.

`SelectPlan` receives one goal and a bounded list of alternatives. A plan is eligible only
when its operation scope is declared closed, its required-check digest matches
the goal, and its predicted gain exceeds the minimum. The selected result is
`PROPOSED_FOR_EXPERIMENT`; `execution_authorized` remains `false`. When the
gain is too small, no plan is selected. The source currently accepts 2–32
alternatives and records the option and eligible counts. Plans with duplicate IDs
or mismatched parallel-field lengths invalidate the request. This is a bounded
case, not a general strategy grammar.

`AssessOutcome` reads a later observation under the same goal and plan IDs. It
requires a distinct run, matching workload, toolchain and runner class, and
passed required checks before reporting an observed gain. A missing comparison
condition stays `INSUFFICIENT_EVIDENCE`. The separate `revision_id` identifies
the new record; it does not rewrite an older plan. Even `OBSERVED_GAIN` does
not claim that the plan caused the difference.

From the repository root:

```sh
go run ./cmd/gooo check examples/ci-strategy/strategy.gooo
go run ./cmd/gooo package execute --json --cases \
  examples/ci-strategy/selection.cases.json \
  examples/ci-strategy/gooo.workspace.json | jq -r .decision
go run ./cmd/gooo package execute --json --cases \
  examples/ci-strategy/outcome.cases.json \
  examples/ci-strategy/outcome.workspace.json | jq -r .decision
```

The three commands currently report `ok`, `PASS`, and `PASS`. The first
selection case proposes the second plan. The other rejects a small predicted
gain. The outcome case has a faster number but a different runner class, so it
remains insufficient evidence. `PASS` means the finite expected cases matched
the Gooo implementation; it is not a production performance result.

| Owned by `.gooo` source | Supplied from outside |
| --- | --- |
| Stable field and activity IDs, eligibility, tie order, rejection and revision states | Run IDs, timings, checks digest, plan predictions, assumptions and comparison conditions |

The execution receipt carries source and condition origins for this example.
The syntax corpus inventories this existing-feature program as a meta source;
it does not increase the fixed syntax capability denominator.
It does not yet emit separate [PROV-O](https://www.w3.org/TR/prov-o/) entities
and activities for goal proposal, plan derivation, execution and independent
observation. In particular, the booleans and digests in input are asserted by
the caller, not verified by a CI adapter. Binding them to exact GitHub run
records and to an execution-plan provenance record is the next step in
[#1762](https://github.com/kimjooyoon/meta-ontology-go/issues/1762). An agent
may suggest values, but this example cannot grant repository write authority.
