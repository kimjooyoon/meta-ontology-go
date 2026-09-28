# CI failure reason catalog

This catalog is the human-readable companion to the versioned
`gooo/ci-failure/v2` manifest. The validator binds each failure code to its
class, severity, blocking scope, parallelization rule, and closed
`next_operation` value. V2 removes human handoff flags and owner routing from
the machine report.

The manifest binds terminal failures to the exact repository, source commit,
base/head, event, workflow, run, attempt, job, artifacts, and catalog digest.
Unknown or unavailable tuple values remain fail-closed. The workflow currently
reports a closed operation key with the evidence; emitting a key does not grant
the report permission to write source, refs, or branch settings.

The required machine fields include `schema`, `version`, failure codes and
classification, source/base/head and event/run/attempt/job tuple, provenance,
evidence references, catalog path/digest, rejections, missing reasons,
artifacts, ordered terminal failures, message, remediation, and
`next_operation`. The validator rejects caller-supplied classification or
operation drift. The catalog must contain exactly one row for every machine
failure code.

The following machine-catalog records are part of the catalog bytes. Their
operation keys form a closed vocabulary checked by the Go validator.

<!-- machine-catalog: CI-TEST-001|test|error|local|true|REPAIR_SOURCE -->
<!-- machine-catalog: CI-SCOPE-001|scope|error|global|false|RECOMPUTE_SCOPE -->
<!-- machine-catalog: CI-CAPS-001|caps|error|global|false|REPAIR_SOURCE -->
<!-- machine-catalog: CI-CONTRACT-001|contract|critical|global|false|RETAIN_UNRESOLVED -->
<!-- machine-catalog: CI-DEPENDENCY-001|dependency|warning|local|true|CONTINUE_LOCAL -->
<!-- machine-catalog: CI-GATE-001|gate|blocked|global|false|REOBSERVE_GATE -->
<!-- machine-catalog: CI-ARTIFACT-001|artifact|error|global|false|REBUILD_ARTIFACT -->
<!-- machine-catalog: CI-FRESHNESS-001|freshness|error|global|false|RERUN_CURRENT_HEAD -->
<!-- machine-catalog: CI-PROVENANCE-001|provenance|blocked|global|false|RECOMPUTE_PROVENANCE -->
<!-- machine-catalog: CI-PROMOTION-AUTH-001|gate|blocked|global|false|REBUILD_PROMOTION_PROOF -->
<!-- machine-catalog: CI-PROMOTION-OBSERVATION-001|gate|blocked|global|false|REOBSERVE_PROMOTION_TUPLE -->
<!-- machine-catalog: CI-ROOT-OF-TRUST-001|trust-root|blocked|global|false|REVALIDATE_TRUST_ROOT -->
<!-- machine-catalog: CI-ROOT-OF-TRUST-BOOTSTRAP-001|trust-root|blocked|global|false|REVALIDATE_TRUST_ROOT -->
<!-- machine-catalog: CI-UNCLASSIFIED-001|unclassified|blocked|global|false|CLASSIFY_FROM_EVIDENCE -->

## Codes

| Code | Class / severity | Criteria | System action | Safety invariant |
| --- | --- | --- | --- | --- |
| `CI-TEST-001` | test / error | A canonical implementation, formatting, vet, race, or conformance job is terminal and unsuccessful. | `REPAIR_SOURCE`: use exact-head job evidence to repair source, then recompute all checks. | Never weaken tests or treat an unsuccessful job as success. |
| `CI-SCOPE-001` | scope / error | Exact base/head changed-path evidence is missing or contradictory, or the PR target is invalid. | `RECOMPUTE_SCOPE`: derive scope again from pinned base/head commits. | Do not use branch aliases as scope evidence. |
| `CI-CAPS-001` | caps / error | DAMP file or DRY function cap is exceeded. | `REPAIR_SOURCE`: split the implicated unit, then recompute caps. | Do not relabel a cap failure as generic scope success. |
| `CI-CONTRACT-001` | contract / critical | Existing code and the requested semantic contract have incompatible meanings. | `RETAIN_UNRESOLVED`: preserve both meanings and keep the candidate blocked. | Do not invent a preference or silently overwrite either contract. |
| `CI-DEPENDENCY-001` | dependency / warning | A dependency outside this PR blocks only this PR's local proof or implementation. | `CONTINUE_LOCAL`: retain the dependency tuple and run independent work. | Do not turn a local dependency into a global stop. |
| `CI-GATE-001` | gate / blocked | A required canonical check or protected promotion predicate blocks a promotion. | `REOBSERVE_GATE`: fetch exact refs and reevaluate the required predicate. | Do not admin-bypass, force-push, or infer authorization. |
| `CI-ARTIFACT-001` | artifact / error | Required artifact count, name, size, expiry, or digest is missing or invalid. | `REBUILD_ARTIFACT`: regenerate from this exact run and verify its digest. | Do not reuse evidence from another run/attempt. |
| `CI-FRESHNESS-001` | freshness / error | A head, base, run, attempt, or job is stale, replayed, or mismatched. | `RERUN_CURRENT_HEAD`: bind fresh checks to the current exact head. | Do not reuse historical green evidence. |
| `CI-PROVENANCE-001` | provenance / blocked | Required provenance evidence is absent, unavailable, or unverifiable. | `RECOMPUTE_PROVENANCE`: fetch and bind the missing exact evidence. | Do not infer protection or provenance from CI output. |
| `CI-PROMOTION-AUTH-001` | gate / blocked | A proof bundle cannot produce a digest-bound fast-forward authorization for the exact `dev` tree to `main` route. | `REBUILD_PROMOTION_PROOF`: regenerate evidence from exact refs, checks, and artifacts. | Green jobs alone are not promotion authorization. |
| `CI-PROMOTION-OBSERVATION-001` | gate / blocked | The live same-repository `main` PR is draft, dirty, stale, merged, unknown, or not bound to current refs, exact source tree, and fast-forward topology. | `REOBSERVE_PROMOTION_TUPLE`: reread PR state and refs, then reevaluate. | Do not authorize from draft, stale snapshot, behind, divergent, or unknown state. |
| `CI-ROOT-OF-TRUST-001` | trust-root / blocked | The active policy source or trust-root digest is missing or inconsistent. | `REVALIDATE_TRUST_ROOT`: recompute the pinned policy and workflow identities. | Never accept candidate-controlled trust policy as authority for itself. |
| `CI-ROOT-OF-TRUST-BOOTSTRAP-001` | trust-root / blocked | A bootstrap transition cannot prove the predecessor and successor policy relation. | `REVALIDATE_TRUST_ROOT`: rebuild the exact predecessor/successor evidence. | Never infer bootstrap authorization from a successful candidate check. |
| `CI-UNCLASSIFIED-001` | unclassified / blocked | A terminal failure cannot be safely classified. | `CLASSIFY_FROM_EVIDENCE`: classify exact terminal jobs and retain unknown fields. | Do not guess a reason or treat it as green. |

Any missing machine observation must carry a non-empty reason. Missing
credentials or APIs are not permission to guess: use a fail-closed gate or
provenance result. A system consumer can resolve the operation key and replay
the exact evidence without assigning a person.
