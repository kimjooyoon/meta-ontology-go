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
| `toolchain` | This invocation's Go version process observation. A reused reference can leave it empty. |
| `toolchain_reference.source_check` | The owned reference's original version-check observation. |
| `toolchain_reference.source_check_sha256` | Digest binding the original reference key, tool digest, version, source check and byte output. |
| `toolchain_reference.reused` | Reuse of the retained version reference in this invocation. |
| `artifact.source_build` | The original build observation retained by this artifact. |
| `artifact.reused` | Reuse of that verified executable in this invocation. |

On first use, `toolchain` and `toolchain_reference.source_check` can describe the
same version process. A later request can retain that original check while its
current `toolchain` observation is empty. The first compilation can likewise
appear in both `build` and `artifact.source_build`; later artifact reuse can
retain the original build while current `build` is empty. These roles allow the
same activity on first use and preserve the absence of current work on reuse.

The two reuse flags describe separate choices. A retained toolchain reference
can support a new compilation after a source change. An original successful
check or build describes its recorded activity. Read current process flags and
the current native runs for what happened in the new request. An empty current
observation retains its null exit code; the view does not invent a process.

The source-check binding uses Go JSON encoding of the ordered tuple
`[key_sha256, go_tool_sha256, go_version, source_check, source_output]` from
`toolchain_reference`. The producer encodes `source_output` as Go `[]byte`, hence
a base64 JSON string. The typed process object's field order and optional fields
belong to that producer's contract. This is a content binding whose equality can
persist across requests; the current process observations describe occurrences.

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
