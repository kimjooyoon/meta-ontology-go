# Meta execution cost report

This diagnostic consumer reads plain driver-boundary NDJSON from standard input
and emits a JSON interval report. It does not execute compiler operations, write
repositories, change CI gates or authorize candidate application.

In a CI workspace, after extracting the JSON driver events without modifying their
contents:

```sh
go run ./scripts/meta-cost-report < driver-events.ndjson > cost-report.json
go run ./scripts/meta-cost-report -format markdown < driver-events.ndjson > cost-report.md
```

The consumer accepts existing v1 driver events without cost fields as unmeasured.
The proposed timing extension in PR #751 provides measured intervals. No sibling
branch or repository is required to build this consumer.

Each interval retains the invocation, source, plan, manifest, Gooo activity,
indicator, subject, input-contract and pass bindings. Duplicate event identities,
reused starts, mismatched bindings, invalid durations and unknown cost decisions
are rejected. Unfinished starts and unknown returns remain explicit counts.

There is intentionally no additive grand total: an action interval contains its
process intervals. Source authenticity is UNVERIFIED; a parsed log is not a signed
CI receipt. Improvement remains UNKNOWN. This parser does not establish toolchain
equivalence, cache misses, semantic conformance or external utility. It is not a
replacement for the existing canonical receipts or independent replay.

The Markdown view exposes measured, unmeasured and incomplete counts separately,
with one row per interval. It deliberately omits a grand total and any claimed
speedup. JSON remains the detailed view with full input bindings. Input labels
are escaped before inclusion in Markdown tables, and output errors are fatal.

The first implementation is a standalone consumer, not yet wired into automatic
artifact publication. Native CI must validate it before merge. Follow-up work
must exercise the consumer on actual #751 execution logs and publish a readable
CI summary without promoting these diagnostic intervals into authority evidence.
