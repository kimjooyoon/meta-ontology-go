package lsp

import (
	"bufio"
	"context"
	"encoding/json"
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
}

// Server implements the supported .gooo text-document LSP surface. Workspace
// and source-map features remain deliberately unadvertised and unsupported.
type Server struct {
	parser      Parser
	documents   map[string]*document
	initialized bool
	shutdown    bool
	exited      bool
	mu          sync.RWMutex
	parseMu     sync.Mutex
	inflight    map[string]*inFlightRequest
}

func NewServer(parsers ...Parser) *Server {
	parser := Parser(SyntaxParser{})
	if len(parsers) > 0 && parsers[0] != nil {
		parser = parsers[0]
	}
	return &Server{parser: parser, documents: make(map[string]*document), inflight: make(map[string]*inFlightRequest)}
}

func (server *Server) Serve(input io.Reader, output io.Writer) error {
	return server.ServeContext(context.Background(), input, output)
}

func (server *Server) ServeContext(ctx context.Context, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	loop := newRequestLoop(server, output)
	for {
		if err := ctx.Err(); err != nil {
			loop.cancelAll()
			return err
		}
		if err := loop.drain(); err != nil {
			loop.cancelAll()
			return err
		}
		payload, err := readFrameContext(ctx, reader, input)
		if errors.Is(err, io.EOF) {
			loop.cancelAll()
			return loop.wait(ctx)
		}
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				loop.cancelAll()
				return ctxErr
			}
			loop.cancelAll()
			return err
		}
		if request, valid := decodeRequest(payload); valid {
			if request.Method == "$/cancelRequest" {
				server.cancelRequest(request)
				continue
			}
			if server.canRunAsync(request) {
				if err := loop.start(ctx, request, payload); err != nil {
					return err
				}
				continue
			}
		}
		response, notifications, dispatchErr := server.dispatch(ctx, payload)
		if dispatchErr != nil {
			loop.cancelAll()
			return dispatchErr
		}
		if response != nil {
			if err := writeResponse(output, response); err != nil {
				return err
			}
		}
		for _, notification := range notifications {
			if err := WriteMessage(output, notification); err != nil {
				return err
			}
		}
		shutdown, exited := server.lifecycleState()
		if exited {
			loop.cancelAll()
			if err := loop.wait(ctx); err != nil {
				return err
			}
			if shutdown {
				return nil
			}
			return ErrExitWithoutShutdown
		}
	}
}

func (server *Server) dispatch(ctx context.Context, payload []byte) (*responseEnvelope, [][]byte, error) {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return errorResponse(nil, parseError, "Parse error"), nil, nil
	}
	if request.JSONRPC != jsonRPCVersion || request.Method == "" {
		return responseOrNil(request.ID, invalidRequest, "Invalid Request"), nil, nil
	}
	shutdown, _ := server.lifecycleState()
	if shutdown && request.Method != "exit" {
		return responseOrNil(request.ID, invalidRequest, "server is shut down"), nil, nil
	}
	switch request.Method {
	case "initialize":
		return server.initialize(request)
	case "initialized", "$/cancelRequest":
		return nil, nil, nil
	case "shutdown":
		return server.shutdownRequest(request), nil, nil
	case "exit":
		server.markExited()
		return nil, nil, nil
	case "textDocument/didOpen":
		return server.didOpen(ctx, request)
	case "textDocument/didChange":
		return server.didChange(ctx, request)
	case "textDocument/didClose":
		return server.didClose(request)
	case "textDocument/hover":
		return server.hoverRequestContext(ctx, request)
	case "textDocument/completion":
		return server.completionRequest(ctx, request)
	case "textDocument/definition":
		return server.definitionRequest(ctx, request)
	case "workspace/symbol", "textDocument/references", "textDocument/rename", "textDocument/formatting":
		return responseOrNil(request.ID, methodNotFound, "method is deferred by this LSP baseline"), nil, nil
	default:
		return responseOrNil(request.ID, methodNotFound, "Method not found"), nil, nil
	}
}

func (server *Server) initialize(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params InitializeParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid initialize parameters"), nil, nil
	}
	server.mu.Lock()
	server.initialized = true
	server.mu.Unlock()
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   TextDocumentSyncOptions{OpenClose: true, Change: 2},
			HoverProvider:      true,
			CompletionProvider: &CompletionOptions{},
			DefinitionProvider: true,
		},
		ServerInfo: ServerInfo{Name: "gooo-lsp", Version: "current-ddaf"},
	}
	return resultResponse(request.ID, result), nil, nil
}

func (server *Server) shutdownRequest(request requestEnvelope) *responseEnvelope {
	server.mu.Lock()
	server.shutdown = true
	server.mu.Unlock()
	if request.ID == nil {
		return nil
	}
	return resultResponse(request.ID, nil)
}

func (server *Server) didOpen(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidOpenTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" || params.TextDocument.Version < 0 {
		return responseOrNil(request.ID, invalidParams, "Invalid didOpen parameters"), nil, nil
	}
	server.mu.RLock()
	_, exists := server.documents[params.TextDocument.URI]
	server.mu.RUnlock()
	if exists {
		return responseOrNil(request.ID, invalidParams, "Document is already open"), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, params.TextDocument.Text)
	if err != nil {
		return parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	server.mu.Lock()
	server.documents[params.TextDocument.URI] = &document{
		version: params.TextDocument.Version, text: params.TextDocument.Text, result: result,
	}
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (server *Server) didChange(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidChangeTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didChange parameters"), nil, nil
	}
	server.mu.RLock()
	document, exists := server.documents[params.TextDocument.URI]
	if exists {
		documentVersion, documentText := document.version, document.text
		server.mu.RUnlock()
		return server.changeDocument(ctx, request, params, documentVersion, documentText)
	}
	server.mu.RUnlock()
	return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
}

func (server *Server) changeDocument(ctx context.Context, request requestEnvelope, params DidChangeTextDocumentParams, version int, source string) (*responseEnvelope, [][]byte, error) {
	if params.TextDocument.Version <= version {
		return responseOrNil(request.ID, invalidParams, ErrInvalidVersion.Error()), nil, nil
	}
	text, err := applyChanges(source, params.ContentChanges)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, err.Error()), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, text)
	if err != nil {
		return parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	server.mu.Lock()
	document, exists := server.documents[params.TextDocument.URI]
	if !exists || document.version != version || document.text != source {
		server.mu.Unlock()
		return nil, nil, nil
	}
	document.version, document.text, document.result = params.TextDocument.Version, text, result
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (server *Server) didClose(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidCloseTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didClose parameters"), nil, nil
	}
	server.mu.Lock()
	if _, exists := server.documents[params.TextDocument.URI]; !exists {
		server.mu.Unlock()
		return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
	}
	delete(server.documents, params.TextDocument.URI)
	server.mu.Unlock()
	notification, err := diagnosticsNotification(params.TextDocument.URI, nil)
	return nil, oneNotification(notification, err), err
}

func (server *Server) hoverRequestContext(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid hover parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	hover, _ := server.hover(params)
	return resultResponse(request.ID, hover), nil, nil
}

func (server *Server) completionRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid completion parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, server.completion(params.TextDocument.URI)), nil, nil
}

func (server *Server) definitionRequest(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid definition parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, server.definition(params)), nil, nil
}
