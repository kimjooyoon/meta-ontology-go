# Read runtime identities by their role

The development source adds an explicit advisory view:

```sh
go run ./cmd/gooo version --build --json --describe-runtime
```

`runtime_observation_fields` describes paths relative to a
`gooo/body-composition-runtime/v2` record. Its source is
[`fields.gooo`](../../internal/meta/runtimeidentity/fields.gooo), activity
`urn:gooo:runtime-identity:describe`. The existing Gooo-to-Go library generator
projects its records and ordered entries; Go connects them to this metadata query.
The published 0.6.39 binary predates this query modifier.

| Field | Role |
| --- | --- |
| `runtime_suite_sha256` | Content identity of the current case suite. Identical cases can retain this digest across separately executed requests. |
| `artifact.key_sha256` | Reuse identity for a compiled program and its bound source, tool and environment. Caller values remain current inputs. |
| `executable_sha256` | Identity of compiled executable bytes. Equal bytes can occur across builds or runs. |
| `artifact.source_build_sha256` | Digest binding the retained artifact key, executable digest and original build observation. |
| `build` | This invocation's build observation. Inspect `started` and `completed` before interpreting exit data. |
| `runs` | This invocation's native process observations. Retain their current start/completion/cancellation and result data. |

For example, executing a case suite and then replaying it with identical cases
can produce the same suite digest while starting new native processes. A consumer
should inspect the current run observations to determine what happened in each
request. The suite digest answers which cases were supplied.

This query describes the producer's fields. Caller-supplied IDs, a distinct
receipt digest or a repeated suite digest do not by themselves authenticate
execution. Source/build/run observations and their producer remain available
through the original receipts. The descriptive view grants no execution or merge
authority and changes no receipt or replay validation.

Descriptions are instantiated only for the explicit query above. Ordinary
version output and normal program execution keep their existing paths. There is
no new source parsing, hashing or model call during the query: its Go function
was generated at development time. No latency or memory improvement is inferred.
