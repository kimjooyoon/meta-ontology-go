# Continue development without bypassing a blocker

The current objective is to define the compiler, find defects, and repair them.
A failed operation blocks its dependent operations, not automatically the whole
language project. A merge is one possible delivery operation, not the definition
of development progress. This is an operating contract, not an implemented
automatic scheduler or proof of whole-language self-improvement.

## Scope the failure before choosing the next operation

Record the exact operation, input/source identity, observed failure, affected
capability, dependency frontier, bounded remedy, evidence needed to resume, and
next independent operation. Do not guess a cause from a job's title or infer
authentication failure from a missing token alone.

| Observed boundary | What must wait | Work that can continue |
| --- | --- | --- |
| Parser or lowering defect | Claims depending on that construct | Counterexample, scoped repair, independent constructs |
| Unsupported generator construct | Generated execution for that construct | Syntax, type checking, IR, explicit rejection tests |
| CI queue, runner or credential failure | Fresh execution evidence from that executor | Source changes, fixture design, read-only analysis |
| Guardian subject mismatch | That candidate's authorization | Conflict resolution, compiler fixes, branch CI |
| Integration conflict | Integration of the conflicting files | Bounded reconciliation, disjoint work |
| Missing required merge evidence | Protected merge | Commit, push, review, non-dependent implementation |

Independence must be justified by inputs, outputs, changed files, shared state
and authority, not simply by different task names. A new operation cannot
consume a failed result as though it passed. A shared compiler failure may block
all executions using that compiler while still permitting source-level repair.

## Evidence and authority are different

1. Define the operation and acceptance condition before running it.
2. Associate each claim with a check that actually executes its behavior.
3. Bind results to source, toolchain, fixture, workflow, run and attempt.
4. Keep failures and unsupported behavior visible. Missing evidence is UNKNOWN.
5. Apply the fix on a development branch without declaring the result verified.
6. Consume the relevant CI result; continue independent work while it runs.
7. Merge only when the repository's required protections permit it.

An unrelated optional report is not automatically a prerequisite for a compiler
fix. Conversely, a collection of green reports does not prove the compiler's
tests ran. Optional failures are classified and retained, not relabeled as
success. Required checks, approval policy, protected branches and authority
boundaries are not disabled or bypassed by this operating contract.

Whole-project suspension is reserved for a boundary that actually affects every
available safe operation, such as an explicit user pause, unknown ownership of
overlapping edits, or an integrity/authority incident covering the workspace.
Even then, unaffected read-only investigation can proceed if permitted. No
universal promise is made that every possible failure has a safe workaround.

## Concrete compiler case: PR 891

Observed parent: `ab89341cfb6790b9e64aac37fa04491750b7b435`.
Integration base: `372d8d5591717ba570dfad21894fa0428b32571d` (`dev`).
The merge exposed six conflicting files and two representations of bindings.
The remedy is to retain the canonical `BindingDecl` / `RuntimeBinding` path and
make `CompileTypedPlan` consume `Get(document)`, not maintain a second parser,
identity resolver or port-type checker. Endpoint evidence is retained in the
canonical model; the existing IR carries the whole binding span.

Targeted acceptance is defined in executable tests:

- `TestTypedPlanUsesCanonicalIdentitiesAndStableOrder`: source-defined chain,
  canonical identities and order independent of binding declaration order.
- `TestTypedPlanRejectsInvalidBindingDocuments`: eight rejection cases for
  missing edges, unknown activity, wrong port, duplicate, cycle, conflicting
  input, type mismatch and repeated input. Failures return no partial plan.
- `TestTypedPlanRetainsEndpointProvenanceThroughCanonicalModel`: endpoint source
  coordinates and agreement with lowered IR on identity, type, ports and span.
- `TestTypedPlanFailureDoesNotBlockIndependentDocument`: one invalid document
  does not change the next compilation of a valid independent document.

These are acceptance definitions, not pass counts. Actions test events determine
which cases executed and passed. Existing v26.2 grant-fixture changes inherited
by this branch are preserved; these compiler checks do not validate those
changes or grant them authorization.

The previous domain workflow attempted generation despite the generator's
explicit unsupported-binding boundary. The corrected workflow tests the
compiler and records this limitation rather than deleting bindings to generate
a different program or claiming deterministic generation without an artifact.
Its branch-push route can collect compiler evidence without waiting for merge
authorization; it cannot authorize a merge itself.

Guardian run `35458233471`, artifact `10589610367`, reported
`CI-FOUNDATION-AUTHORIZATION-001`: "Guardian dispatch live candidate tuple is
not exact". Artifact archive digest:
`sha256:3999c35c56aa29b47eaa1bf06b0cdcd52f607ceada7bbdb4a1146e704741a088`.
This records an authorization observation, not a diagnosed local credential
failure. No secret, protection or authority change is part of this remedy.

## Completion and recurrence

Keep each failure's original evidence and append the remedy commit, new run,
affected test outcomes, merge result (if any), and recurrence when observed.
Do not overwrite the old failure with the new state. A tested compiler fix, a
merged change, a successful execution and external usefulness are separate
claims. CI runs, commits and declarations are not improvement scores.

Local Go tests, builds, generators, fixes and formatters remain prohibited for
this work. Execution evidence is produced by GitHub Actions. Missing matching
before/after observations keep improvement UNKNOWN.
