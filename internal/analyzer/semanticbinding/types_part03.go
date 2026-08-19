package semanticbinding

import (
	"fmt"
	"strings"
)

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Span.Filename == "" {
		return string(e.Code) + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %s", e.Span.Filename, e.Span, e.Message)
}
func (s Span) String() string {
	if s.Filename == "" {
		return fmt.Sprintf("%d:%d-%d:%d", s.Start.Line, s.Start.Column, s.End.Line, s.End.Column)
	}
	return fmt.Sprintf("%s:%d:%d-%d:%d", s.Filename, s.Start.Line, s.Start.Column, s.End.Line, s.End.Column)
}
func (r Role) valid() bool {
	return r == RoleHandwrittenImpl || r == RoleGeneratedImpl || r == RoleAdapter
}
func (s Span) valid() bool {
	return s.Filename != "" && s.Start.Offset >= 0 && s.End.Offset >= s.Start.Offset
}
func (i Input) sourceInputs() ([]SourceFile, error) {
	if len(i.Sources) > 0 && len(i.Files) > 0 {
		return nil, &Error{Code: CodeInput, Message: "use Sources or Files, not both", FullSuiteFallback: true}
	}
	sources := i.Sources
	if len(sources) == 0 {
		sources = i.Files
	}
	if len(sources) == 0 {
		return nil, &Error{Code: CodeInput, Message: "at least one source file is required", FullSuiteFallback: true}
	}
	result := make([]SourceFile, len(sources))
	copy(result, sources)
	for index := range result {
		if result[index].PackagePath == "" {
			result[index].PackagePath = i.PackagePath
		}
		if strings.TrimSpace(result[index].Filename) == "" || strings.TrimSpace(result[index].PackagePath) == "" {
			return nil, &Error{Code: CodeInput, Message: "filename and package path are required", FullSuiteFallback: true}
		}
		if result[index].Source == nil {
			return nil, &Error{Code: CodeInput, Message: "source bytes are required", FullSuiteFallback: true}
		}
		result[index].Filename = strings.TrimSpace(result[index].Filename)
		result[index].PackagePath = strings.TrimSpace(result[index].PackagePath)
	}
	return result, nil
}
