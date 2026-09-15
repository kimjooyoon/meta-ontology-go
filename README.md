# meta-ontology-go

`meta-ontology-go` is an experimental semantic compiler implemented in Go. Its
surface files use the `.gooo` extension: **Go Of Ontology**.

The project has one deliberate authority boundary: `.gooo` declarations express
business intent. The compiler lowers that view to a normalized semantic IR and
projects structural boundaries to Go. Handwritten Go remains the source of truth
for irreducible implementation logic; generated Go, query output, documentation,
and CI results are derived views or evidence.

```text
.gooo intent ──lower──> semantic IR ──project──> generated Go
     │                         │                    │
     │                         └── facts/evidence   └── handwritten slots
     └── source spans, IDs, and explicit assertions
```

<!-- PUBLIC-TRUST-BADGES:BEGIN -->
### Public trust surface

These badges are generated from the lowered public-trust `.gooo` policy. Workflow badges report workflow results; they do not claim branch-protection or ruleset enforcement.

#### Language / Release

[![Go 1.27.0 toolchain](https://img.shields.io/badge/Go-1.27.0-00ADD8?logo=go&logoColor=white)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/go.mod)
[![Published release v0.4.0-dev](https://img.shields.io/github/v/release/kimjooyoon/meta-ontology-go?include_prereleases&label=published%20release)](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.4.0-dev)

#### Build / Conformance

[![CI workflow result](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/ci.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/ci.yml)
[![Compiler compatibility evidence](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/self-improvement-compiler-compatibility.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/self-improvement-compiler-compatibility.yml)
[![Experimental release readiness](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/gooo-release-readiness.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/gooo-release-readiness.yml)

#### Security / Supply Chain

[![CodeQL code scanning configured](https://img.shields.io/badge/CodeQL-code%20scanning-2ea44f?logo=github)](https://github.com/kimjooyoon/meta-ontology-go/security/code-scanning)
[![Dependabot weekly updates](https://img.shields.io/badge/Dependabot-weekly%20updates-0366d6?logo=dependabot)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/.github/dependabot.yml)
[![Dependency review on pull requests](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/dependency-review.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/dependency-review.yml)
[![Private vulnerability reporting enabled](https://img.shields.io/badge/Private%20vulnerability%20reporting-enabled-2ea44f?logo=github)](https://github.com/kimjooyoon/meta-ontology-go/security/advisories/new)

#### Evidence / Project Health

[![Project status: experimental](https://img.shields.io/badge/Project-experimental-orange)](https://github.com/kimjooyoon/meta-ontology-go#project-status)

#### Community

[![MIT licensed](https://img.shields.io/github/license/kimjooyoon/meta-ontology-go)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/LICENSE)

The complete row ledger, including unavailable and refuted claims, is emitted by the `Public trust surface` workflow.
<!-- PUBLIC-TRUST-BADGES:END -->

The repository is intentionally small. The supported language sketch currently
covers packages, namespaces, entities with URI-like IDs, and activities with
entity inputs and an entity result. The semantic kernel also defines a
PROV-inspired vocabulary, deterministic normalization, explicit candidate facts,
and marker-based generated regions. See [the architecture](docs/architecture.md)
and [the language sketch](docs/spec.md).

## The deterministic pressure loop

[![Animated explanation of the semantic self-improvement loop: policy-defined base metrics rise from an observed floor; a deterministic selector chooses a focus subset while every baseline remains guarded; 100 heuristic attempts fan out; single-pressure regressions are rejected; source-backed evidence and path proof let deterministic CI requalify all dimensions; the verified ceiling ratchets into epoch 2's immutable floor](docs/assets/metric-pressure-loop/metric-pressure-loop.gif)](docs/assets/metric-pressure-loop/metric-pressure-loop.png)

The [static PNG preview](docs/assets/metric-pressure-loop/metric-pressure-loop.png)
is useful in viewers that do not animate GIFs. The loop is illustrative: each
system's protected policy/SPI declares its own `N` base metrics, `M` cross
pressures, and active `K`; the language guarantees at least **two independent,
non-compensating pressure dimensions**, not one universal set of numbers. The
animation uses `N=6`, `M=4`, and `K=2` only as one concrete policy instance, and
the 100 parallel agents are an illustrative workload. Every agent focuses only
on the selected subset while all `N` baseline metrics remain non-regression
floors. A performance gain that damages completeness, or the reverse, is
rejected. Attempts may use local inference and may PASS, FAIL, or remain
UNKNOWN, but missing evaluator/oracle/evidence is fail-closed: **Deterministic
CI — not inference** checks exact source-backed evidence across the full
declared vector. Proof computes remaining viable paths, and only a qualified
ceiling becomes the next immutable floor with the next metric/SPI and
provenance obligations. Self-improvement here means verified contract,
evaluator, and evidence gains compound; agents do not rewrite the judge or
lower thresholds.

## The ontology in motion

These ten deterministic GIFs are one visual denominator, generated by the shared
renderer in [`docs/assets/ontology-visuals`](docs/assets/ontology-visuals) and
bound to the source concepts listed in [`visual-manifest.json`](docs/assets/ontology-visuals/visual-manifest.json).
They use labels and shapes as well as color, so CLOSED, UNKNOWN, and REFUTED are
not color-only states. The exact checked-in asset count, byte total, and SHA-256
digests are recorded in [`generated-asset-lock.json`](docs/assets/ontology-visuals/generated-asset-lock.json).

<table>
<tr>
<td width="50%"><img src="docs/assets/ontology-visuals/01-intent-ir-lowering.gif" alt="Animated engineering story showing a billing .gooo activity becoming inspectable semantic IR, generated Go source, and a PASS receipt." width="460"><br><strong>1. Source to generated Go</strong><br>A billing declaration becomes generated Go plus a receipt.</td>
<td width="50%"><img src="docs/assets/ontology-visuals/02-authority-boundary.gif" alt="Animated engineering story showing a declaration becoming typed Order, PayOrder, and Receipt semantic nodes with used and wasGeneratedBy edges consumed by a backend." width="460"><br><strong>2. Semantic IR as graph</strong><br>A typed graph is consumed by the backend.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/03-munchausen-proof-choice.gif" alt="Animated engineering story showing activity.gooo and entities.gooo reorder into canonical filename order, merge into a billing package API, and emit a PayOrder receipt." width="460"><br><strong>3. Multifile package resolution</strong><br>Unordered source files become a deterministic package API.</td>
<td><img src="docs/assets/ontology-visuals/04-claim-evidence-lifecycle.gif" alt="Animated three-lane engineering handoff where an author writes main.gooo, a compiler creates generated.go, and a reviewer consumes receipt.json across a no-cross-write boundary." width="460"><br><strong>4. Agent handoff by receipt</strong><br>Author, compiler, and reviewer agents exchange caller-owned artifacts.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/05-unknown-cause-descent.gif" alt="Animated engineering story showing a missing artifact create stage, step, reason, class, next_operation, and blocked_by fields before evidence re-evaluates the same claim as CLOSED." width="460"><br><strong>5. UNKNOWN to CLOSED resolution</strong><br>Six fields guide a resolver to a new evidence-backed decision.</td>
<td><img src="docs/assets/ontology-visuals/06-precedence-counterexample.gif" alt="Animated engineering story where a counterexample for the same claim enters a precedence stack, selects REFUTED over UNKNOWN, and appends the preserved record." width="460"><br><strong>6. Refutation precedence</strong><br>A known contradiction wins and remains in the ledger.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/07-package-resolution.gif" alt="Animated engineering story showing pinned source, contract, and toolchain digests produce identical output twice, then a changed byte creates a mismatch and blocks adoption." width="460"><br><strong>7. Deterministic replay block</strong><br>Byte drift blocks adoption after replay comparison.</td>
<td><img src="docs/assets/ontology-visuals/08-incremental-conformance.gif" alt="Animated engineering story showing a changed subject with six-digest identity routed to one REUSE receipt, with EXECUTE, UNKNOWN, and REFUTED kept as inactive alternatives." width="460"><br><strong>8. Incremental conformance router</strong><br>One identity selects one reused receipt.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/09-bootstrap-oracle.gif" alt="Animated engineering story moving from observed metric bug to meta-rule change, exact-head CI, dev adoption, post-adoption receipt, and main eligibility in causal lanes." width="460"><br><strong>9. Self-improvement gate cascade</strong><br>Each receipt unlocks the next engineering gate.</td>
<td><img src="docs/assets/ontology-visuals/10-promotion-lineage.gif" alt="Animated experimental domain projection showing OpenAPI-style service facts and OpenTofu plan facts connecting through a proposed Gooo contract to an UNKNOWN mismatch dossier without infrastructure mutation." width="460"><br><strong>10. Experimental domain projection</strong><br>Proposed cross-domain projection remains UNKNOWN until implemented evidence exists.</td>
</tr>
</table>

Regenerate the complete set with `go run ./docs/assets/ontology-visuals`.

Regenerate and verify the checked-in media with:

```sh
go run ./docs/assets/metric-pressure-loop
go run ./docs/assets/metric-pressure-loop -check
```

## Quick start

Run the repository checks from the project root:

```sh
gofmt -l .
go test ./...
go vet ./...
go run ./cmd/gooo check examples/billing/main.gooo
```

The conformance walkthrough in [docs/conformance.md](docs/conformance.md) uses
the checked-in examples and shows the expected command shapes. It also calls out
which CLI surfaces are not yet supported. The current GitHub Actions workflow is
documented in [CONTRIBUTING.md](CONTRIBUTING.md); it should be treated as the
source of truth for required CI, not as a promise of future compiler features.

## Branch and promotion contract

Work branches target `dev`. The only promotion route is an exact,
same-repository `dev`-to-`main` pull request; no intermediary branch is part of
the current contract. Governance is `ci_only`: review and approval fields do not
authorize a protected-branch promotion.

The six canonical proof jobs are `gofmt`, `go vet`, `go test`, `go test -race`,
`Semantic conformance`, and `CI policy`. Those names identify workflow evidence,
not branch-protection enforcement. The v19 live audit found no required status
checks on `dev`; `main` retains its existing seven-context branch-protection
listing, while the named `dev` and `main` rulesets are disabled. No badge above
turns a workflow result into an enforcement claim.

The root README reaches the default `main` page only after a later, legitimate
protected-main promotion. This v19 change is a `dev`-targeted trust-surface
update and does not alter protection, rulesets, or the promotion route.

For the promotion route, CI emits a digest-bound `promotion_authorization` with
`source=dev`, `target=main`, and `operation=fast_forward`. It passes only for
fresh exact refs and topology (`ahead > 0`, `behind = 0`, `main` as merge base),
the required proof and Guardian evidence, both exact seven-context protection
snapshots, and a clean, open, non-draft, unmerged same-repository pull request.
The proof producer never mutates refs or protection. After a final exact reread,
only a normal CAS/fast-forward update is allowed; force updates are prohibited.

## Project status

This is an experimental language, not a stable application framework. In
particular, the repository does not currently promise a production LSP, a stable
`analyze` CLI, automatic promotion of ambiguous Go observations, or durable
provenance publishing. Internal packages and design notes may describe those
directions, but a feature is supported only when its command/API and conformance
evidence are present.

## Governance

[AGENTS.md](AGENTS.md) defines authority boundaries and agent roles.
[CONTRIBUTING.md](CONTRIBUTING.md) defines branch, PR, review, and CI workflow.
[docs/governance.md](docs/governance.md) records the SSOT boundary, BX laws, line
caps, and evidence policy. [docs/metrics-rfc.md](docs/metrics-rfc.md) defines
the design-only deterministic metric contract. [docs/conformance.md](docs/conformance.md)
is the runnable example index. The [Deterministic CI Evolution Retrospective](docs/deterministic-ci-evolution.md)
is the append-only read-only evidence record.

[W3C PROV-O]: https://www.w3.org/TR/prov-o/
