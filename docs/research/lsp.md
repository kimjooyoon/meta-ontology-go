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

## Source references

- [LSP 3.18 specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/)
- [Initialize and capability negotiation](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/general/initialize.md)
- [Position and encoding](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/position.md)
- [Text document synchronization](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/specification.md#textDocument_synchronization)
- [Semantic tokens](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/semanticTokens.md)
- [Diagnostics](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/publishDiagnostics.md)
- [Workspace edits](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/types/workspaceEdit.md)
- [Completion](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/completion.md), [definition](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/definition.md), and [hover](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.18/language/hover.md)
- Go standard library: [`go/parser`](https://pkg.go.dev/go/parser), [`go/ast`](https://pkg.go.dev/go/ast), [`go/token`](https://pkg.go.dev/go/token), [`go/types`](https://pkg.go.dev/go/types), and [`go/importer`](https://pkg.go.dev/go/importer)
- Go workspace discovery: [`go list`](https://pkg.go.dev/cmd/go#hdr-List_packages)
- [`gopls` behavior and workspace model](https://go.dev/gopls/) and [workspace setup](https://go.dev/gopls/workspace)
