# Live governance snapshot

This observation records the public GitHub branch-protection and repository-ruleset
envelope. It is a semantic change-control receipt, not a settings writer, health
score, or privileged branch-policy claim.

The source authority is the GitHub REST API. The pinned branch-protection
documentation uses API version `2022-11-28`; the four public repository/branch/
ruleset requests use `2026-03-10`.
The contract pins the public documentation URLs and four exact public requests. A
single capture is replayed locally in the evaluator; replay is not a second
network fetch.

The contract has 12 cells and 12 Gooo activities with FOUNDATION, COHERENCE, and
REGRESSION proof routes of 4/4/4. A current dev enforcement/context mismatch is
an exact REFUTED observation: it does not authorize settings changes or promotion.
Missing or unavailable public evidence is UNKNOWN with its complete causal
frontier. Disabled rulesets are observed but are not treated as active authority.

The branch summary payload is the sole branch-protection authority in this
contract. The privileged `/branches/{branch}/protection` endpoint is not
requested; fields outside the public branch envelope are not claimed. Payload
digests use the explicit `canonical-json-v1` representation.

The public endpoint contract is documented by:

- https://docs.github.com/en/rest/branches/branch-protection?apiVersion=2022-11-28
- https://docs.github.com/en/rest/branches/branches?apiVersion=latest
- https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets
