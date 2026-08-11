package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrExitWithoutShutdown = errors.New("lsp: exit received before shutdown")
	ErrInvalidVersion      = errors.New("lsp: document version is not newer")
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
}

func NewServer(parsers ...Parser) *Server {
	parser := Parser(SyntaxParser{})
	if len(parsers) > 0 && parsers[0] != nil {
		parser = parsers[0]
	}
	return &Server{parser: parser, documents: make(map[string]*document)}
}

func (server *Server) Serve(input io.Reader, output io.Writer) error {
	return server.ServeContext(context.Background(), input, output)
}

func (server *Server) ServeContext(ctx context.Context, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		payload, err := readFrame(reader)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		response, notifications, dispatchErr := server.dispatch(ctx, payload)
		if dispatchErr != nil {
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
		if server.exited {
			if server.shutdown {
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
	if server.shutdown && request.Method != "exit" {
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
		server.exited = true
		return nil, nil, nil
	case "textDocument/didOpen":
		return server.didOpen(ctx, request)
	case "textDocument/didChange":
		return server.didChange(ctx, request)
	case "textDocument/didClose":
		return server.didClose(request)
	case "textDocument/hover":
		return server.hoverRequest(request)
	case "textDocument/completion":
		return server.completionRequest(request)
	case "textDocument/definition":
		return server.definitionRequest(request)
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
	server.initialized = true
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   TextDocumentSyncOptions{OpenClose: true, Change: 1},
			HoverProvider:      true,
			CompletionProvider: &CompletionOptions{},
			DefinitionProvider: true,
		},
		ServerInfo: ServerInfo{Name: "gooo-lsp", Version: "current-ddaf"},
	}
	return resultResponse(request.ID, result), nil, nil
}

func (server *Server) shutdownRequest(request requestEnvelope) *responseEnvelope {
	server.shutdown = true
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
	if _, exists := server.documents[params.TextDocument.URI]; exists {
		return responseOrNil(request.ID, invalidParams, "Document is already open"), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, params.TextDocument.Text)
	if err != nil {
		return server.parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	server.documents[params.TextDocument.URI] = &document{
		version: params.TextDocument.Version, text: params.TextDocument.Text, result: result,
	}
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (server *Server) didChange(ctx context.Context, request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidChangeTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didChange parameters"), nil, nil
	}
	document, exists := server.documents[params.TextDocument.URI]
	if !exists {
		return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
	}
	if params.TextDocument.Version <= document.version {
		return responseOrNil(request.ID, invalidParams, ErrInvalidVersion.Error()), nil, nil
	}
	text, err := applyChanges(document.text, params.ContentChanges)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, err.Error()), nil, nil
	}
	result, err := server.parse(ctx, params.TextDocument.URI, text)
	if err != nil {
		return server.parseFailure(request.ID, ctx, err), nil, errIfCanceled(ctx, err)
	}
	document.version, document.text, document.result = params.TextDocument.Version, text, result
	notification, err := diagnosticsNotification(params.TextDocument.URI, result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (server *Server) didClose(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidCloseTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didClose parameters"), nil, nil
	}
	if _, exists := server.documents[params.TextDocument.URI]; !exists {
		return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
	}
	delete(server.documents, params.TextDocument.URI)
	notification, err := diagnosticsNotification(params.TextDocument.URI, nil)
	return nil, oneNotification(notification, err), err
}

func (server *Server) hoverRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid hover parameters"), nil, nil
	}
	hover, _ := server.hover(params)
	return resultResponse(request.ID, hover), nil, nil
}

func (server *Server) completionRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid completion parameters"), nil, nil
	}
	return resultResponse(request.ID, server.completion(params.TextDocument.URI)), nil, nil
}

func (server *Server) definitionRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return responseOrNil(request.ID, invalidParams, "Invalid definition parameters"), nil, nil
	}
	return resultResponse(request.ID, server.definition(params)), nil, nil
}

func (server *Server) parse(ctx context.Context, uri, source string) (ParseResult, error) {
	if parser, ok := server.parser.(ContextParser); ok {
		return parser.ParseContext(ctx, uri, source)
	}
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	return server.parser.Parse(uri, source), nil
}

func (server *Server) parseFailure(id json.RawMessage, ctx context.Context, err error) *responseEnvelope {
	if ctx.Err() != nil {
		return nil
	}
	return responseOrNil(id, internalError, err.Error())
}

func errIfCanceled(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func decodeParams(raw json.RawMessage, target any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func responseOrNil(id json.RawMessage, code int, message string) *responseEnvelope {
	if id == nil {
		return nil
	}
	return errorResponse(id, code, message)
}

func resultResponse(id json.RawMessage, result any) *responseEnvelope {
	encoded, err := json.Marshal(result)
	if err != nil {
		return errorResponse(id, internalError, err.Error())
	}
	return &responseEnvelope{JSONRPC: jsonRPCVersion, ID: responseID(id), Result: encoded}
}

func errorResponse(id json.RawMessage, code int, message string) *responseEnvelope {
	return &responseEnvelope{JSONRPC: jsonRPCVersion, ID: responseID(id), Error: &errorObject{Code: code, Message: message}}
}

func responseID(id json.RawMessage) json.RawMessage {
	if id == nil {
		return json.RawMessage("null")
	}
	return id
}

func writeResponse(output io.Writer, response *responseEnvelope) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("lsp: encode response: %w", err)
	}
	return WriteMessage(output, payload)
}

func oneNotification(notification []byte, err error) [][]byte {
	if err != nil || notification == nil {
		return nil
	}
	return [][]byte{notification}
}
