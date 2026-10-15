# Source editing and dependency review — 2026-10-15

This review started with 13 open pull requests, no open Dependabot vulnerability
alerts and seven open CodeQL alerts on `dev`. Alert severity is the scanner's
classification; this review has not demonstrated remote exploitation.

## Compiler changes

| Alerts | Root cause | Change | Preserved meaning |
| --- | --- | --- | --- |
| 12, 13 | Assembly integers were converted even after a range diagnostic. | Convert only inside the valid range branch; duplicate attempts retain the first value. | Attempts remain 1–64; occurrence remains 0–127. |
| 14 | Measurement constructed a quoted source fragment and replaced the first matching text. | Parse the selected activity, bind its source digest and edit only its literal span with `strconv.Quote`. | The empty-program baseline and changed-program fingerprint remain distinct observations. |
| 15 | Replacement capacity used unchecked integer addition. | Share a span-checked literal editor with an explicit addressable-size check. | All bytes outside the selected literal remain unchanged. |
| 21–23 | The interpreter narrowed int64 storage through byte, int32 and native int casts. | Express the low bits and signed interpretation directly in int64 storage. | Go's typed integer wrap semantics remain intact, including negative values. |

The source revision proposal path uses the same literal editor. This addresses
the adjacent capacity calculation as well as the reported body-fill site.
Measurement edits no longer change an earlier identifier or another activity
whose text happens to contain the same program. Escaped and raw literals are
handled through their parsed value and exact source span.

The new checks cover 66 integer-width comparisons, invalid and boundary parser
values, capacity arithmetic without huge allocations, literal escaping,
unchanged surrounding bytes, repeated text and stale source identities. One
new parser fixture initially omitted its required value case; the fixture was
repaired and only that failed check was repeated. The first result is retained.

## Dependency integration

| Source PR | Selected version | Integration issue |
| --- | --- | --- |
| #1592 | `actions/cache` v6 | Update both restore and save, and the workflow contract that still required v5. |
| #1593 | `github/codeql-action/upload-sarif` v4.38.3 | Use the published commit and correct its stale version comment. Its original failures included time-limited commands and downstream missing artifacts. |
| #1594 | `actions/create-github-app-token` v3.3.0 | Update the exact trusted action reference asserted by the promotion protocol test. |
| #1595 | `actions/setup-go` v7.0.0 | Bind workflows to the new availability observation; preserve the earlier v6.4.0 observation. |

The four published action definitions use Node 24. The v7 setup-go archive
previously returned HTTP 404 in a retained observation; the new direct fetch
returned HTTP 200. This is an observation of availability at two times, not a
claim about the cause of the earlier failure. Metadata, commit identities and
digests are in [the action origin record](observations/action-origin-upgrade/observations.json).

The dependency changes are integrated together so their source bindings and
consumers can be checked on one revision. Existing checks and branch rules keep
their roles. CI results, the merge and subsequent scanner results must be read
separately before calling the remediation complete.

## Sources

- [CodeQL: incorrect integer conversion](https://codeql.github.com/codeql-query-help/go/go-incorrect-integer-conversion/)
- [CodeQL: unsafe quoting](https://codeql.github.com/codeql-query-help/go/go-unsafe-quoting/)
- [CodeQL: allocation size overflow](https://codeql.github.com/codeql-query-help/go/go-allocation-size-overflow/)
- [setup-go v7.0.0](https://github.com/actions/setup-go/releases/tag/v7.0.0)
- [cache v6](https://github.com/actions/cache/tree/55cc8345863c7cc4c66a329aec7e433d2d1c52a9)
- [create-github-app-token v3.3.0](https://github.com/actions/create-github-app-token/releases/tag/v3.3.0)
- [CodeQL Action v4.38.3](https://github.com/github/codeql-action/releases/tag/v4.38.3)
