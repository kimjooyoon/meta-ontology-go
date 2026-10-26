# Recovery decisions with original observations

This case keeps an image that was already published and proposes a separate
operation for its original build provenance. Gooo joins the declared scope,
publication record and independent readback before reusing `ProposeRecovery`.
Display names do not identify external effects: image publication and build
provenance transport have different operation IDs over the same image subject.

The retained records are from
[publication run 39672180710, attempt 1](https://github.com/kimjooyoon/gooo-go-ports/actions/runs/39672180710)
and [independent readback 39673212888, attempt 1](https://github.com/kimjooyoon/gooo-go-ports/actions/runs/39673212888).
The original publisher's push step succeeded and its attestation step failed.
The later readback observed the same manifest subject and image configuration.
The [original workflow](https://github.com/kimjooyoon/gooo-go-ports/blob/6ff4e3a6c3baf9b518838bd35d2429decdf991bf/.github/workflows/service-grant-mcp-oci.yml)
remains the prior-plan reference. The goal and proposed revision IDs in
`testdata/scope.json` are authored retrospective annotations, not IDs that the
old workflow emitted. The successful recovery run 39687579594 is **not an input**.

## Run and change the policy

From the repository root:

```sh
go run ./cmd/gooo package generate --go-package recoveryorigin \
  --out examples/ci-strategy/recovery-origin \
  examples/ci-strategy/recovery-origin.workspace.json
go test ./examples/ci-strategy/recovery-origin -v
```

The generated `Evaluate(scope, published, readback)` entry is an ordinary Go
library. `Read(directory, role, jobID)` reads the retained API snapshots and ZIP
members, hashes their original bytes and exposes typed observations. Gooo owns
the origin joins, bounded step lookup, effect identity, revision and next action.
The Go reader only decodes records, reads bounded files and copies fields. It
returns the raw file digests separately so a caller can retain its input origins.

The declared scope pins the repository, source, workflow bytes, run/attempt/job,
artifact names, relevant steps and observation cutoff. Changing a run alone
cannot reuse the prior scope. `observed_at` bounds this historical observation;
it is not a freshness guarantee about the registry today. Step meanings are
declared against that exact workflow, rather than guessed from display text.

| Gooo result | Next action |
| --- | --- |
| `PROPOSED_RECOVERY` / `BOUND_HISTORICAL_OBSERVATION` | Keep the observed publication; assess the proof-transport proposal |
| `REJECTED_REPLAY` | Revise a proposal that repeats the observed publication, even under a new display name |
| `UNRESOLVED_ORIGIN` / `REFRESH_ORIGIN` | Collect or correct the source records; completed-effect IDs stay empty |
| `REVISE_SCOPE` | Supply a supported operation over the declared subject |

Every result retains `execution_authorized=false`. The next operation still
needs its own available proof payload, current-state check, budget, executor and
independent post-effect observation. An attestation-step failure alone does not
establish which payload is available for retry.

## What is measured

These are regression denominators for this one case, not general language or
recovery accuracy percentages. Expected outcomes are handwritten separately
from the generated implementation.

| Observation | Denominator |
| --- | --- |
| Actual retained history yields the expected proposal | 1 / 1 history |
| Wrong/missing/contradicted origin remains unresolved | 20 / 20 normalized counterexamples |
| Same publication under different display names is rejected | 2 / 2 names |
| Different subject or unknown operation cannot widen scope | 2 / 2 proposals |
| Changed or missing raw input cannot become a completed effect | 6 / 6 reader counterexamples |
| Goal, plan, revision, activity and original observation IDs retained | 7 / 7 checked identities in the actual history |

Initial implementation adds 175 authored Gooo lines, 204 handwritten production
Go lines for IO/decoding, and a 363-line generated projection. This is initial
adapter cost; it does not demonstrate a declining marginal Go cost yet. The
existing recovery selector is reused unchanged. A separate consistency test
regenerates the library from both Gooo sources and checks its source bindings.
No new grammar capability is counted and no timing improvement is claimed.

`testdata/history.zip` losslessly retains nine original files (132,069 expanded
bytes, 65,767 archive bytes). `testdata/origins.json` records their original names
and SHA-256 values. The test pins the whole retained archive independently.
No new registry publication, image download, model call, training or performance
experiment runs in these tests.

## Collection boundary

The adapter reads already collected records. The collector must establish the
GitHub API snapshots' authenticity and the source owner's declared scope before
using them. This example does not authenticate edited local JSON, verify a
signature, fetch current registry state, or establish that arbitrary future
workflows have the same step meanings. The fixture hashes identify the reviewed
historical bytes; matching a digest alone does not establish a new producer's
authority. These remaining operation boundaries belong to
[#1762](https://github.com/kimjooyoon/meta-ontology-go/issues/1762).
