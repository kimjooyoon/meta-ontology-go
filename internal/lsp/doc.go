// Package lsp provides a dependency-free JSON-RPC/LSP server for .gooo files.
// ServeContext can unblock a read on cancellation when its input implements
// io.Closer; non-closeable readers cannot be interrupted by standard Go I/O.
package lsp
