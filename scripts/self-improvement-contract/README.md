# Self-Improvement Contract Compiler

This metaprogram compiles the Gooo self-improvement declaration through the
repository's canonical syntax and bidirectional semantic lowering pipeline.

It compares the resulting semantic model with the live generation registry.
No executor list or proof route is inferred from filenames outside that
registry.

The Gooo declaration now authors the existing activity IDs explicitly. Contract
stages and executor inputs are resolved from the lowered IDs, `prov:used`, and
`prov:wasGeneratedBy` facts. Display names can change, including to Korean, while
the same IDs and relationships retain the same semantic hash. The source hash
still changes. Input order remains part of the contract through the IR's source
input sequence; the unordered PROV relation does not erase it.

Each covered executor includes the expected activity ID, current activity/entity
names, and the source span of its admitted input fact. These spans refer to the
source bytes bound by the report's source hash. They describe the observed
contract and do not authorize edits. A missing or changed identity, wrong input
or result, or use of the isolated observation remains a failed contract.

The report separates eight indicators across foundation, coherence, and
regression routes. It binds the exact commit and Gooo source digest, requires
a valid semantic IR, verifies the closed loop and Munchhausen choice, checks
every registered executor, verifies the isolated read-only language
observation activity, and replays the projection deterministically.

The tool only observes and reports. It never authorizes promotion or edits
the source contract.
