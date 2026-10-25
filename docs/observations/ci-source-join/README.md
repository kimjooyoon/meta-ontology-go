# Original CI failure and source dependency correction

PR [#1585](https://github.com/kimjooyoon/meta-ontology-go/pull/1585) had 51
original source workflows: 50 succeeded and the core workflow failed. Its six
canonical checks and own native common execution succeeded. The proof fixture
started with about 23 seconds left and cold-compiled the CLI until the shared
40-minute deadline. The PR was closed unmerged after all original workflows
were terminal. Compressed original job data and failure logs are retained here.

The changed workflow starts source observations with the other checks, then
joins the exact original five non-policy results before final route admission.
Previously the race job took 22m11s and the subsequent policy job took 13m55s.
The independent observation work now overlaps those checks. The initial clock,
budget, all six checks and subsequent fresh reproduction remain in place.

The evidence job also publishes the exact source's warm-built execution tools
for one day. The proof job verifies source/run/attempt, platform, Go version and
tool bytes, then generates the fixture and computes the proof anew. The small
origin manifest is retained with the proof; binaries are separate short-lived
artifacts.

`observations.json` records new contract checks using mocked API rows, a minimal
temporary Go binary and shell-source syntax checks. They do not measure CI time
savings. The new original source run owns that measurement and merge eligibility.

The first source run exposed two stale regression assertions. Their
[source-contract correction](../ci-command-evolution/README.md) retains the
original command inventory and adds explicit descendant bindings.

On 25 October, dependency PR #1780 exposed a separate late rejection: its
`dependabot/*` head failed the existing feature-branch policy after a reported
1,583,566 ms work prefix. The dev-target preflight now supplies head/base to the
same verifier before metric generation, transformations and replay. The final
exact-head gate and canonical check join remain in place. This rejects work
early; it does not authorize a merge or establish CI time savings.

The command-evolution ledger retains every original policy command: twelve
bodies are unchanged, while original index 3 gains the route check. Existing
baseline and earlier workflow-origin metadata remain historical; the tests bind
all current descendants directly to the current workflow. No original command
or observation was deleted to make the new population match.
