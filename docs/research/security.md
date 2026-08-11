# Security threat model and CI security gates

Status: design research for the compiler, LSP, generator, and verification loop.
This document proposes controls; it does not change the implementation or make a
future control true merely by documenting it.

## Security objective

The security boundary is the path from authoritative `.gooo` intent to a generated
Go projection and its evidence. An attacker must not be able to turn a workspace
input, an editor request, a generated file, or a provenance record into:

- a read or write outside the authorized workspace/output roots;
- execution of an attacker-selected command, hook, or executable;
- silent replacement of source, handwritten logic, or unrelated generated output;
- a false semantic identity, scope approval, freshness claim, or evidence chain; or
- unbounded CPU, memory, file-descriptor, process, or disk consumption.

The intended security property is therefore:

```text
untrusted input
    -> bounded parse/analyze
    -> source-backed semantic delta
    -> independently checked projection
    -> append-only, hash-bound evidence
    -> least-privilege build artifact
```

This is a compiler security problem as well as a supply-chain problem. Generated Go
is executable output, but the compiler also reads and writes files, starts tools,
serves an LSP protocol, and makes semantic/provenance decisions that CI may use for
authorization.

## System and trust boundaries

```mermaid
flowchart LR
    U["User, PR, or LSP client"] --> W["Untrusted workspace input"]
    W --> P["Parser / analyzer / LSP"]
    P --> I["Semantic IR and delta"]
    I --> G["Generator and generated Go"]
    I --> E["Provenance and evidence"]
    G --> B["Build / test runner"]
    E --> V["Independent verifier / CI gate"]
    V --> M["Merge or release decision"]
    K["Protected ontology, verifier, and workflow"] --> V
    K --> P
```

### Trusted or protected

- The operating system, Go runtime, repository rules, and the isolated CI runner.
- Maintainer-controlled ontology vocabulary, semantic identity rules, verifier
  semantics, CI policy, and workflow definitions.
- The independent gate that checks semantic scope, generated integrity, freshness,
  and evidence. No Builder should be able to change and approve this gate.

### Untrusted or attacker-influenced

- `.gooo`, Go, generated-output, workspace configuration, file names, symlinks, and
  contents in a checkout, including a pull request from a fork.
- LSP messages, document URIs, workspace folders, cancellation behavior, and client
  capabilities.
- Semantic names, aliases, asserted relations, source spans, hashes, timestamps,
  evidence payloads, and any data copied from generated output.
- Environment variables, PATH entries, Git refs, command arguments, and tool output
  derived from a workspace or PR event.

### Assumptions and non-goals

This model assumes the host kernel and the protected GitHub/repository account are not
already compromised. It does not attempt to defend against a user who intentionally
grants the compiler full local-machine privileges, or against a maintainer-approved
change to the protected kernel. Those are separate operational controls. It does
cover a malicious project, malicious PR author, compromised LSP client, stale cache,
and a compromised or misleading generated projection.

## Assets and impact

| Asset | Security property | Impact if compromised |
| --- | --- | --- |
| `.gooo` and handwritten Go | Integrity and locality | Wrong business behavior, source loss, or unauthorized semantic change |
| Semantic IDs and IR | Identity, namespace isolation, consistency | Scope bypass, confused-deputy relations, incorrect generated code |
| Generated Go and source maps | Determinism and correspondence | Backdoor, build drift, or reviewable source no longer matching output |
| Provenance/evidence | Authenticity, freshness, append-only history | False claim that an artifact was reviewed, tested, or generated from trusted input |
| Workspace and runner files | Confidentiality and containment | Secret/source disclosure, overwrite of repository or host files |
| CPU, memory, processes, descriptors, and disk | Availability | LSP/CI denial of service and runner starvation |
| CI credentials and tokens | Confidentiality and least privilege | Repository takeover, artifact tampering, or secret exfiltration |

## Threat register

Severity is the expected consequence after a plausible exploit, not a statement that
the current skeleton already contains the full vulnerable implementation.

| ID | Threat and abuse path | Primary assets | Severity | Required security property |
| --- | --- | --- | --- | --- |
| T1 | Path traversal or symlink escape in input, cache, source-map, or `--out` path | Workspace, source, runner | High | Every filesystem operation is contained by an authorized root after canonicalization and symlink checks |
| T2 | Arbitrary command through shell interpolation, executable lookup, hooks, or PR-controlled arguments | Runner, tokens, repository | Critical | Only fixed, allowlisted programs with argv separation, scrubbed environment, timeout, and no shell are executable |
| T3 | Generated-file overwrite through collision, forged marker, symlink, or partial write | Source, handwritten logic, generated output | High | Generation is fail-closed, atomic, provenance-bound, and refuses non-generated or ambiguous targets |
| T4 | Provenance spoofing through forged IDs, source spans, hashes, evidence, or replayed receipts | IR, evidence, merge decision | Critical | Evidence is source-bound, append-only, fresh, canonicalized, and independently verified |
| T5 | Untrusted workspace or PR code reaches privileged CI, secrets, network, hooks, or the verifier itself | Runner, tokens, verifier policy | Critical | Untrusted code runs with read-only, no-secret, isolated privileges; protected checks run from trusted code |
| T6 | Resource exhaustion through parser depth, graph fan-out, LSP floods, generated explosion, cache churn, or child-process output | CI/LSP availability | High | Explicit budgets, cancellation, bounded queues/output, quotas, and deterministic rejection |
| T7 | Semantic namespace confusion or identity collision caused by display names/aliases | IR, generated code, scope | High | Stable absolute IDs and namespaces are canonical identity; names never authorize a relation |
| T8 | Stale or cross-build cache/evidence is accepted as current | Generated output, evidence | High | Cache keys and evidence bind source, policy, toolchain, options, and revision; freshness is checked before use |

T1–T6 are the requested primary threats. T7 and T8 are included because they turn a
successful filesystem or CI attack into a semantic authorization bypass.

## Threat analysis and controls

### T1 — Path traversal, symlink escape, and path confusion

Relevant inputs include source URIs received by LSP, `--out` and cache directories,
generated file names derived from IDs or names, source-map paths, workspace roots, and
Git paths. `filepath.Clean` alone is insufficient: it does not establish that a path
stays below a root when symlinks, junctions/reparse points, absolute paths, alternate
separators, or a time-of-check/time-of-use race are involved. This is the class
described by [CWE-22](https://cwe.mitre.org/data/definitions/22.html).

Required design rules:

1. Treat a workspace root and an output root as capabilities, not as strings supplied
   by the document. Reject NUL bytes, empty/ambiguous roots, absolute paths, drive or
   UNC paths where they are not explicitly supported, and components that resolve to
   `..`.
2. For every derived target, canonicalize the root and candidate, require the
   candidate to be a descendant of the root using a component-aware comparison, and
   reject symlinks/junctions/reparse points in the path. Re-check at open/rename time
   to reduce TOCTOU exposure; prefer no-follow or directory-handle primitives where
   the platform supports them.
3. Generate file names from a constrained, deterministic mapping of stable semantic
   IDs. Do not concatenate a display name, alias, user-provided path, or URI directly
   into a filesystem path. A path map must reject collisions after normalization.
4. Keep caches and temporary files below an ephemeral, per-invocation directory. Do
   not follow symlinks from a cache entry, and do not treat cache contents as trusted
   input without checking the content digest and schema version.
5. Restrict LSP `file:` URIs to registered workspace roots. A client request must not
   make the server read arbitrary local files, and a workspace folder must not widen
   the root without an explicit trusted configuration decision.

CI tests should exercise `../`, absolute, drive/UNC, separator-mixed, Unicode-normalized,
symlinked, and pre-existing target paths. Tests must verify both read and write
   behavior and must leave no file outside a temporary root.

### T2 — Arbitrary command execution

The current verifier uses `os/exec` with a fixed `git` command and separate arguments,
which is safer than interpolating a shell command. The Go `os/exec` package does not
invoke a shell by default, but it still runs a real process with inherited working
directory and environment unless the caller constrains them. Future compiler features
must not turn a DSL field, semantic name, LSP message, Git ref, workspace config, or
generated string into a program name or shell fragment.

Required design rules:

- Never use `sh -c`, `bash -c`, `cmd.exe /c`, or equivalent composition for workspace
  data. Use `exec.CommandContext` with a fixed executable and an argv list. Validate
  Git revisions as full object IDs or an allowlisted ref grammar, and terminate option
  parsing with `--` wherever the invoked tool supports it.
- Prefer an absolute, configured tool path or a verified toolchain directory over a
  mutable PATH. Reject `ErrDot` and executables found in the workspace. Do not execute
  files produced by the project, including generated Go, scripts, hooks, or binaries,
  during parse, LSP, or verification stages.
- Build an allowlist of commands required by each stage. A command not on the stage's
  allowlist is a policy failure, not a best-effort fallback. The generator itself
  should not need a command runner.
- Set a minimal environment, safe working directory, closed standard input, bounded
  stdout/stderr, and a context deadline. Use a process-group strategy where supported
  so cancellation also handles descendants. Record command identity and exit status,
  but never record secret-bearing environment values.
- Disable repository hooks and user/global Git configuration for CI operations. Do not
  pass secrets to jobs that inspect untrusted source. Network access should be denied
  by default for parse, generate, analyze, and LSP jobs.

The security test suite should poison PATH, include shell metacharacters in every
string-shaped input, use a fake executable in the workspace, provide a hanging child,
and emit more output than the configured limit. A passing test must demonstrate that
no unexpected process starts and that cancellation returns within the budget.

### T3 — Generated-file overwrite and locality failure

Generated markers are useful for locality but are not an authorization mechanism:
attacker-controlled text can contain plausible markers. A generator that blindly
opens `out/name.go`, replaces the first marker pair, or follows a symlink can destroy
handwritten logic or overwrite an unrelated file.

Required generation contract:

1. Resolve and authorize the output root before planning files. Produce a complete,
   deterministic plan in memory: target path, expected generated-region hash, source
   semantic IDs, and output digest. Abort before writing if any target is outside the
   root, collides after normalization, or exceeds file/count/byte budgets.
2. If a target does not exist, create it only under the authorized root. If it exists,
   require a valid, unique generated-region structure and a matching prior provenance
   record. Refuse to overwrite a non-generated file, malformed marker pair, duplicate
   marker, symlink, or file whose handwritten slot cannot be preserved exactly.
3. Write a temporary file in the same authorized directory, apply restrictive mode,
   flush and close it, verify the digest and formatting, then atomically rename it.
   Never truncate the destination before all checks pass. On failure, leave the prior
   file unchanged and clean up only the invocation's temporary files.
4. Generated output must include stable semantic IDs and source-map links, but those
   links are checked against the current DSL/IR. Output text alone cannot authorize a
   new relation or overwrite.
5. A freshness gate must run generation twice from identical inputs and compare bytes,
   markers, source maps, and file manifests. A locality test must prove that changing
   one semantic subgraph does not rewrite unrelated regions or handwritten slots.

### T4 — Provenance spoofing and false evidence

The project treats `.gooo` as authoritative for business intent, semantic IDs as
identity, and provenance/evidence as append-only records. This distinction is a
security boundary. A generated comment, Go symbol name, user assertion, or evidence
payload must not be able to promote itself into authoritative intent merely because it
looks canonical.

Required evidence rules:

- Canonicalize and validate absolute IDs, namespaces, relation vocabulary, source
  spans, and hashes before storing them. Reject duplicate IDs, namespace confusion,
  invalid spans, ambiguous aliases, and an ID whose meaning changes when its display
  name changes.
- Classify observations as syntactic, deterministic, or candidate. Only a
  source-backed deterministic fact with a valid span and matching input digest may
  update the IR. Candidate facts remain reviewable evidence and cannot authorize
  scope, merge, release, or capability use.
- Bind every receipt to the exact source bytes or commit, DSL/IR fingerprint,
  generator/verifier identity, policy version, toolchain, options, dependency lock,
  runner identity, and output digest. A receipt for a prior HEAD, different policy,
  or different generator is stale, even if its text is identical.
- Keep evidence append-only and linked to the activity that produced and verified it.
  Do not allow an input file to claim that it was verified by copying a `verified`
  field from generated output. The independent verifier must recompute the claim from
  source, output, and evidence digests.
- Require freshness and anti-replay checks: current revision, expected parent/build
  context, unique invocation identity, and non-reuse of a receipt across incompatible
  inputs. Signatures or attestations are useful only after their subject digest and
  signer trust policy are independently checked.

The intended model aligns with [W3C PROV-O](https://www.w3.org/TR/prov-o/): an
`Entity`, `Activity`, and `Agent` describe what was used, produced, and associated,
but a PROV-shaped record is not automatically proof. A proof is a validated,
source-bound provenance path satisfying the protected claim and policy.

### T5 — Untrusted workspace and CI execution

An untrusted checkout is data for the compiler and potentially code for the Go build;
it must not receive the privileges of the verifier or release job. The current
workflow is a good baseline in two respects: it uses `pull_request`, declares
`contents: read`, and does not provide secrets to the shown jobs. It still checks out
PR content and executes repository-controlled Go, and its action references are tags
(`actions/checkout@v4`, `actions/setup-go@v5`) rather than immutable commit SHAs.

The most important gap is verifier self-modification: a PR can change files under the
currently allowed `scripts/**` or `internal/verify/**` scope and then have the policy
job execute the changed verifier. A check that the candidate can rewrite is not an
independent gate. Path scope alone cannot solve this because the policy implementation
is itself in the candidate tree.

Required CI rules:

- Use untrusted `pull_request` jobs for parse, test, generate, and static checks with
  no secrets, read-only token, ephemeral runner, restricted network, and no artifact
  or cache reuse that can write into a trusted job's namespace. Never check out an
  untrusted PR in a privileged `pull_request_target` or `workflow_run` job.
- Run the protected verifier, branch policy, ontology policy, and security checks from
  the trusted base revision or a protected reusable workflow. Compare protected files
  against the base and require designated maintainers for changes to them. Do not let
  the candidate select the verifier, policy version, or required checks.
- Pin third-party actions to full commit SHAs, explicitly declare minimal permissions,
  restrict allowed actions, and keep workflow expressions out of shell source. GitHub
  documents branch names, PR titles, bodies, and similar context values as untrusted
  input for script-injection purposes; pass them as environment data only after
  validation, never by interpolation into a command.
- Separate Builder, Guardian, and Gate roles. The author of a feature or evidence
  record must not be the sole approver of its semantic scope or provenance. A draft PR
  targeting `integration` remains subject to the trusted gate before merge.
- Treat uploaded artifacts, caches, generated files, and test reports from an
  untrusted job as untrusted inputs. Verify their digest and schema in a fresh job;
  never execute them or restore them into a privileged workspace.

These controls follow GitHub's [script-injection guidance](https://docs.github.com/en/actions/concepts/security/script-injections),
[secure-use guidance](https://docs.github.com/en/actions/reference/security/secure-use),
and [workflow execution protections](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/actions-policies/workflow-execution-protections).

### T6 — Resource exhaustion and denial of service

The compiler and LSP must assume hostile size and shape, not only hostile syntax.
Examples include deeply nested expressions, a declaration/fact fan-out that creates a
quadratic closure, repeated identical LSP requests, huge diagnostics, generated-file
explosion, cache-key cardinality, a child process that never closes output, and a
client that opens many documents concurrently.

Every stage needs a budget with deterministic failure and cancellation. Initial values
below are starting policy candidates, to be measured against real projects before
being made compatibility commitments:

| Budget | Initial ceiling | Failure behavior |
| --- | ---: | --- |
| One source document | 8 MiB | Reject before full parse |
| One LSP message | 1 MiB | Close or reject the request |
| Tokens / AST nodes per document | 1,000,000 / 2,000,000 | Return bounded diagnostic |
| Nesting / recursive query depth | 256 / 256 | Stop with a limit diagnostic |
| Declarations / semantic facts | 100,000 / 1,000,000 | Abort lowering or closure |
| Generated files / bytes per invocation | 10,000 / 256 MiB | Abort before writes |
| LSP request CPU / wall time | 2 s / 5 s | Cancel and preserve prior result |
| Child-process wall time / output | 60 s / 16 MiB | Kill, close pipes, record failure |
| Cache bytes / entries per workspace | 1 GiB / 100,000 | Evict only reconstructable entries |

Implementation requirements:

- Enforce limits while reading, tokenizing, lowering, serializing, and emitting; a
  final length check is too late. Prefer streaming decoders and bounded readers over
  unbounded `io.ReadAll` on workspace or tool output.
- Propagate `context.Context` cancellation through parser, analyzer, graph queries,
  generator, and subprocesses. Bound LSP concurrency and queue depth; coalesce or
  drop superseded diagnostics for the same document.
- Use iterative or depth-limited algorithms for attacker-controlled nesting. Bound
  transitive closure and inverse/derived relation materialization; never assume graph
  monotonicity makes a query finite.
- Reserve disk space by plan, write atomically, and enforce per-workspace quotas.
  Cache eviction must not delete authoritative source or append-only evidence.
- Add adversarial benchmark tests and run them with the race detector where shared
  state is involved. The test must assert a bounded result, not merely eventual
  completion on a healthy machine.

### T7/T8 — Identity confusion and stale projection/cache

Stable IDs are URI-like semantic identity; display names and aliases are not. The
compiler must reject a duplicate ID in different declarations unless the semantics
explicitly define one canonical owner. A relation must resolve through the namespace
and registered identity, not by matching a Go or DSL string. This prevents a benign
`Payment` in one context from being confused with another.

Every reconstructable projection and cache entry should be keyed by at least:

```text
schema/version + source digest + semantic IR digest + generator/verifier digest
              + policy digest + toolchain/dependency digest + options digest
```

On any mismatch, recompute. Cache hits are optimization evidence, not authority. A
stale semantic graph, source map, generated file, or receipt must fail freshness rather
than silently becoming the current view.

## CI security gate proposal

The existing CI checks (`gofmt`, `go vet`, unit tests, race tests, semantic
conformance, generated freshness, Go caps, changed-path scope, and
`agent/*` → `integration` branch policy) remain required. Add the following gates as
the corresponding compiler surfaces become available.

| Gate | What it proves | Required checks | Failure policy |
| --- | --- | --- | --- |
| G0 Trusted policy | Candidate cannot redefine the gate | Run verifier/security policy from protected base; protected-file diff and CODEOWNERS/required review | Block merge; no candidate override |
| G1 Filesystem containment | Reads/writes stay in authorized roots | Traversal, symlink, junction, collision, URI, atomic-write, and cleanup tests | Fail closed; preserve prior files |
| G2 Command boundary | No arbitrary execution | Static call-site audit; fixed executable allowlist; argv-only; PATH/env/hook tests; timeout/output tests | Block any shell or unapproved command |
| G3 Generator locality | Output is deterministic and non-destructive | Two-run byte equality, marker integrity, source-map digest, handwritten-slot preservation, overwrite refusal | Abort before write |
| G4 Provenance integrity | Claims are source-backed and fresh | Canonical IDs, valid spans, input/output hashes, candidate quarantine, stale/replay receipts, independent recomputation | Reject evidence and semantic delta |
| G5 Untrusted CI | PR content cannot access trust or secrets | `pull_request`, read-only permissions, no secrets/network by default, SHA-pinned actions, no privileged checkout | Cancel job or require maintainer path |
| G6 Resource budgets | Hostile inputs terminate within limits | Size/depth/fan-out/LSP/concurrency/cache/subprocess tests and bounded diagnostics | Deterministic limit error |
| G7 Supply-chain receipt | Artifact provenance is reviewable | Reproducible inputs, toolchain/dependency digest, signed or otherwise authenticated build receipt | Do not publish or merge artifact |
| G8 Role separation | No one actor closes the loop | Independent Guardian/Gate review and required checks on protected policy/ontology | Require second role/approval |

### Minimum security test corpus

Before the compiler is exposed to untrusted workspaces, add fixtures for:

- path traversal, absolute/drive/UNC paths, separator confusion, NUL, Unicode
  normalization, symlinked roots/targets, generated-name collisions, and interrupted
  atomic writes;
- shell metacharacters, PATH poisoning, workspace-local executables, invalid Git refs,
  inherited secret variables, hanging commands, descendant processes, and oversized
  stdout/stderr;
- duplicate or renamed IDs, cross-namespace references, missing/forged source spans,
  candidate facts presented as deterministic, changed input bytes with old evidence,
  mismatched generator/policy digests, and replayed receipts;
- deeply nested input, graph fan-out, large diagnostics, repeated/cancelled LSP
  requests, generated output fan-out, cache quota exhaustion, and concurrent access;
- a PR that edits `scripts/verify/**`, `internal/verify/**`, ontology files, or the
  workflow while leaving application code unchanged. The trusted-base gate must still
  detect and route this as a protected-kernel change.

Each fixture should assert the security invariant and the absence of side effects,
not just an error string. Tests that intentionally execute a child process belong in
an isolated subprocess test with no secrets and an explicit timeout.

## Threat → fixture → CI gate → rollback matrix

The following matrix turns the threat register into executable experiment contracts.
Fixture names are proposed stable IDs, not files added by this research-only change.
The eventual corpus can live below `testdata/security/` and `testdata/fuzz/`; every
fixture must carry its expected classification, budget, and side-effect oracle.

| Threat | Fixture and adversarial setup | CI gate and oracle | Rollback / containment |
| --- | --- | --- | --- |
| T1 path traversal | `path-escape`: `../`, absolute, drive/UNC, mixed separators, NUL, Unicode-normalized names, symlinked workspace/output roots, and a pre-existing outside target | G1; reject with a stable boundary error, produce no outside read/write, and leave a before/after filesystem manifest identical outside the temp root | R0 fail closed before mutation; delete only invocation-owned temp files; quarantine the workspace if an escape is observed |
| T2 arbitrary command | `command-boundary`: shell metacharacters in DSL names, flags, refs, and LSP data; PATH containing a fake tool; workspace-local executable; hanging child; oversized output | G2; process census contains only the allowlist, argv is unchanged as data, no secret appears in output, and timeout/output budgets are enforced | R3 cancel the process group and close pipes; discard temp output; invalidate any evidence from the interrupted run; block the job |
| T3 generated overwrite | `overwrite-collision`: handwritten target, forged/duplicate markers, symlink target, changed prior generated digest, and simulated failure between temp write and rename | G3; plan rejects ambiguity before writing, generated bytes are deterministic, handwritten bytes are unchanged, and the prior manifest survives failure | R1 restore the last verified projection only; never restore over authoritative DSL or handwritten source automatically; quarantine partial output |
| T4 provenance forgery | `provenance-replay`: forged source span, candidate fact marked deterministic, duplicate ID, changed source with an old receipt, mismatched generator/policy/toolchain digest | G4; verifier recomputes the claim, rejects stale or unauthenticated evidence, and records no accepted semantic delta | R2 invalidate the receipt lineage and all derived cache entries; require a fresh build from the trusted revision; preserve the rejected artifact for review |
| T5 untrusted CI/workspace | `kernel-edit-pr`: PR changes workflow, verifier, ontology, or allowed-scope policy; workspace contains hostile config/hooks; PR job attempts secret/network access | G0/G5; trusted-base gate detects protected-file changes, untrusted job has read-only/no-secret permissions, and candidate policy is never the trusted verifier | R4 cancel and block merge; discard caches/artifacts from the untrusted job; route protected-kernel changes to required maintainers |
| T6 resource exhaustion | `resource-budget`: oversized source/LSP messages, deep nesting, high graph fan-out, diagnostic flood, generated explosion, cache churn, and child with unclosed pipes | G6; each case ends within CPU/wall/memory/output/count budgets, produces bounded diagnostics, and does not starve a concurrent request | R3 cancel the request/process; preserve the last good LSP result or projection; quarantine the minimizing seed and evict only reconstructable cache data |
| T7 identity confusion | `identity-namespace`: same display name in two namespaces, alias rename, case/Unicode collision, duplicate stable ID, and relation resolved by name only | G4; IDs remain distinct, aliases do not authorize relations, and ambiguous facts stay candidates | R0 reject the delta; retain the prior IR and evidence; require an explicit source-backed identity change |
| T8 stale cache/evidence | `freshness-replay`: reuse a cache or receipt across source HEAD, policy, generator, option, or toolchain changes | G4/G7; key mismatch forces recomputation and stale materialization cannot satisfy a merge or release gate | R2 mark derived data stale, delete or quarantine it, and rebuild from current authoritative inputs |
| FZ malicious DSL fuzz | `dsl-fuzz`: grammar-preserving and grammar-breaking mutations over valid seeds, invalid bytes, truncation, huge identifiers, duplicate declarations, path-shaped IDs, and nesting/fan-out extremes | G1/G3/G4/G6; no panic, escape, unexpected process, unbounded growth, non-deterministic diagnostic, unauthorized evidence, or generated write | R3 stop the failing worker; R2 invalidate derived state; retain minimized seed and metadata; block release until the regression is fixed and replayed |

### Rollback codes and required evidence

Rollback is deliberately conservative. It must restore safety without silently
rewriting authoritative intent or deleting evidence of an attack.

| Code | Action | Evidence to retain |
| --- | --- | --- |
| R0 | Reject before a write or semantic commit; compare the filesystem manifest to the pre-run snapshot | Fixture ID, normalized input digest, rejection class, and zero-mutation result |
| R1 | Restore only the last verified generated manifest or discard the new projection; never auto-revert `.gooo` or handwritten Go | Prior output digest, attempted output digest, marker/locality result, and restore result |
| R2 | Mark receipts, caches, IR projections, and search materializations stale or quarantined; rebuild from authoritative inputs | Source/IR/policy/toolchain digests, evidence lineage, cache keys, and invalidation reason |
| R3 | Cancel the request, kill descendants where supported, close pipes, and clean invocation-owned temporaries | Process identity, deadline, output byte count, cancellation result, and resource budget |
| R4 | Stop merge/release, discard untrusted job artifacts, and require independent review or a new trusted run | Candidate/base revisions, protected-file diff, job permissions, artifact digests, and approver role |

Every rollback record is append-only. A rollback must not be represented as a new
business fact, and a failed or interrupted run must never leave a receipt that looks
like a successful verification.

## Malicious DSL fuzz experiment

### Goals and input families

The fuzz target should exercise the whole security-sensitive path in bounded stages,
not only the lexer. Start from the valid billing example and small hand-written
negative fixtures, then mutate one dimension at a time so a failure has a useful
semantic explanation.

| Family | Mutations | Properties under test |
| --- | --- | --- |
| Lexical | Invalid UTF-8, NUL, unterminated strings/comments, delimiter deletion, token duplication, extreme identifier length | No panic, bounded tokenization, deterministic diagnostics, no filesystem or process side effect |
| Structural | Deep nesting, empty/duplicate declarations, missing references, declaration/fact fan-out, truncated documents | Bounded parser/IR size, explicit rejection, no accidental implicit deletion or candidate promotion |
| Identity | Absolute/path-shaped IDs, duplicate IDs, namespace swaps, alias churn, case and Unicode normalization | Canonical identity is stable; ambiguity is rejected; names never become authorization |
| Projection | Go keywords, invalid/colliding Go names, marker text in names, output fan-out, source-map span extremes | Plan is deterministic, bounded, locality-preserving, and refuses unsafe targets before write |
| Provenance | Forged spans, old input/output digests, changed policy/toolchain, duplicate evidence, candidate marked verified | Evidence is source-bound and fresh; rejected observations cannot update IR or merge scope |
| Transport | Oversized JSON-RPC frame, fragmented message, duplicate request ID, cancellation storm, invalid document URI | LSP framing and URI roots are bounded; cancellation leaves no stale or outside result |

### Harness and oracle

The eventual Go fuzz target should use a small seed corpus through `f.Add` and retain
minimized failing inputs as versioned corpus entries. Go's official fuzzing workflow
also reruns saved failures during ordinary tests, which makes a security regression
reproducible in the regular CI path. The fuzz body should:

1. Create a fresh temporary workspace, output root, cache root, and minimal environment
   for each input. Never point a fuzz run at the repository checkout or a developer's
   home directory.
2. Snapshot the filesystem and process state before the run. Run parse/lower/analyze/
   plan in memory first; only exercise writes through the authorized test root.
3. Apply the fixed time, memory, file-count, generated-byte, graph-size, and output
   budgets. A timeout, panic, unexpected process, outside path, or unbounded allocation
   is a security failure even if no semantic assertion fails.
4. Compare the normalized diagnostic, semantic classification, filesystem manifest,
   process census, generated plan, evidence delta, and cache delta with the expected
   result. Re-run the same seed to check deterministic outcome and digest.
5. For accepted valid seeds, check semantic round-trip and generated locality. For
   rejected seeds, require zero authoritative delta, zero successful verification
   receipt, and no partial generated file.

The fuzz harness should emit only redacted metadata and a digest of the input to CI
logs. The minimized seed itself is stored as a reviewable fixture, not printed into a
log where workspace contents or injected control characters could be misinterpreted.

### Execution tiers

| Tier | Trigger and budget | Required result |
| --- | --- | --- |
| Seed replay | Every PR, ordinary `go test`; all checked-in security and minimized fuzz seeds | Zero regressions, zero side effects, deterministic classification |
| PR fuzz smoke | Every PR security job, fixed short `-fuzztime` and bounded workers | No new panic, timeout, escape, process, evidence, or budget violation |
| Nightly exploration | Scheduled isolated runner with a larger but finite time budget and fresh corpus snapshot | New findings are minimized, triaged, and either fixed or explicitly tracked; no silent corpus loss |
| Release gate | Current corpus plus targeted replay for every unresolved security finding | Zero untriaged security failures and a fresh trusted evidence record |

Fuzzing is not a license to make CI unbounded. The time budget, worker count, memory
limit, and corpus revision are part of the evidence record. A finding that only
reproduces under a larger budget remains a valid availability finding and must be
triaged rather than hidden by lowering the budget.

### Fuzz finding rollback

When a fuzz input violates an invariant, the runner should stop the affected case,
capture the seed hash and policy/toolchain digests, and run the R0–R4 sequence as
applicable: stop descendants, compare the filesystem/process manifests, discard
invocation-owned outputs, invalidate derived cache/evidence, and block the relevant
merge or release gate. The input is then minimized and added to the seed corpus with
an expected failure or fixed regression test. A finding is closed only after an
independent replay proves the property, not merely after the original run stops
crashing.

## Falsifiable implementation contracts

The security research is useful only if a future AST/IR/BX/codegen/LSP/cache/
provenance/CI implementation can run the same fixture and produce comparable
evidence. The contracts below are proposals for measurement, not measurements of
the current skeleton. An absent entry point is `deferred`; it must not be reported as
a successful implementation.

### Shared experiment record

Every fixture and fuzz finding should have one normalized record. The record may be
serialized as JSON, but its fields and meanings must remain stable across Go-hosted
and future gooo-hosted stages.

| Field | Input/output contract |
| --- | --- |
| `fixture_id` | Stable ID such as `ast/truncated-string-001`; never derive identity from a display name |
| `stage` | `ast`, `ir`, `bx`, `codegen`, `lsp`, `cache`, `provenance`, `ci`, or `fuzz` |
| `input_digest` | Digest of exact input bytes plus declared workspace/options; no normalized-only digest |
| `policy_digest` | Digest/version of limits, ontology, verifier, and security policy used |
| `expected_class` | `accept`, `reject`, `candidate`, `blocked`, or `deferred` |
| `observed_class` | Same vocabulary; unknown or missing is a failure of the harness |
| `metrics` | Wall/CPU time, peak memory, allocations, size counts, process count, path counts, and output bytes |
| `semantic_output` | Canonical IR/fingerprint or a bounded diagnostic; no implicit semantic fact from absence |
| `side_effects` | Read/write paths, generated manifest, evidence delta, cache delta, child processes, and network attempts |
| `oracle` | Deterministic predicate comparing expected and observed class, metrics, semantics, and side effects |
| `rollback_code` | One of R0–R4 when the oracle fails; empty only for a passing, side-effect-free fixture |
| `evidence_digest` | Digest of the record and referenced artifacts; never a claim that the record was independently verified |

The minimum measurement set is: `wall_ns`, `cpu_ns` when available, `peak_bytes`,
`alloc_bytes`, token/AST-node/declaration/fact counts, generated file/byte counts,
LSP queue depth, cache hit/miss, `processes_started`, `paths_read`,
`paths_written`, `paths_outside_root`, `network_attempts`, diagnostic digest,
semantic fingerprint, accepted/rejected evidence count, and rollback result. A
security pass requires `processes_started == 0` for compiler-only fixtures,
`paths_outside_root == 0`, `network_attempts == 0` unless explicitly allowed, and
all observed counts within the declared budget.

### Hypothesis and counterexample matrix

Each hypothesis is falsifiable: the listed counterexample must fail the gate even if
the returned error text looks reasonable. The fixture is the smallest proposed
reproducer, not a claim that the corresponding implementation exists today.

| Hypothesis | Minimal fixture and counterexample | Measurements | Pass / fail / deferred | Follow-up implementation contract |
| --- | --- | --- | --- | --- |
| H-AST-1: parsing is total and bounded | `ast/truncated-string-001`: valid billing prefix ending inside a quoted ID, plus invalid UTF-8 and NUL variants. Counterexample: panic, partial valid AST, or input-dependent hang | Input bytes, tokens, AST nodes, diagnostics digest, wall/peak memory, outside paths/processes | Pass if deterministic bounded rejection and zero side effects; fail on panic, hang, or accepted partial semantics; deferred if parser API is absent | `Parse(bytes, source_span, limits) -> {AST|diagnostics, source_digest, metrics}`; never execute workspace data |
| H-IR-1: canonicalization is idempotent and namespace-safe | `ir/namespace-collision-001`: two `Payment` display names with distinct stable IDs, alias rename, and duplicate-ID negative case. Counterexample: same fingerprint for distinct IDs or alias creates a new ID | Canonical node/fact counts, ID map, fingerprint before/after canonicalization, diagnostics | Pass if `canon(canon(IR)) == canon(IR)` and distinct IDs remain distinct; fail on name-based merge or silent duplicate acceptance; deferred if canonical IR is absent | `Lower(AST, policy) -> canonical IR + diagnostics + source-backed facts`; ID/namespace validation precedes derivation |
| H-BX-1: round-trip does not invent semantic facts | `bx/helper-boundary-001`: registered semantic call plus `strings.TrimSpace` helper and missing-source-span observation. Counterexample: helper lifted as a semantic relation or source-less fact accepted | Added/removed delta, candidate count, locality set, source-span validity, semantic fingerprint | Pass if DSL→IR→Go→lift preserves semantic equivalence and helper remains implementation detail; fail on spurious/unsupported promotion; deferred if analyzer/BX boundary is absent | `Lift(GoObservation) -> {deterministic delta|candidate|syntactic}`; `Reconcile` accepts only source-backed deterministic facts |
| H-CG-1: generation plans before it mutates and preserves locality | `codegen/marker-collision-001`: one changed activity, unrelated handwritten slot, forged markers, symlink target, and pre-existing non-generated file. Counterexample: partial write, marker-based overwrite, or unrelated digest change | Plan count, target manifest, generated bytes, slot hash, files/bytes written, paths outside root, two-run digest | Pass if conflict aborts before writes and valid run is deterministic/local; fail on unsafe overwrite or unrelated change; deferred if generator is absent | `Plan(IR, prior_manifest) -> plan`; `Commit(plan) -> manifest`; commit is atomic and provenance-bound |
| H-LSP-1: URI containment and cancellation are observable | `lsp/outside-uri-cancel-001`: outside-root `file:` URI, oversized request, duplicate request ID, cancellation immediately before response. Counterexample: outside read, stale response after cancel, or unbounded queue | Request bytes, queue depth, wall/CPU time, response digest, paths read, cancellation latency, diagnostics size | Pass if rejected or cancelled within budget with no outside access/stale response; fail on leaked path/result; deferred if LSP transport is absent | `Handle(ctx, workspace_roots, request) -> bounded response/diagnostic`; every read requires an authorized root and context |
| H-CACHE-1: security-relevant inputs are complete cache keys | `cache/policy-toolchain-miss-001`: one-at-a-time changes to policy, generator, toolchain, options, and source bytes. Counterexample: stale hit after any change | Key component digests, hit/miss, recompute count, artifact digest, stale-entry handling | Pass if every security-relevant change misses and recomputes; fail on stale hit or authoritative-source eviction; deferred if cache is absent | `Lookup(tuple) -> hit(artifact_digest)|miss`; tuple includes schema/source/IR/generator/policy/toolchain/options |
| H-PROV-1: evidence cannot self-authorize | `provenance/old-receipt-001`: old receipt plus changed input, forged span, candidate fact marked verified, duplicate evidence, mismatched policy digest. Counterexample: accepted claim, merge scope, or release receipt | Accepted/rejected evidence, lineage mismatch count, current input/output digests, accepted delta count, freshness result | Pass if only current source-backed evidence is accepted; fail if stale/forged/candidate evidence authorizes anything; deferred if store/verifier is absent | `VerifyClaim(claim, evidence, current_inputs) -> verdict + append-only receipt`; candidate evidence never grants authority |
| H-CI-1: trusted gate cannot be rewritten by its candidate | `ci/protected-kernel-edit-001`: PR modifies verifier, workflow, ontology, or allowed-scope policy. Counterexample: trusted job executes candidate verifier or accepts candidate policy | Protected-file diff, candidate-policy process count, token permissions, artifact/cache trust label, gate verdict | Pass if trusted base detects edit and candidate-policy execution in trusted gate is zero; fail on self-approval; deferred until trusted-base execution exists | `Gate(base, head, trusted_verifier, evidence) -> verdict`; verifier/policy/workflow identity comes from protected base |
| H-FZ-1: malicious DSL fuzz preserves stage invariants | `fuzz/dsl-seed-001`: billing seed mutated by truncation, invalid bytes, deep nesting, duplicate IDs, path-shaped IDs, fan-out, marker injection. Counterexample: panic, timeout, outside path, process, non-deterministic error, or forged fact accepted | Cases, corpus revision, crash/timeout count, p95/max wall time, peak memory, paths/processes, evidence/cache deltas, minimized-seed digest | Pass if seed replay and bounded smoke fuzz have zero untriaged violations; fail on any security side effect; deferred if fuzz target/harness is absent | Fuzz runner returns the shared experiment record, isolates temp roots, persists minimized seeds, and replays failures in ordinary CI |

The negative fixture must be tested in both directions where applicable: a malicious
input must be rejected without side effects, while a minimally repaired input must
reach the intended stage and produce the expected semantic/output fingerprint. This
prevents a gate from passing only because it rejects all input.

### Decision semantics

Use four explicit outcomes in CI and evidence:

| Outcome | Meaning | Merge/release consequence |
| --- | --- | --- |
| `pass` | The entry point ran, the oracle passed, and all required evidence is present | May satisfy only the named gate and scope |
| `fail` | The entry point ran and an invariant, budget, or oracle failed | Block; execute rollback and retain evidence |
| `blocked` | A required protected dependency or independent verifier rejected the operation | Block; do not substitute a candidate result |
| `deferred` | The entry point is not implemented or intentionally not run | Never counts as pass; only documents a scaffold baseline |

The current CLI `check` behavior belongs to `deferred`, not `pass`. A future CI
adapter must carry this outcome in its evidence rather than converting a missing
implementation into a green security result.

## Self-hosting contract and evidence

Self-hosting is a separate security milestone, not an implication of having a Go
parser or a code generator. The initial compiler is Go-hosted: Go is the trusted
implementation language for the compiler and verifier, while `.gooo` remains the
authoritative source for the semantic declarations it covers. A future gooo-hosted
stage may describe more of the compiler, IR, generator, or verification topology in
`.gooo`, but it must retain an independent trust anchor during bootstrap.

| Stage | Intended authority and implementation | Evidence required | Security claim and status |
| --- | --- | --- | --- |
| Go-hosted initial | Go implementation hosts parsing, lowering, projection, and verification; `.gooo` owns declared business intent; handwritten Go owns irreducible slots | Go source/revision, toolchain digest, DSL/IR/output digests, verifier result, round-trip/locality results, and deferred-command results | Claims only the checks that actually ran. Current repository is a Go-hosted scaffold; an unimplemented CLI check is `deferred`, not `pass` |
| Dual/transition | A trusted Go-hosted compiler produces the candidate compiler projection while a second path compares semantic IR, generated regions, and evidence | Independent Go verifier plus candidate output, normalized semantic diff, reproducible input/toolchain, and provenance path from trusted builder to candidate | Candidate cannot approve its own bootstrap; mismatches block promotion and retain the Go-hosted builder |
| gooo-hosted target | Compiler boundaries are declared in `.gooo` and lowered to a buildable projection with explicit handwritten slots; the prior trusted stage builds the next stage | `B0(source) → B1`, then `B1(source) → B2`, with source/IR/output/policy/toolchain digests and independent `B0` verification of both | Only a verified fixed point or accepted semantic equivalence may claim self-hosting; this stage is a design target, not implemented |

### Bootstrap and comparison protocol

Use explicit build labels so an output cannot claim a stronger stage than its builder:

```text
B0 = trusted Go-hosted compiler and verifier
B1 = B0 compiles the proposed gooo-hosted compiler source
B2 = B1 compiles the same source again

required: semantic(B1) == semantic(B2)
          generated/locality/evidence invariants pass for both
          B0 independently verifies B1 and B2
```

Textual byte equality is desirable for deterministic generated regions, but semantic
equivalence and source-map/provenance consistency are the minimum comparison contract.
`B1` or `B2` must not be accepted solely because the candidate compiler says that it
verified itself. A successful build without an independent `B0` verification record
is a build result, not self-hosting evidence.

The self-hosting gate should check:

- the builder identity and policy revision are explicit in every receipt;
- the bootstrap input is the same authoritative DSL/IR and is not silently replaced
  by generated output;
- generated regions and handwritten slots remain local and reproducible across B1/B2;
- semantic deltas, source spans, and evidence links are equivalent or explained by an
  approved, source-backed transition;
- `deferred`, `blocked`, and `not-applicable` outcomes are not counted as `pass`;
- failed bootstrap output, caches, and receipts are quarantined while `B0` remains
  the rollback builder.

If B1/B2 diverge, the rollback is to keep the last verified Go-hosted stage, quarantine
the candidate projections and their derived evidence, invalidate their cache keys, and
open a reproducible bootstrap fixture. Do not auto-rewrite the authoritative `.gooo`
source or mark the candidate as partially self-hosted. Promotion requires a new
trusted build and independent review of the semantic delta.

## Rollout and evidence

1. **Now:** keep this document as the security contract; retain read-only `pull_request`
   permissions and the existing deterministic checks; record the verifier
   self-modification gap as a blocking design item before security-sensitive merges.
2. **Before parser/generator release:** implement G1–G4 and G6 with malicious fixtures,
   deterministic limits, and evidence that includes source/output/policy digests.
3. **Before LSP release:** enforce URI/workspace containment, request budgets,
   cancellation, bounded diagnostics, and no execute-command capability by default.
4. **Before CI can authorize merge/release:** implement G0, G5, G7, and G8 with
   protected workflows, immutable action references, least-privilege tokens, and
   independent verification of receipts.
5. **Before a self-hosting claim:** implement the B0/B1/B2 bootstrap comparison and
   require independent B0 evidence. Until then, report the gooo-hosted stage as
   `planned` or `deferred`, never as a successful security or release gate.

An evidence record for a successful gate should identify: the protected policy
revision, candidate revision, input and output digests, generator/verifier digests,
limits used, command allowlist, test corpus version, result, and the independent
verifier identity. These records are evidence of a decision; they do not become new
business intent.

## References

- [CWE-22: Improper Limitation of a Pathname to a Restricted Directory](https://cwe.mitre.org/data/definitions/22.html)
- [Go `os/exec` package](https://pkg.go.dev/os/exec) and [Go `path/filepath` package](https://pkg.go.dev/path/filepath)
- [Go fuzzing tutorial](https://go.dev/doc/tutorial/fuzz) and [Go fuzzing security guidance](https://go.dev/doc/security/fuzz/)
- [W3C PROV-O](https://www.w3.org/TR/prov-o/)
- GitHub Actions [script injections](https://docs.github.com/en/actions/concepts/security/script-injections),
  [secure use](https://docs.github.com/en/actions/reference/security/secure-use), and
  [workflow execution protections](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/actions-policies/workflow-execution-protections)
- [SLSA security levels](https://slsa.dev/spec/v1.0/levels) and [in-toto](https://in-toto.io/)
