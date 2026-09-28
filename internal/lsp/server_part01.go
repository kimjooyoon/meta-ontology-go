package lsp

import (
	"context"
	"errors"
	"io"
	"sync"
)

var (
	ErrExitWithoutShutdown = errors.New("lsp: exit received before shutdown")
	ErrInvalidVersion      = errors.New("lsp: document version is not newer")
	ErrStaleResult         = errors.New("lsp: stale result suppressed")
)

type document struct {
	version int
	text    string
	result  ParseResult
	// A key is populated only by refresh; lifecycle parses seed diagnostics but
	// are not reused as feature-refresh evidence.
	cacheKey documentCacheKey
}

// Server implements the supported .gooo text-document LSP surface. Workspace
// symbols use open documents only; edits and source maps remain unsupported.
type Server struct {
	parser        Parser
	documents     map[string]*document
	initialized   bool
	shutdown      bool
	exited        bool
	mu            sync.RWMutex
	parseMu       sync.Mutex
	inflight      map[string]*inFlightRequest
	cacheIdentity documentCacheIdentity
}

func NewServer(parsers ...Parser) *Server {
	parser := Parser(SyntaxParser{})
	if len(parsers) > 0 && parsers[0] != nil {
		parser = parsers[0]
	}
	return &Server{
		parser: parser, documents: make(map[string]*document),
		inflight: make(map[string]*inFlightRequest), cacheIdentity: newDocumentCacheIdentity(parser),
	}
}
func (server *Server) Serve(input io.Reader, output io.Writer) error {
	return server.ServeContext(context.Background(), input, output)
}
