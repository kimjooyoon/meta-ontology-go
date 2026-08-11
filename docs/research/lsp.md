# LSP research and conformance plan

This note records the LSP boundary for `.gooo` and a testable path from the current
minimal server to an editor-quality implementation. It is deliberately a research
document: it does not change `internal/lsp` or the protocol surface. The reference
specification is the current [LSP 3.18 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/);
the implementation should advertise only the capabilities it actually implements.

The current implementation baseline is commit `25e7974` on `agent/lsp`. It has
Content-Length framing, a parser seam, document open/change/close handling, parse
diagnostics, and minimal same-document completion, definition, and hover. It has no
semantic-token or workspace-edit implementation, and it does not yet connect Go AST
or type information to registered semantic identities. The observations below are
intended to guide that implementation branch without overlapping it.

## Project boundary

The `.gooo` DSL is authoritative for business meaning and stable semantic IDs. The
LSP is a projection and interaction boundary over the DSL, semantic IR, generated Go,
and handwritten slots described in [the architecture](../architecture.md) and [the
language sketch](../spec.md). That implies four constraints:

1. A position, range, symbol, diagnostic, token, or edit must be traceable to a
   document snapshot and source span.
2. Display-name lookup is only a convenience. Cross-file definition, hover, rename,
   and Go lifting must resolve through namespace-safe semantic identity.
3. LSP capabilities are a contract, not a wish list. A server must not advertise
   incremental sync, semantic tokens, or workspace edits until the corresponding
   request and state behavior is complete.
4. Results computed for an older snapshot must not overwrite newer diagnostics,
   tokens, or edits. A request should carry the document version/snapshot it read and
   be discarded or reported as stale before publication.

## Lifecycle and transport

### Wire contract

LSP uses JSON-RPC 2.0 over a byte-framed stream. `Content-Length` is required and is
the byte length of the UTF-8 JSON content; headers and content are separated by
`\r\n\r\n`. A processed request always receives a response. A processed
notification never receives a response, including an error response. Unknown
notifications may be ignored, while an unknown request returns `-32601`.

The current framing code handles partial reads, short writes, malformed headers, and
a maximum message size. The conformance gap to close is the full request/notification
distinction: several request handlers currently build a response even when the
incoming message has no ID. Add a message-kind check before dispatching a response.

### State machine

```text
created -- initialize --> initializing -- initialize response --> running
running -- shutdown --> shutdown-requested -- exit --> exited
created -- exit --> exited-with-error
```

The [initialize rules](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/general/initialize.md)
are normative for ordering:

- `initialize` is the first request and may be sent only once.
- Before initialization, requests fail with `-32002` (`ServerNotInitialized`);
  notifications are dropped except `exit`.
- The client sends no further requests or notifications until it receives the
  `InitializeResult`, then sends the `initialized` notification.
- The server must not send ordinary requests or notifications before the initialize
  response. Capabilities and negotiated settings belong in this handshake.
- `shutdown` is a request whose successful result is `null`. The server waits for
  `exit`; exiting without a prior shutdown is an error exit.

`Server` should make these states explicit rather than using an `initialized` boolean
that is only written. Requests after shutdown should be rejected except for `exit`.
`$/cancelRequest` must identify a pending request and cause it to return a response;
it must never leave a request hanging. A `context.Context` should flow from dispatch
into parse, IR, Go analysis, and publication so cancellation cannot publish stale
work.

The initialize params should retain at least `rootUri` (or `workspaceFolders`),
client capabilities, and initialization options. The response should include the
server-selected `positionEncoding`, text synchronization mode, and only the feature
providers that are live. Keep unknown client capability fields forward-compatible.

## Document synchronization and UTF-16 positions

### Snapshot model

Treat each open URI as a versioned document snapshot:

```text
didOpen(uri, version, text)  -> snapshot S1 -> parse/lower -> publish
didChange(uri, version, edits) -> apply to Sn -> snapshot Sn+1 -> revalidate
didClose(uri) -> release open ownership and clear file-owned diagnostics
```

The [text synchronization rules](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/specification.md#textDocument_synchronization)
require the server to implement open, change, and close together or none of them.
`TextDocumentSyncKind.Full` (`1`) means every change contains the complete document;
`Incremental` (`2`) means the initial full text is followed by range edits. The
current baseline advertises `Change: 1` but applies range edits. Choose one of these
consistent contracts; incremental sync is the better fit for a long-lived editor,
provided the range applier is made strict and version-aware.

For incremental sync:

- apply changes in the array order supplied by the client;
- reject a range outside the current snapshot instead of silently slicing bytes;
- validate the version transition according to the server's chosen policy;
- keep an immutable text/symbol/IR snapshot for each request;
- use the open buffer as the source of truth while it is open, and disk only for
  unopened workspace documents.

### Position boundary

LSP positions are zero-based. Since LSP 3.17, the client and server can negotiate an
encoding; if the client omits the list, UTF-16 is the default and remains mandatory.
The [Position definition](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/position.md)
requires positions beyond the line length to clamp to the line end. For the initial
`.gooo` server, select and advertise `utf-16` explicitly, while storing internal
offsets as byte offsets into the exact UTF-8 snapshot.

Conversion rules:

- `Position.character` counts UTF-16 code units, not bytes, Unicode scalar values,
  grapheme clusters, or display columns.
- Convert only at the LSP boundary. Lexer/parser/IR spans should remain byte ranges.
- Never cut through a UTF-8 sequence. If a client sends a position inside a surrogate
  pair or otherwise between encoded units, clamp to a safe byte boundary.
- Use the same converter for diagnostics, hover ranges, definitions, completion text
  edits, semantic-token `deltaStart`/`length`, and workspace edits.
- Define and test the line-ending policy. `\r\n` must not create a phantom character
  that makes a returned range fail when the editor applies it.

The existing `positionOffset`/`offsetPosition` functions already count UTF-16 units
for non-BMP runes. They need boundary tests and a single shared source of truth for
all future features. A useful fixture is `a😀b`: byte offsets and UTF-16 columns are
different, and the emoji occupies two UTF-16 units.

## Diagnostics

`textDocument/publishDiagnostics` is a server notification. Diagnostics are owned by
the server, and each publication replaces the previous set for that URI. An empty
array is therefore required to clear old errors. See the [publish diagnostics
contract](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/publishDiagnostics.md).

The `.gooo` diagnostic pipeline should preserve the following order and provenance:

```text
snapshot -> lex/parse diagnostics -> lower/IR diagnostics
         -> registered Go projection diagnostics -> publish for that snapshot
```

Every diagnostic should have a stable code, `source: "gooo"`, a UTF-16 range, and a
message safe to show without internal stack details. Recommended categories are
`lex.*`, `parse.*`, `semantic.*`, `projection.*`, and `go.*`. The diagnostic data
should retain the semantic ID and source span internally even if the client receives
only the standard LSP fields. Add `version` to the publication when the client
advertises `versionSupport`.

The current baseline publishes parse diagnostics on open and change and an empty set
on close. It does not yet publish semantic/Go diagnostics or associate a publication
with a version. Future work must avoid publishing a result after a newer change has
won the document race; either drop it or return `ContentModified` for a request that
supports that error.

## Completion, definition, and hover

These are request/response features over the synchronized snapshot. A missing symbol
or an unsupported location returns `null`/an empty result, not a fabricated location.

| Feature | LSP contract | `.gooo` behavior to target | Current baseline |
| --- | --- | --- | --- |
| Completion | `textDocument/completion`; client may filter/sort; use `textEdit` when the server owns replacement range; `isIncomplete` controls refetch behavior. [Spec](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/completion.md) | Offer grammar keywords, namespace-safe entities/activities, and registered Go symbols only when the context permits them. Include stable IDs in detail/documentation and keep ordering deterministic. | Static keywords plus current-document symbols; no context or text edits. |
| Definition | `textDocument/definition`; result may be `Location`, `Location[]`, or `LocationLink[]`. [Spec](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/definition.md) | Resolve references through semantic identity, then return DSL declaration or registered Go projection location. Prefer `LocationLink` only when the client advertises support. | Same-document name lookup; no workspace index or identity-aware cross-file resolution. |
| Hover | `textDocument/hover`; result is `Hover` or `null`, with client-negotiated markup formats. [Spec](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/hover.md) | Show declaration kind, stable semantic ID, inputs/outputs, provenance links, and Go projection/source-map information without treating a display name as identity. | Plaintext symbol detail and a local range. |

Completion, definition, and hover should share one `SymbolAt(snapshot, position)`
resolver. It should return the AST node, semantic identity, source range, and
authority (DSL, IR, generated Go, or handwritten Go). This prevents each feature from
inventing a different name-matching rule.

## Semantic tokens

Semantic tokens are not syntax highlighting text; they are a compact semantic view
that the client renders. The [semantic token specification](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/semanticTokens.md)
requires a legend and a relative integer stream. Each token is five integers:

```text
[deltaLine, deltaStart, length, tokenTypeIndex, tokenModifierBits]
```

`deltaStart` is relative to the previous token on the same line and resets when the
line changes. `length` and `deltaStart` use the negotiated position encoding. The
array length must be a multiple of five, tokens must be ordered, and overlap/multiline
behavior must respect the client capabilities.

Recommended initial legend:

- token types: `keyword`, `namespace`, `type`, `function`, `parameter`, `string`,
  `operator`, `comment`;
- modifiers: `declaration`, `definition`, `readonly`;
- use standard types where possible and add domain-specific types only when the
  semantic distinction is useful to a client.

Token classification should come from lexer spans plus the lowered symbol table, not
from capitalization or string coincidence. In particular, an entity declaration and
an entity reference should share identity while using `declaration`/`definition`
modifiers as appropriate. A semantic ID string is a `string` token; its value must not
be rewritten by highlighting.

Start with `textDocument/semanticTokens/full` and no delta support. Add a stable
`resultId` and `full.delta` only after snapshot caching and invalidation are proven.
Range tokens can follow full tokens. A full response has the shape `{ data: [] }`; a
delta response must transform exactly the previous result named by `previousResultId`.

## Workspace edits

`WorkspaceEdit` represents changes across resources. The [workspace edit contract](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/workspaceEdit.md)
allows either `changes` or `documentChanges`. Prefer `documentChanges` when the client
supports versioned edits, because each `TextDocumentEdit` can name the version that
was read. Resource operations are ordered and must be applied in order.

The first useful `.gooo` edit is semantic rename:

1. `textDocument/prepareRename` validates that the position denotes a renameable DSL
   name and returns its exact range/placeholder.
2. `textDocument/rename` resolves the stable semantic ID and computes non-overlapping
   edits for the declaration and all identity-resolved references across open and
   unopened documents.
3. The edit changes a display name or alias, never the stable `id "..."` value.
4. Generated Go is regenerated through its projection boundary; generated regions are
   not edited as an unrelated text file.
5. If a client cannot accept versioned `documentChanges`, fall back to `changes` only
   when the edits are safe for the client's capability set.

For server-initiated code actions or projection updates, use `workspace/applyEdit`
only when the client advertised `workspace.applyEdit`. Include a change annotation
such as “rename semantic ID display name” when the client supports annotations. A
workspace edit must be rejected or recomputed if any document version has changed;
silently applying a stale edit violates locality and can corrupt a user's buffer.

## Go tools integration

The repository's standard-library-only constraint is compatible with a narrow,
deterministic Go projection analyzer. It is not equivalent to embedding `gopls`.
`gopls` is the official Go language server and a useful behavior reference; its
documentation says that it invokes the `go` command to learn workspace information
and supports module, multi-module, and GOPATH layouts. Use it as a compatibility
oracle, not as a dependency of `internal/lsp`.

### Proposed adapter

```text
LSP snapshot + workspace root
    -> go list -json (build context/package files)
    -> go/parser + go/ast + go/token.FileSet
    -> go/types (Defs, Uses, Types, Selections, Scopes)
    -> registered semantic-symbol filter
    -> semantic delta + Go locations/diagnostics
```

1. **Discover the build view.** Run `go list -json` with the workspace root as the
   working directory and the relevant build flags. The official [go list
   documentation](https://pkg.go.dev/cmd/go#hdr-List_packages) defines the JSON
   package metadata and file sets. Bound the process with the request context, retain
   stderr as evidence, and key the result by root, environment, flags, and module
   files.
2. **Parse exact files.** Use [`go/parser.ParseFile`](https://pkg.go.dev/go/parser),
   [`go/ast`](https://pkg.go.dev/go/ast), and [`go/token.FileSet`](https://pkg.go.dev/go/token)
   with comments enabled. Go token columns are byte-based; convert `token.Pos` to a
   byte offset first, then use the shared UTF-16 converter for LSP ranges.
3. **Type-check deliberately.** Use [`go/types.Config.Check`](https://pkg.go.dev/go/types)
   and populate `types.Info.Defs`, `Uses`, `Types`, `Selections`, and `Scopes` for
   definition/hover/lifting. `go/importer` itself warns that its simple importers are
   not reliable for module-aware loading and points to the external
   `golang.org/x/tools/go/packages` loader. Because this project is standard-library
   only, start with workspace-local and standard-library cases, or implement and test
   a small importer backed by `go list` before considering a dependency exception.
4. **Filter semantic facts.** Only lift calls, references, and generated symbols that
   carry a registered semantic ID or an unambiguous source-map entry. Calls to
   helpers such as `strings.TrimSpace` remain implementation details, matching the
   DSL's bidirectional boundary. A type-check success alone is not a business fact.
5. **Preserve locality and evidence.** Map Go objects back to semantic IDs, source
   spans, generated-region markers, and the current snapshot hash. Emit candidates
   for ambiguous facts; promote only source-backed or explicit DSL assertions. Cache
   parsed/type-checked artifacts by content and build context, never by display name.

This adapter should be introduced behind an interface so the LSP transport does not
depend on the Go loader or make network/filesystem calls for every hover. The Go tool
process and all filesystem reads must be cancellable, bounded, and reproducible in CI.

## Hosting stages and self-hosting contract

Self-hosting is a staged target, not an assumption about the current implementation.
The initial server is Go-hosted: Go owns the process, transport, parser adapter,
snapshot scheduler, and projection interfaces, while `.gooo` owns business meaning.
The future gooo-hosted stage may express more of the compiler/LSP topology in `.gooo`
and regenerate its Go host, but it must preserve the same protocol and semantic
contracts. The future stage is not implemented and must not be reported as passing
until the evidence below exists.

| Dimension | Go-hosted initial stage | gooo-hosted future stage | Required comparable evidence |
| --- | --- | --- | --- |
| Authority | `.gooo` declarations and stable IDs; handwritten Go owns irreducible logic | Same authorities; `.gooo` may also describe compiler/LSP topology | Identity and authority classification is identical on the fixture corpus. |
| Runtime host | `cmd/gooo` and `internal/lsp` compiled by Go | A verified `.gooo` projection bootstraps the same Go runtime or a later host | Framed protocol traces, diagnostics, tokens, locations, and edits are semantically equivalent. |
| Bootstrap | Handwritten/generated Go is the trusted execution seed | Previous verified compiler builds the next host; a small Go seed remains for recovery | Two independent bootstrap paths produce matching normalized IR and generated-region maps. |
| Navigation | Go source maps connect generated regions to `.gooo` | Self-described source maps add compiler/LSP declarations as targets | `LocationLink` targets, stable IDs, and locality checks match across hosts. |
| Verification | Go unit/vet/race tests plus conformance fixtures | The same tests plus self-build, reproducibility, and bootstrap-delta evidence | No policy or verifier is weakened by the new host. |
| Failure handling | Fall back to the Go host when a projection is unavailable | Fall back to the last verified host; do not claim self-hosting from a partial build | A failed bootstrap is observable, bounded, and never promoted to a successful artifact. |

The transition gate should compare hosts on the same input snapshot, client
capabilities, toolchain metadata, and fixture manifest. Required evidence includes:

- semantic IR equivalence, not textual equality, for parse, lower, and lifted facts;
- identical diagnostic sets, result-ID behavior, UTF-16 ranges, `LocationLink`s, token
  streams, and workspace edits after canonical ordering;
- generated-region and source-map locality, including an unchanged handwritten slot;
- reproducible generated output and a recorded provenance chain for the bootstrap
  compiler, host binary, test run, and evidence artifacts; and
- a negative test proving that a future-host failure remains “not implemented” or
  “not verified,” rather than being mislabeled as a passing self-hosted stage.

Until those comparisons pass, the status of gooo-hosted operation is **planned**, and
the Go-hosted server remains the only implementation claim in this research note.

## `internal/lsp` conformance checklist

The checklist is intentionally explicit so a future guardian can verify behavior
without changing the feature. “Baseline” refers to `agent/lsp` commit `25e7974`.

| Area | Conformance item | Baseline | Exit criterion |
| --- | --- | --- | --- |
| Transport | Content-Length counts UTF-8 bytes; partial frames and short writes work | Partial | Framing tests include non-ASCII payloads, multiple frames, `Content-Type`, and EOF/error cases. |
| Transport | Requests receive one response; notifications receive none | Gap | A notification with invalid params produces no response; every request ID is echoed exactly. |
| Lifecycle | Pre-initialize requests return `-32002`; pre-initialize notifications are dropped | Gap | State-machine test covers initialize-before-use and exit exception. |
| Lifecycle | Initialize is once-only; `initialized` gates normal work | Gap | Duplicate/late initialize and early document requests have deterministic outcomes. |
| Lifecycle | Shutdown requires a following exit; exit-before-shutdown is non-zero | Partial | Tests cover requests after shutdown and clean stream/process termination. |
| Lifecycle | Cancellation returns a response and stops cancellable work | Gap | `$/cancelRequest` is correlated by ID and no stale publication occurs. |
| Sync | Open/change/close are implemented as one capability | Partial | Capability and handlers agree; version and URI ownership are tested. |
| Sync | Full vs incremental change declaration matches the applier | Gap | Full and incremental fixtures pass their respective capability contract. |
| Position | UTF-16 conversion is shared and clamps safely | Core exists | ASCII, BMP, non-BMP, CRLF, EOF, and invalid-boundary cases round-trip. |
| Diagnostics | New publication replaces old; empty publication clears | Partial | Versioned snapshot tests prove stale results cannot win. |
| Diagnostics | Stable codes and source spans survive parse/IR/Go phases | Parse only | A fixture covers lex, parse, semantic, and projection diagnostics. |
| Completion | Capability, context, filtering, replacement range, and stable ordering agree | Minimal | Prefix and grammar-context tests assert exact items and edits. |
| Definition | Result resolves identity across the workspace | Same-file names | Namespace collision, cross-file, generated, and missing-definition tests pass. |
| Hover | Markup negotiation and semantic/provenance detail are stable | Plaintext local | Client format negotiation and snapshot consistency are tested. |
| Semantic tokens | Legend, full response, relative encoding, UTF-16 lengths | Absent | Golden decode/re-encode tests pass; data length is always a multiple of five. |
| Workspace edit | Versioned, non-overlapping edits preserve IDs and generated locality | Absent | Rename/apply tests cover open buffers, disk files, and stale versions. |
| Go bridge | `go list`, AST, types, source maps, and cancellation are bounded | Parser seam only | A module fixture proves registered calls lift and ordinary helpers do not. |
| Concurrency | Requests use immutable snapshots and race-free caches | Synchronous baseline | Parallel feature requests and `go test -race` show no data races or stale writes. |

## Proposed future tests

These tests should be added to the implementation branch in small, reviewable groups.
They are listed here so the verification evidence can later be tied to a specific
protocol invariant.

| ID | Scenario | Required assertion |
| --- | --- | --- |
| LSP-T01 | Read/write fragmented frames containing emoji and two back-to-back messages | Payload length is byte-based; exactly two messages are decoded; no frame consumes the next one. |
| LSP-T02 | Request/notification matrix before and after initialization | Pre-init request is `-32002`; pre-init notification is dropped; processed notifications never get a response. |
| LSP-T03 | Duplicate initialize, initialized ordering, shutdown, exit | Only the legal lifecycle sequence succeeds; exit-before-shutdown returns the documented non-zero result. |
| LSP-T04 | Cancel a slow parse/type request, then change the document | Cancellation returns; no result or diagnostic from the old snapshot is published after the new one. |
| LSP-T05 | Apply multiple full and range changes with versions 1, 2, 3 | The capability matches the payload shape; edits apply in order; invalid/out-of-order versions are deterministic. |
| LSP-T06 | `a😀b`, Korean text, combining marks, CRLF, and cursor at EOF | Byte offsets map to UTF-16 positions and back without splitting UTF-8 or misplacing a range. |
| LSP-T07 | Syntax error fixed on the next change | The error publication is replaced, and an empty diagnostic array clears the client state. |
| LSP-T08 | Completion inside each header/declaration/reference context | Only valid items are returned, with deterministic ordering and exact replacement `TextEdit`s. |
| LSP-T09 | Same display name in two namespaces | Definition/hover resolve the requested stable ID; string coincidence never crosses namespaces. |
| LSP-T10 | Full semantic token response for a multilingual `.gooo` fixture | Decoding the five-integer stream yields sorted, non-overlapping tokens with the advertised legend. |
| LSP-T11 | Edit one declaration and request rename | Workspace edit changes names/references but never the stable ID or unrelated generated region. |
| LSP-T12 | Rename against a changed document version | The edit is rejected or recomputed; no stale text is applied. |
| LSP-T13 | Go fixture with registered generated symbol plus `strings.TrimSpace` | Only the registered symbol produces a semantic delta; helper calls remain implementation details. |
| LSP-T14 | `go list` failure, type error, and canceled Go subprocess | Diagnostics are actionable, stderr is retained as evidence, and cancellation terminates the work. |
| LSP-T15 | Parallel hover/completion/diagnostics over changing snapshots | Results are tied to snapshot hashes; `go test -race` remains clean. |

## Experiment design: correctness before optimization

The following experiments turn the checklist into falsifiable hypotheses. Each
experiment has a fixture, an observable, and a pass criterion. A passing measurement
does not prove protocol conformance by itself: correctness fixtures must pass first,
and performance budgets are only meaningful when the fixture, toolchain, CPU, and
cache state are recorded with the result.

### UTF-8 storage versus UTF-16 LSP positions

**Hypothesis H-UTF16:** keeping source spans as UTF-8 byte offsets and converting only
at the LSP boundary is lossless for every valid position that does not split an encoded
character. It is both safer and cheaper than storing editor columns in the parser or
semantic IR.

**Fixture U-01** is a valid `.gooo` document containing ASCII, Korean identifiers,
non-BMP text in a semantic ID, combining marks, `\r\n` line endings, an empty line,
and a token immediately after each of those cases. For example, the semantic ID may
contain `😀` while the declaration name is `주문`; the fixture must not rely on emoji
being a valid identifier. The harness records the byte range of every token and the
LSP range returned for it.

Measure both directions for every safe boundary:

```text
byte offset -> (line, UTF-16 character) -> byte offset
LSP range  -> byte range -> apply edit -> expected UTF-8 text
```

The expected result is exact equality for safe boundaries, monotonicity within a line,
and no slicing in the middle of a UTF-8 sequence. A requested position beyond the
line clamps to the line end as specified. A position inside a surrogate pair is not a
valid semantic boundary; the server must clamp deterministically and must never panic
or silently address the following character. The same converter must be used by
diagnostics, `LocationLink`, completion edits, semantic tokens, and workspace edits.

**Pass criteria:** 100% of fixture round trips match; all returned ranges are valid
for the exact snapshot that produced them; the UTF-16 implementation agrees with a
reference converter for ASCII, BMP, non-BMP, and combining-mark cases; no test reports
an invalid UTF-8 slice. Add a regression case whenever a client reports a position
encoding mismatch.

### Incremental diagnostics and invalidation

There are two different meanings of “incremental diagnostics” and they should not be
confused:

1. **Incremental computation:** after a change, reuse snapshots and semantic
   dependencies so only the affected closure is re-lexed, lowered, type-checked, and
   diagnosed.
2. **Incremental protocol result:** in LSP 3.17+, `textDocument/diagnostic` can carry
   `previousResultId`; the server returns a full report with a new `resultId` or an
   `unchanged` report with the existing one. Push diagnostics still replace the full
   URI set; they are not item-level patches.

The [pull diagnostics specification](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/pullDiagnostics.md)
also allows workspace reports to carry per-document versions and previous result IDs.
That is a better fit for the semantic graph than inventing a custom diagnostic delta.
The initial implementation can keep push diagnostics, but its internal result cache
should use the same model so a later pull provider is a projection change rather than
a second diagnostic engine.

**Hypothesis H-DIAG:** for an edit that changes one declaration, the affected
diagnostic closure is exactly the declaration's dependency closure; unaffected
documents reuse their prior diagnostic result ID; a clean full rebuild produces the
same diagnostics and ordering as incremental computation.

**Fixture D-01** has ten `.gooo` documents: one declaration file, several activities
that use it, one namespace collision, and unrelated documents. Run these steps:

1. Open all files at version 1 and compute a clean baseline.
2. Change only the declaration's display name, then only its stable ID, then only an
   unrelated comment.
3. Repeat each change with the dependency index warm and cold.
4. Race an older slow diagnostic job against a newer version and record which
   publication wins.

Record `snapshotVersion`, `snapshotHash`, `affectedURIs`, `parseCount`,
`lowerCount`, `diagnosticCount`, `publishedCount`, `staleDroppedCount`, cache hits,
and the result ID for each URI. A result ID should be derived from the diagnostic
inputs and verifier/generator version, not from wall-clock time.

**Pass criteria:**

- Incremental diagnostics are byte-for-byte and order-for-order equivalent to a clean
  full rebuild for the same snapshot.
- The affected URI set is neither smaller than the semantic dependency closure nor
  larger than the declared policy allows; unrelated comment edits do not invalidate
  semantic diagnostics.
- An unchanged pull returns `kind: "unchanged"` only when the client supplied the
  matching previous result ID; an unknown or evicted ID returns a full report.
- An older job never publishes over a newer snapshot. Push mode publishes an empty
  array when all diagnostics disappear; pull mode returns a full report with zero
  items.
- A workspace report never wins over a newer document-pull report for the same URI.

### Workspace edits and semantic rename

**Hypothesis H-EDIT:** a workspace edit derived from one stable semantic ID is
deterministic, non-overlapping, version-safe, and local. Reapplying the edit to the
same snapshots yields the same semantic IR after normalization, while stable IDs and
unrelated generated regions remain unchanged.

**Fixture E-01** contains:

- an open `.gooo` file declaring `billing://entity/order`;
- an unopened `.gooo` file referring to the same ID;
- a generated Go file with a marked generated region and a source-map entry;
- a handwritten Go slot that mentions the display name only as a comment; and
- a second namespace with a same-spelled but different semantic ID.

Exercise both client capability branches: `WorkspaceEdit.changes` for a client that
does not support versioned document changes, and `documentChanges` with versions for a
client that does. Apply edits in a test client in descending range order per document
only after validating that the server-provided ranges are non-overlapping and within
the snapshot version. The descending application is a harness detail; the wire edit
itself must remain a set of valid, non-overlapping `TextEdit`s.

Measure URI count, edit count, ranges, changed bytes, semantic IDs touched, generated
regions touched, compute latency, and apply latency. Re-run the request after a
version bump and after a same-spelled symbol is added to another namespace.

**Pass criteria:**

- The expected declaration and identity-resolved references change, and no other
  semantic declaration changes.
- The stable `id "..."` literal is never an edit target.
- An edit with a stale document version is rejected or recomputed; it is never
  silently applied.
- Applying the edit and then regenerating Go gives the same normalized IR and the
  same generated-region boundaries as a clean regeneration.
- Serialized workspace edits are deterministic across repeated runs and contain no
  overlapping ranges, duplicate URI operations, or edits outside the source snapshot.
- The capability branch is truthful: versioned `documentChanges` are not sent to a
  client that did not advertise support.

### Generated-region navigation

Generated regions are a projection boundary, not an authority boundary. Navigation
should make the relationship visible without encouraging edits directly in generated
text. LSP's [`LocationLink`](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/locationLink.md)
provides exactly the needed shape:

```text
originSelectionRange   generated Go symbol or call
targetUri              .gooo source URI
targetRange            complete declaration span
targetSelectionRange   declaration name/ID to reveal
```

**Hypothesis H-GEN:** with a source map keyed by stable semantic ID and generated
region marker, definition and hover can navigate from generated Go to the authoritative
DSL declaration and from DSL to its generated projection without relying on line
numbers or display-name coincidence.

**Fixture G-01** has two generated activities, two handwritten slots, one regenerated
region shifted by inserted lines, and one intentionally stale/missing source-map
entry. Query definition at the generated declaration, the generated call site, the
DSL declaration, and the handwritten slot. Then regenerate with an unrelated DSL
declaration added.

**Pass criteria:**

- Every valid mapped query returns the expected URI and ranges; `targetSelectionRange`
  is contained by `targetRange` and points at the semantic declaration name.
- Adding or removing lines in a generated region does not change the target DSL range
  or stable identity.
- A stale source map returns no fabricated target and emits a diagnosable stale-map
  result; it never navigates to a neighboring symbol with the same display name.
- Regeneration changes only the expected generated region. Handwritten slots and
  unrelated generated regions remain byte-stable.
- The reverse projection is explicit: if a DSL declaration has no generated target,
  definition returns `null` rather than a guessed file location.

### Cancellation and timeout behavior

LSP cancellation is a notification, but the canceled request still needs a response;
the server must not leave it pending. The protocol distinguishes client cancellation
(`RequestCancelled`, `-32800`) from server cancellation (`ServerCancelled`, `-32802`)
and content modified (`ContentModified`, `-32801`). For diagnostic pull, a server
cancel response can carry `retriggerRequest` so the client knows whether to retry.

The current baseline processes one framed message synchronously. Therefore a
`$/cancelRequest` queued behind a slow parse cannot be observed until that parse
returns. This is a measured architectural limitation, not a reason to pretend the
request was canceled. A future concurrent dispatcher must preserve one writer for
responses and must correlate cancellation by request ID.

**Hypothesis H-CANCEL:** every cancellable phase reaches a cancellation checkpoint
within a bounded interval, returns exactly one response, releases temporary state,
and cannot publish diagnostics/tokens/edits from the canceled snapshot. A server
timeout behaves like explicit server cancellation, with a feature-appropriate error
and no leaked child process or goroutine.

**Fixture C-01** uses an injectable blocking parser, a blocking semantic index, and a
blocking `go list` runner. Each dependency has barriers before work, during work, and
immediately before publication. Run cancellation at every barrier and run a deadline
that expires at each barrier. Also cancel after the result has been queued but before
the single writer emits it.

Record request ID, snapshot version, cancellation source, checkpoint, elapsed time to
acknowledgment, response code, child-process status, goroutine count, and publication
count. Use `exec.CommandContext` for Go subprocesses and check `ctx.Err()` before
expensive phases and immediately before publishing a result.

**Pass criteria:**

- A client-canceled request returns one response with `RequestCancelled` (or a
  successful partial result where the feature explicitly permits it), never zero or
  two responses.
- A server deadline returns `ServerCancelled` with a documented retry policy; a
  diagnostic pull includes `retriggerRequest` when appropriate.
- No canceled result is published, no newer result is overwritten, and no child
  process or goroutine remains after the bounded cleanup window.
- On the reference runner, a cancellation already observed at a checkpoint is
  acknowledged within 100 ms p95 and 250 ms p99. This is an initial budget to be
  measured, not a protocol guarantee for a blocked OS call.
- The synchronous baseline is marked non-conformant for mid-request cancellation
  until the dispatcher can read and route cancellation concurrently.

## Conformance fixture contract

The future LSP test harness should be an in-memory client/server conversation over the
real Content-Length framing. Fixtures must be data, not test-specific control flow, so
they can be replayed against alternative transports or a guardian implementation.

Each fixture should define these logical files:

```text
fixture/
  manifest.json       protocol version, position encoding, scale, feature flags
  workspace/          .gooo, generated Go, handwritten Go, and source-map inputs
  steps.json           ordered requests, notifications, delays, edits, cancellations
  expected.json        responses, notifications, diagnostics, tokens, edits, metrics
```

`manifest.json` records the toolchain, OS/architecture, `GOMAXPROCS`, and whether the
run is cold-cache or warm-cache. `steps.json` may use named barriers for cancellation
and timeout experiments but must not encode implementation-private function names.
`expected.json` uses canonical ordering for diagnostics, tokens, locations, and edit
operations. It also records messages that must not appear, which is essential for
notification and stale-publication conformance.

| Fixture | Primary assertion | Scale variants |
| --- | --- | --- |
| U-01 `unicode-boundary` | UTF-8 byte storage and UTF-16 LSP ranges agree | tiny, long line, CRLF |
| D-01 `incremental-diagnostics` | dependency closure and result IDs are correct | 10 files, 100 files |
| E-01 `semantic-rename` | versioned workspace edits preserve identity/locality | open-only, open+disk |
| G-01 `generated-navigation` | source-map `LocationLink` survives regeneration | one region, many regions |
| C-01 `cancel-timeout` | cancellation/timeout response and cleanup are bounded | parser, index, subprocess |
| S-01 `interactive-small` | editor requests meet interactive budgets | 1 file, 10 files |
| S-02 `workspace-medium` | indexing and cross-file features stay bounded | 100 files, 20k lines |

The harness should compare protocol messages after normalizing only nondeterministic
fields explicitly listed in the manifest (for example a server version). It must not
normalize URI, range, semantic ID, result ID, diagnostic order, or edit order. A
fixture passes only when both the semantic result and the protocol trace pass.

## Latency and allocation budgets

These are starting hypotheses for a reference runner, not promises about every editor
or workspace. The reference run should use a pinned Go toolchain, a fixed OS/arch,
`GOMAXPROCS=1`, no race instrumentation, and ten or more repetitions. Report warm and
cold cache separately. A regression is actionable when p95 exceeds the absolute
budget or is more than 20% slower than the checked-in baseline for the same fixture.
Do not turn a noisy cross-machine measurement into a merge gate without recording the
environment and variance.

| Operation | Fixture | Initial p95 hypothesis | Initial allocation hypothesis |
| --- | --- | ---: | ---: |
| UTF-16 range conversion | U-01, 100 KiB longest line | ≤ 100 µs/op | ≤ 2 allocs/op |
| Parse + lower one open document | S-01, 2k source lines | ≤ 20 ms/op | ≤ 2× input bytes/op |
| Incremental diagnostics, one affected of 100 docs | D-01/S-02 | ≤ 50 ms/op | ≤ 1 MB/op |
| No-change diagnostic pull | D-01 warm cache | ≤ 10 ms/op | ≤ 32 KB/op |
| Completion, hover, same-workspace definition | S-02 warm index | ≤ 20 ms/op | ≤ 64 KB/op |
| Full semantic tokens | S-02, 20k source lines | ≤ 100 ms/op | ≤ 4× encoded token payload/op |
| Cross-file semantic rename edit | E-01/S-02 | ≤ 100 ms/op | ≤ 1 MB/op |
| Generated-region navigation lookup | G-01 warm source map | ≤ 5 ms/op | ≤ 16 KB/op |
| Go discovery/type bridge, warm subprocess data | S-02 | ≤ 500 ms/op | report, do not gate initially |
| Cancellation after checkpoint | C-01 | ≤ 100 ms p95 / 250 ms p99 ack | zero retained work after cleanup |

Measure benchmarks with Go's `testing.B`, `b.ReportAllocs()`, `-benchmem`,
`-count=10`, and a fixed `-benchtime` such as `200ms`. Use `testing.AllocsPerRun`
for focused converters and edit application where a stable allocation count is more
useful than a broad benchmark. Keep setup (fixture loading, index construction, and
cache warming) outside the timed operation unless the operation explicitly includes
it. For investigation rather than gating, capture CPU and heap profiles with
[`runtime/pprof`](https://pkg.go.dev/runtime/pprof); do not commit profiles or make a
profile-dependent pass criterion.

Latency evidence should include p50, p95, p99, min/max, `ns/op`, `B/op`,
`allocs/op`, cache hit rate, affected URI count, and stale/canceled work count. A
benchmark that passes latency while recomputing the entire workspace is not a pass:
the semantic closure and cache counters are part of the acceptance result. Run
`-race` separately because race instrumentation changes latency and allocation
profiles; it remains a correctness gate, not a performance baseline.

The budgets deliberately separate deterministic in-process features from `go list`
and type-check subprocesses. If Go tool startup dominates the latter, measure warm
and cold startup as separate evidence and improve cancellation/timeout behavior before
attempting to tighten the numeric budget.

## Source references

- [LSP 3.18 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/)
- [Initialize and capability negotiation](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/general/initialize.md)
- [Position and encoding](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/position.md)
- [Text document synchronization](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/specification.md#textDocument_synchronization)
- [Semantic tokens](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/semanticTokens.md)
- [Diagnostics](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/publishDiagnostics.md)
- [Workspace edits](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/workspaceEdit.md)
- [Pull diagnostics and result IDs](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/pullDiagnostics.md)
- [Location links](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/locationLink.md)
- [Completion](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/completion.md), [definition](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/definition.md), and [hover](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/hover.md)
- Go standard library: [`go/parser`](https://pkg.go.dev/go/parser), [`go/ast`](https://pkg.go.dev/go/ast), [`go/token`](https://pkg.go.dev/go/token), [`go/types`](https://pkg.go.dev/go/types), and [`go/importer`](https://pkg.go.dev/go/importer)
- Go workspace discovery: [`go list`](https://pkg.go.dev/cmd/go#hdr-List_packages)
- Go benchmark allocation and profiling APIs: [`testing.B.ReportAllocs`](https://pkg.go.dev/testing#B.ReportAllocs), [`testing.AllocsPerRun`](https://pkg.go.dev/testing#AllocsPerRun), [`go test` benchmark flags](https://pkg.go.dev/cmd/go#hdr-Testing_flags), and [`runtime/pprof`](https://pkg.go.dev/runtime/pprof)
- [`gopls` behavior and workspace model](https://go.dev/gopls/) and [workspace setup](https://go.dev/gopls/workspace)
