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
	ErrServerExited        = errors.New("lsp: server exited")
	ErrExitWithoutShutdown = errors.New("lsp: exit received before shutdown")
)

type document struct {
	version int
	text    string
	result  ParseResult
}

// Server implements the JSON-RPC/LSP lifecycle and baseline .gooo features.
type Server struct {
	parser      Parser
	documents   map[string]*document
	initialized bool
	shutdown    bool
	exited      bool
}

// NewServer creates a server using the syntax adapter unless a parser is given.
func NewServer(parsers ...Parser) *Server {
	parser := Parser(SyntaxParser{})
	if len(parsers) > 0 && parsers[0] != nil {
		parser = parsers[0]
	}
	return &Server{parser: parser, documents: make(map[string]*document)}
}

// Serve processes framed JSON-RPC messages until input closes or exit arrives.
func (s *Server) Serve(input io.Reader, output io.Writer) error {
	return s.ServeContext(context.Background(), input, output)
}

// ServeContext is Serve with cancellation support between messages.
func (s *Server) ServeContext(ctx context.Context, input io.Reader, output io.Writer) error {
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
		response, notifications, dispatchErr := s.dispatch(payload)
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
		if s.exited {
			if s.shutdown {
				return nil
			}
			return ErrExitWithoutShutdown
		}
	}
}

func (s *Server) dispatch(payload []byte) (*responseEnvelope, [][]byte, error) {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return errorResponse(nil, parseError, "Parse error"), nil, nil
	}
	if request.JSONRPC != jsonRPCVersion || request.Method == "" {
		return errorResponse(request.ID, invalidRequest, "Invalid Request"), nil, nil
	}
	if s.shutdown && request.Method != "exit" {
		return responseOrNil(request.ID, invalidRequest, "server is shut down"), nil, nil
	}
	switch request.Method {
	case "initialize":
		return s.initialize(request)
	case "initialized", "$/cancelRequest":
		return nil, nil, nil
	case "shutdown":
		return s.shutdownRequest(request), nil, nil
	case "exit":
		s.exited = true
		return nil, nil, nil
	case "textDocument/didOpen":
		return s.didOpen(request)
	case "textDocument/didChange":
		return s.didChange(request)
	case "textDocument/didClose":
		return s.didClose(request)
	case "textDocument/hover":
		return s.hoverRequest(request)
	case "textDocument/completion":
		return s.completionRequest(request)
	case "textDocument/definition":
		return s.definitionRequest(request)
	default:
		if request.ID == nil {
			return nil, nil, nil
		}
		return errorResponse(request.ID, methodNotFound, "Method not found"), nil, nil
	}
}

func (s *Server) initialize(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params InitializeParams
	if err := decodeParams(request.Params, &params); err != nil {
		return errorResponse(request.ID, invalidParams, "Invalid initialize parameters"), nil, nil
	}
	s.initialized = true
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   TextDocumentSyncOptions{OpenClose: true, Change: 1},
			HoverProvider:      true,
			CompletionProvider: CompletionOptions{},
			DefinitionProvider: true,
		},
		ServerInfo: ServerInfo{Name: "gooo-lsp", Version: "0.1"},
	}
	return resultResponse(request.ID, result), nil, nil
}

func (s *Server) shutdownRequest(request requestEnvelope) *responseEnvelope {
	s.shutdown = true
	if request.ID == nil {
		return nil
	}
	return resultResponse(request.ID, nil)
}

func (s *Server) didOpen(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidOpenTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didOpen parameters"), nil, nil
	}
	document := &document{version: params.TextDocument.Version, text: params.TextDocument.Text}
	document.result = s.parser.Parse(params.TextDocument.URI, document.text)
	s.documents[params.TextDocument.URI] = document
	notification, err := diagnosticsNotification(params.TextDocument.URI, document.result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (s *Server) didChange(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidChangeTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didChange parameters"), nil, nil
	}
	document, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return responseOrNil(request.ID, invalidParams, "Document is not open"), nil, nil
	}
	text, err := applyChanges(document.text, params.ContentChanges)
	if err != nil {
		return responseOrNil(request.ID, invalidParams, err.Error()), nil, nil
	}
	document.version = params.TextDocument.Version
	document.text = text
	document.result = s.parser.Parse(params.TextDocument.URI, text)
	notification, err := diagnosticsNotification(params.TextDocument.URI, document.result.Diagnostics)
	return nil, oneNotification(notification, err), err
}

func (s *Server) didClose(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params DidCloseTextDocumentParams
	if err := decodeParams(request.Params, &params); err != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid didClose parameters"), nil, nil
	}
	delete(s.documents, params.TextDocument.URI)
	notification, err := diagnosticsNotification(params.TextDocument.URI, nil)
	return nil, oneNotification(notification, err), err
}

func (s *Server) hoverRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return errorResponse(request.ID, invalidParams, "Invalid hover parameters"), nil, nil
	}
	hover, _ := s.hover(params)
	return resultResponse(request.ID, hover), nil, nil
}

func (s *Server) completionRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return errorResponse(request.ID, invalidParams, "Invalid completion parameters"), nil, nil
	}
	return resultResponse(request.ID, s.completion(params.TextDocument.URI)), nil, nil
}

func (s *Server) definitionRequest(request requestEnvelope) (*responseEnvelope, [][]byte, error) {
	var params TextDocumentPositionParams
	if err := decodeParams(request.Params, &params); err != nil {
		return errorResponse(request.ID, invalidParams, "Invalid definition parameters"), nil, nil
	}
	return resultResponse(request.ID, s.definition(params)), nil, nil
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
