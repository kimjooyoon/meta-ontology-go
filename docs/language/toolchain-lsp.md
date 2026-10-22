# Toolchain LSP

Gooo already has a dependency-free language server. This increment does not replace it. It makes the existing server measurable and binds its editor projection to meta-code.

The fixed `gooo/toolchain-lsp-corpus/v1` contract contains 22 cases: 16 JSON-RPC/LSP protocol cases and 6 immutable semantic-coupling cases. It executes initialize, synchronized open/change/close, diagnostics, hover, completion, definition, references, document/workspace symbols, semantic tokens, shutdown/exit, UTF-16 positions, and deterministic replay.

The implemented subset is checked against the official LSP 3.18 specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/

The CLI's `initialize.serverInfo.version` identifies the running Gooo language version.
A server created directly without version metadata reports `unspecified`. Agents that
need the exact compiler source, module, and toolchain should query the same executable
with `gooo version --build --json`; the LSP version is a display hint. The current
LSP CLI remains experimental, so this handshake does not establish full editor compatibility.

An unresolved activity input or output now carries its authored reference span
through lowering. `gooo check --semantic --json`, text check, `gooo query
--json`, and the LSP diagnostic point to that occurrence. The span comes from
the syntax node, including the output token, rather than a search through the
error message. Other semantic failures may still report a broader span.

The meta-operation is `project-exact-language-state-to-editor-protocol`. Every one of the 37 indicators names `toolchainlsp.Evaluate` as producer and `self-improvement-cycle` as consumer. The indicators are split into 3 outcomes, 16 drivers, and 18 guardrails.

The Munchhausen choices are explicit. FOUNDATION binds the fixed corpus, exact head, concept, and advertised capabilities. COHERENCE binds parser/semantic state and coupling evidence to standard LSP values. REGRESSION requires UNKNOWN, FAIL_CLOSED, stale, cancellation, and unsupported-method paths to expose no unauthorized navigation.

The gomacro project is a structural reference for staged AST transformation boundaries: https://github.com/cosmos72/gomacro. Gooo does not embed gomacro, inherit ambient Go evaluation authority, or make a novelty claim. Its LSP projects already parsed and normalized state through a read-only protocol boundary.
