package phaseseparation

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type parseFault struct {
	Step   string
	Reason string
}

func (f parseFault) Error() string {
	return fmt.Sprintf("%s: %s", f.Step, f.Reason)
}

func Parse(input []byte) (Source, error) {
	var result Source
	var current *Case
	seenHeader := false
	seenFields := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if !seenHeader {
			if line != "phase-witness v1" {
				return Source{}, parseFault{"parse-header", "UNKNOWN_SOURCE_SYNTAX"}
			}
			seenHeader = true
			continue
		}
		if fields[0] == "case" {
			if len(fields) != 2 || fields[1] == "" {
				return Source{}, parseFault{"parse-case", "UNKNOWN_SOURCE_SYNTAX"}
			}
			result.Cases = append(result.Cases, Case{Name: fields[1]})
			current = &result.Cases[len(result.Cases)-1]
			continue
		}
		if current == nil {
			if len(fields) != 2 || !isHeaderField(fields[0]) || fields[1] == "" || seenFields[fields[0]] {
				return Source{}, parseFault{"parse-header-field", "UNKNOWN_SOURCE_SYNTAX"}
			}
			seenFields[fields[0]] = true
			switch fields[0] {
			case "producer":
				result.Producer = fields[1]
			case "consumer":
				result.Consumer = fields[1]
			case "meta-operation":
				result.MetaOperation = fields[1]
			case "proof-choice":
				result.ProofChoice = fields[1]
			}
			continue
		}
		switch fields[0] {
		case "value":
			value, err := parseValue(line)
			if err != nil {
				return Source{}, parseFault{fmt.Sprintf("parse-value-line-%d", lineNumber), "UNKNOWN_SOURCE_SYNTAX"}
			}
			current.Values = append(current.Values, value)
		case "transfer":
			transfer, err := parseTransfer(line)
			if err != nil {
				return Source{}, parseFault{fmt.Sprintf("parse-transfer-line-%d", lineNumber), "UNKNOWN_SOURCE_SYNTAX"}
			}
			current.Transfers = append(current.Transfers, transfer)
		default:
			return Source{}, parseFault{fmt.Sprintf("parse-record-line-%d", lineNumber), "UNKNOWN_SOURCE_SYNTAX"}
		}
	}
	if err := scanner.Err(); err != nil {
		return Source{}, parseFault{"read-source", "UNKNOWN_SOURCE_SYNTAX"}
	}
	if !seenHeader || len(result.Cases) == 0 || !completeHeader(result) {
		return Source{}, parseFault{"validate-source", "UNKNOWN_SOURCE_CONTRACT"}
	}
	return result, nil
}

func isHeaderField(field string) bool {
	return field == "producer" || field == "consumer" || field == "meta-operation" || field == "proof-choice"
}

func completeHeader(source Source) bool {
	return source.Producer != "" && source.Consumer != "" && source.MetaOperation != "" && source.ProofChoice != ""
}

func parseValue(line string) (Value, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "value"))
	parts := strings.SplitN(rest, " ", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" {
		return Value{}, fmt.Errorf("value shape")
	}
	literal, err := strconv.Unquote(strings.TrimSpace(parts[2]))
	if err != nil {
		return Value{}, err
	}
	return Value{Phase: parts[0], ID: parts[1], Literal: literal}, nil
}

func parseTransfer(line string) (Transfer, error) {
	fields := strings.Fields(line)
	if len(fields) != 6 {
		return Transfer{}, fmt.Errorf("transfer shape")
	}
	if !validPhase(fields[1]) || !validPhase(fields[3]) || fields[2] == "" || fields[4] == "" || fields[5] == "" {
		return Transfer{}, fmt.Errorf("transfer values")
	}
	return Transfer{FromPhase: fields[1], FromID: fields[2], ToPhase: fields[3], ToID: fields[4], Kind: fields[5]}, nil
}

func validPhase(phase string) bool {
	for _, candidate := range phases {
		if phase == candidate {
			return true
		}
	}
	return false
}
