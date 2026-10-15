package bodycodegen

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

func writeOriginInputDeclaration(w io.Writer, report TypedValueOrigins, node uint16, source []byte) error {
	for _, input := range report.Inputs {
		if input.Node != node || input.Declaration == nil {
			continue
		}
		d := input.Declaration
		if d.Status != "MAPPED" {
			_, err := fmt.Fprintf(w, "  input declaration: %s\n", d.Reason)
			return err
		}
		if d.SourceSHA256 != report.SourceSHA256 || d.ActivityID != report.ActivityID || d.Port != input.Port ||
			d.Start < 0 || d.End <= d.Start || d.End > len(source) || string(source[d.Start:d.End]) != d.Expression {
			return fmt.Errorf("origin input declaration differs")
		}
		line := bytes.Count(source[:d.Start], []byte("\n")) + 1
		begin := bytes.LastIndexByte(source[:d.Start], '\n') + 1
		column := utf8.RuneCount(source[begin:d.Start]) + 1
		_, err := fmt.Fprintf(w, "  signature %s at %d:%d %s type=%s\n",
			d.Port, line, column, strconv.Quote(d.Expression), strconv.Quote(d.TypeID))
		return err
	}
	return nil
}
