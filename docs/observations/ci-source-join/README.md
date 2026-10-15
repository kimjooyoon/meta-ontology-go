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
