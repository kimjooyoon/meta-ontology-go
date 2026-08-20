package main

import (
	"go/scanner"
	"go/token"
	"strings"
)

func oneLineTokens(data []byte) (string, bool) {
	files := token.NewFileSet()
	file := files.AddFile("density.go", files.Base(), len(data))
	failures := 0
	var lexer scanner.Scanner
	lexer.Init(file, data, func(token.Position, string) { failures++ }, scanner.ScanComments)
	var output strings.Builder
	first := true
	for {
		_, symbol, literal := lexer.Scan()
		if symbol == token.EOF {
			break
		}
		if symbol == token.COMMENT || symbol == token.ILLEGAL {
			return "", false
		}
		text := literal
		if text == "" {
			text = symbol.String()
		}
		if symbol == token.SEMICOLON {
			text = ";"
		}
		if !first {
			output.WriteByte(' ')
		}
		output.WriteString(text)
		first = false
	}
	return output.String(), failures == 0
}
