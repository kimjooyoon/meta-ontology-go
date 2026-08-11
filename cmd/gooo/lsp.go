package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type lspRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type lspResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
	Error   *lspError       `json:"error,omitempty"`
}

type lspError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type lspServer struct {
	documents map[string]string
	shutdown  bool
}

func runLSP(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: gooo lsp")
		return exitUsage
	}
	server := lspServer{documents: make(map[string]string)}
	return server.serve(stdin, stdout, stderr)
}

func (s *lspServer) serve(stdin io.Reader, stdout, stderr io.Writer) int {
	reader := bufio.NewReader(stdin)
	for {
		payload, err := readLSPMessage(reader)
		if err == io.EOF {
			return exitOK
		}
		if err != nil {
			fmt.Fprintln(stderr, "lsp:", err)
			return exitFailure
		}
		var request lspRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			fmt.Fprintln(stderr, "lsp: invalid JSON:", err)
			return exitFailure
		}
		if request.Method == "exit" {
			if s.shutdown {
				return exitOK
			}
			return exitFailure
		}
		if err := s.handle(request, stdout); err != nil {
			fmt.Fprintln(stderr, "lsp:", err)
			return exitFailure
		}
	}
}

func (s *lspServer) handle(request lspRequest, stdout io.Writer) error {
	switch request.Method {
	case "initialize":
		return writeLSPResponse(stdout, request.ID, map[string]any{
			"capabilities": map[string]any{"textDocumentSync": 1},
		}, nil)
	case "initialized", "$/cancelRequest":
		return nil
	case "shutdown":
		s.shutdown = true
		return writeLSPResponse(stdout, request.ID, nil, nil)
	case "textDocument/didOpen":
		return s.didOpen(request.Params, stdout)
	case "textDocument/didChange":
		return s.didChange(request.Params, stdout)
	case "textDocument/didClose":
		return s.didClose(request.Params)
	case "textDocument/hover":
		return writeLSPResponse(stdout, request.ID, nil, nil)
	case "textDocument/completion":
		return writeLSPResponse(stdout, request.ID, map[string]any{
			"isIncomplete": false,
			"items":        []any{},
		}, nil)
	default:
		if len(request.ID) == 0 {
			return nil
		}
		return writeLSPResponse(stdout, request.ID, nil, &lspError{
			Code:    -32601,
			Message: "method not found: " + request.Method,
		})
	}
}

type lspTextDocument struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

type lspDidOpen struct {
	TextDocument lspTextDocument `json:"textDocument"`
}

func (s *lspServer) didOpen(params json.RawMessage, stdout io.Writer) error {
	var value lspDidOpen
	if err := json.Unmarshal(params, &value); err != nil {
		return err
	}
	s.documents[value.TextDocument.URI] = value.TextDocument.Text
	return s.publish(value.TextDocument.URI, value.TextDocument.Text, stdout)
}

type lspTextChange struct {
	Text string `json:"text"`
}

type lspDidChange struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	ContentChanges []lspTextChange `json:"contentChanges"`
}

func (s *lspServer) didChange(params json.RawMessage, stdout io.Writer) error {
	var value lspDidChange
	if err := json.Unmarshal(params, &value); err != nil {
		return err
	}
	if len(value.ContentChanges) == 0 {
		return nil
	}
	text := value.ContentChanges[len(value.ContentChanges)-1].Text
	s.documents[value.TextDocument.URI] = text
	return s.publish(value.TextDocument.URI, text, stdout)
}

func (s *lspServer) didClose(params json.RawMessage) error {
	var value struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(params, &value); err != nil {
		return err
	}
	delete(s.documents, value.TextDocument.URI)
	return nil
}

func (s *lspServer) publish(uri, source string, stdout io.Writer) error {
	diagnostics := lspDiagnostics(uri, source)
	return writeLSPNotification(stdout, "textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"`
	Code     string   `json:"code,omitempty"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

func lspDiagnostics(uri, source string) []lspDiagnostic {
	file, diagnostics := syntax.ParseFile(uri, source)
	result := make([]lspDiagnostic, 0, len(diagnostics)+1)
	for _, diagnostic := range diagnostics.SortBySpan() {
		severity := 1
		if diagnostic.Severity == syntax.SeverityWarning {
			severity = 2
		}
		result = append(result, lspDiagnostic{
			Range:    syntaxRange(diagnostic.Span),
			Severity: severity,
			Code:     string(diagnostic.Code),
			Source:   "gooo",
			Message:  diagnostic.Message,
		})
	}
	if diagnostics.Error() == nil {
		if _, err := bidir.Lower(file); err != nil {
			result = append(result, lspDiagnostic{
				Range:   lspRange{},
				Code:    "semantic.lower",
				Source:  "gooo",
				Message: err.Error(),
			})
		}
	}
	return result
}

func syntaxRange(span syntax.Span) lspRange {
	return lspRange{
		Start: lspPosition{Line: nonNegative(span.Start.Line - 1), Character: nonNegative(span.Start.Column - 1)},
		End:   lspPosition{Line: nonNegative(span.End.Line - 1), Character: nonNegative(span.End.Column - 1)},
	}
}

func nonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func readLSPMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed header %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(parts[0]), "Content-Length") {
			value, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil || value < 0 {
				return nil, fmt.Errorf("invalid Content-Length %q", parts[1])
			}
			contentLength = value
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeLSPResponse(out io.Writer, id json.RawMessage, result any, responseError *lspError) error {
	return writeLSPMessage(out, lspResponse{JSONRPC: "2.0", ID: id, Result: result, Error: responseError})
}

func writeLSPNotification(out io.Writer, method string, params any) error {
	return writeLSPMessage(out, struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{JSONRPC: "2.0", Method: method, Params: params})
}

func writeLSPMessage(out io.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err = out.Write(payload)
	return err
}
