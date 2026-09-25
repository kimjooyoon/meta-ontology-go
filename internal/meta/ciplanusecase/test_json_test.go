package ciplanusecase

import "fmt"

func mustJSON(files []string) string {
	raw := "["
	for index, file := range files {
		if index != 0 {
			raw += ","
		}
		raw += fmt.Sprintf("%q", file)
	}
	return raw + "]"
}
