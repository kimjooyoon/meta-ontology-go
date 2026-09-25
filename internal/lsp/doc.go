// Package lsp provides a dependency-free JSON-RPC/LSP server for .gooo files.
// ServeContext returns on cancellation even when its input is not an io.Closer.
// For a non-closeable reader, the underlying blocked read cannot be interrupted
// by standard Go I/O and must eventually return for its goroutine to exit.
package lsp
