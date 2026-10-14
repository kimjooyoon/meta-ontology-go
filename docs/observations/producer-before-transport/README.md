# Source producers complete before artifact transport

This full successor retains PR #1572's entire language change. The ordinary
[variables/condition example](../../../examples/everyday-program/README.md)
and first-program guide make existing source behavior easier to try. Its new
finite rows and origin report retain their actual development producer identity.

## Own original observations

PR #1572's [39 original workflows](predecessor-terminal-originals.json.gz)
finished with 38 successes and one failure. Its core workflow and canonical
checks passed. It remains unaccepted because one source-selected workflow failed.

The [race log](race-job.log.gz), [job](race-job.json) and
[event reconstruction](race-event-observation.json) bind run `38614678813`,
job `116026606552`, attempt 1, to source
`10067b26de6d3fa75040173a610fdc522b5da62a`. The unchanged 20-minute job completed
successfully in 12 minutes 56 seconds. It recorded 244 package PASS events,
187 no-test package SKIPs, 8,450 test/subtest PASS events and 17 test SKIPs.
No FAIL or unfinished package/test event remained. The extractor started at
`18:29:11.844045009Z` and finished at `18:37:22.256649849Z`.
These are nested events from one original invocation, not independent feature
counts or a controlled performance comparison.

The public-continuity [original log](continuity-job.log.gz),
[job steps](continuity-job.json) and [annotations](continuity-annotations.json)
bind run `38614678789`, job `116026608235`, attempt 1, to that same source.
The graph-driven partial-reuse producer and contradiction checks succeeded.
Its artifact upload then failed with `ENOTFOUND`. The later semantic-resolution
producer was skipped, and the final upload found no output from that skipped
producer. The DNS failure does not establish a compiler defect or semantic repair
counterexample; the skipped operation still leaves the workflow incomplete.

## Source and transport order

All eight declared Gooo producer/verification stages now precede the five
artifact upload steps. The [step-block comparison](workflow-order-observation.json)
retains all 16 original step blocks byte for byte, including setup, source
commands, contradiction checks, artifact names/paths, pinned action, 30-day
retention and `if-no-files-found: error`. Uploads still run with `always()`.
A transport failure still fails the workflow. There is no masked failure, retry
of an original workflow, relaxed test scope or changed authority.

Moving publication after source production lets all source observations finish
before transport begins. It does not repair GitHub's DNS availability. The next
original workflow must demonstrate both source completion and successful uploads.

## New source observations

- [order65243](order65243/manifest.json): two new source-order contract tests
  passed once. Missing, repeated/late and prematurely uploaded producers are
  rejected by the contract. No stored Gooo protocol case was locally executed.
- [static65581](static65581/ended.json): all 81 carried Go files; zero fix diff
  and zero diagnostics; no tests executed.
- The ordinary example records three new native rows, all passed, and a resolved
  16-node source graph with zero model calls. Its temporary compiler was removed.

Acceptance, merge and the prepared 0.6.35-dev release require the successor's own
source-selected original family, exact six checks, source-bound proof, native
readiness and CodeQL. Predecessor successes are observations of their own source.
