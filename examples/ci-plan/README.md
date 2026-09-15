# Gooo CI plan use case

This example closes one small development loop using a Gooo source program.
It turns a changed-file set into a deterministic check plan and a verification
receipt without executing the selected commands or writing to the repository.

The fixed corpus contains 12 cases: four accepted plans, four rejected inputs,
and four lower-resolution outcomes whose missing rule is named by stage, step,
reason, and file. Every case retains three claims. Evidence discharges a claim;
it does not delete the claim.

The CI scorecard presents three reader views:

* `USER`: decisions, elapsed time, peak memory, and receipt size.
* `TOOL_AUTHOR`: deterministic replays, golden plans, evidence links, and Gooo/Go source counts.
* `GOVERNOR`: persistent claims, exact unknown causality, repository writes, and mutation authority.

This slice does not claim that selected checks were executed, that the complete
Gooo language is semantically correct, or that the planner generalizes beyond
the three registered rules.
