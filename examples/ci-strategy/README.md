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

`ProposeRecovery` expresses the next plan after a partial effect and a failed
activity. It keeps the prior and new plan IDs, failed activity, failure
observation, completed effects and proposed recovery effects in one result. A
recovery that repeats a completed effect is `REJECTED_REPLAY`; duplicate or
missing IDs are `INVALID_PROPOSAL`. With no declared overlap and complete
caller assertions, the result is `PROPOSED_RECOVERY`, still with
`execution_authorized=false`. An absent failure/effect observation or open
operation scope remains `INSUFFICIENT_EVIDENCE`. Inputs are capped at 32 effects
per side, so the comparison has bounded work.

From the repository root:

```sh
go run ./cmd/gooo check examples/ci-strategy/strategy.gooo
go run ./cmd/gooo package execute --json --cases \
  examples/ci-strategy/selection.cases.json \
  examples/ci-strategy/gooo.workspace.json | jq -r .decision
go run ./cmd/gooo package execute --json --cases \
  examples/ci-strategy/outcome.cases.json \
  examples/ci-strategy/outcome.workspace.json | jq -r .decision
go run ./cmd/gooo package execute --json --cases \
  examples/ci-strategy/recovery.cases.json \
  examples/ci-strategy/recovery.workspace.json | jq -r .decision
```

The four commands currently report `ok`, `PASS`, `PASS`, and `PASS`. The first
selection case proposes the second plan. The other rejects a small predicted
gain. The outcome case has a faster number but a different runner class, so it
remains insufficient evidence. `PASS` means the finite expected cases matched
the Gooo implementation; it is not a production performance result.
The recovery cases cover a proof-upload retry after an image was published,
an attempted replay of the image publication, an unobserved completed effect,
and a duplicate recovery effect.

| Owned by `.gooo` source | Supplied from outside |
| --- | --- |
| IDs, selection and revision rules | Run data, proposals and effect observations |

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
The recovery booleans and effect IDs are also caller assertions. Their presence
does not prove that an external effect occurred or that a retry is idempotent.
An execution adapter must bind those claims to independent receipts and check
current scope before it can grant any effect authority.

The [recovery-origin case](recovery-origin/README.md) now connects this selector
to retained publication and independent readback records. Gooo checks their
source/run/attempt/artifact bindings and canonical operation identity; an
unresolved origin produces no completed-effect IDs. The original selector above
still accepts caller assertions. The new case records historical observations
and keeps execution disabled while current operation authority remains open.
