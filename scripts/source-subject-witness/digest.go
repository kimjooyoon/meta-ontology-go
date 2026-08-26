package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

func digestJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestValues[T any](values []T) string {
	rows := make([][]byte, 0, len(values))
	for _, value := range values {
		row, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return string(rows[i]) < string(rows[j]) })
	hash := sha256.New()
	for _, row := range rows {
		hash.Write(row)
		hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func operationSet(rows []sourceIndicator) string {
	seen := make(map[string]bool)
	operations := make([]string, 0)
	for _, row := range rows {
		if !seen[row.MetaOperation] {
			seen[row.MetaOperation] = true
			operations = append(operations, row.MetaOperation)
		}
	}
	sort.Strings(operations)
	return strings.Join(operations, "+")
}
