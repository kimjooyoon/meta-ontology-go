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
		if err := validateInputNameSource(*d, source); err != nil {
			return err
		}
		line := bytes.Count(source[:d.Start], []byte("\n")) + 1
		begin := bytes.LastIndexByte(source[:d.Start], '\n') + 1
		column := utf8.RuneCount(source[begin:d.Start]) + 1
		_, err := fmt.Fprintf(w, "  signature %s at %d:%d %s type=%s\n",
			d.Port, line, column, strconv.Quote(d.Expression), strconv.Quote(d.TypeID))
		if err != nil || d.Name == "" {
			return err
		}
		line = bytes.Count(source[:d.NameStart], []byte("\n")) + 1
		begin = bytes.LastIndexByte(source[:d.NameStart], '\n') + 1
		column = utf8.RuneCount(source[begin:d.NameStart]) + 1
		_, err = fmt.Fprintf(w, "  input name %s at %d:%d\n", strconv.Quote(d.Name), line, column)
		return err
	}
	return nil
}

func validateInputNameSource(d TypedInputDeclaration, source []byte) error {
	if d.Name == "" {
		if d.NameStart != 0 || d.NameEnd != 0 {
			return fmt.Errorf("unnamed input has an authored name range")
		}
		return nil
	}
	if d.NameStart < 0 || d.NameEnd <= d.NameStart || d.NameEnd > d.Start || d.NameEnd > len(source) ||
		string(source[d.NameStart:d.NameEnd]) != d.Name {
		return fmt.Errorf("origin input name differs")
	}
	return nil
}
