package main

import (
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "sort"
)

type feedback struct {
    SourceDigest            string `json:"source_digest"`
    DeclarationDigest       string `json:"declaration_digest"`
    CapabilityReceiptDigest string `json:"capability_receipt_digest"`
    EvidenceDigest          string `json:"evidence_digest"`
    SuggestedQuery          string `json:"suggested_query"`
    Outcome                 string `json:"outcome"`
    NonExecuting            bool   `json:"non_executing"`
    NonAuthorizing          bool   `json:"non_authorizing"`
}

type candidateKey struct {
    SourceDigest      string
    DeclarationDigest string
    SuggestedQuery    string
}

type candidate struct {
    SourceDigest       string   `json:"source_digest"`
    DeclarationDigest  string   `json:"declaration_digest"`
    SuggestedQuery     string   `json:"suggested_query"`
    ObservationCount   int      `json:"observation_count"`
    UsefulCount        int      `json:"useful_count"`
    NotUsefulCount     int      `json:"not_useful_count"`
    UnresolvedCount    int      `json:"unresolved_count"`
    EvidenceDigests   []string `json:"evidence_digests"`
    NextAction         string   `json:"next_action"`
}

type report struct {
    Version    string      `json:"version"`
    Candidates []candidate `json:"candidates"`
}

func (f feedback) validate() error {
    if f.SourceDigest == "" || f.DeclarationDigest == "" || f.CapabilityReceiptDigest == "" || f.EvidenceDigest == "" {
        return errors.New("feedback identity and evidence digests are required")
    }
    if f.SuggestedQuery == "" {
        return errors.New("suggested query is required")
    }
    switch f.Outcome {
    case "useful", "not_useful", "unresolved":
    default:
        return fmt.Errorf("unsupported outcome %q", f.Outcome)
    }
    if !f.NonExecuting || !f.NonAuthorizing {
        return errors.New("feedback must be non-executing and non-authorizing")
    }
    return nil
}

func decodeFeedback(r io.Reader) ([]feedback, error) {
    decoder := json.NewDecoder(r)
    decoder.DisallowUnknownFields()
    var records []feedback
    if err := decoder.Decode(&records); err != nil {
        return nil, fmt.Errorf("decode feedback: %w", err)
    }
    var trailing any
    if err := decoder.Decode(&trailing); err != io.EOF {
        if err == nil {
            return nil, errors.New("trailing JSON is not allowed")
        }
        return nil, fmt.Errorf("decode trailing JSON: %w", err)
    }
    return records, nil
}

func aggregate(records []feedback) ([]candidate, error) {
    grouped := make(map[candidateKey]*candidate)
    evidence := make(map[candidateKey]map[string]struct{})
    for index, record := range records {
        if err := record.validate(); err != nil {
            return nil, fmt.Errorf("feedback %d: %w", index, err)
        }
        key := candidateKey{
            SourceDigest:      record.SourceDigest,
            DeclarationDigest: record.DeclarationDigest,
            SuggestedQuery:    record.SuggestedQuery,
        }
        item := grouped[key]
        if item == nil {
            item = &candidate{
                SourceDigest:      key.SourceDigest,
                DeclarationDigest: key.DeclarationDigest,
                SuggestedQuery:    key.SuggestedQuery,
                NextAction:        "COLLECT_MORE_OBSERVATIONS",
            }
            grouped[key] = item
            evidence[key] = make(map[string]struct{})
        }
        item.ObservationCount++
        switch record.Outcome {
        case "useful":
            item.UsefulCount++
        case "not_useful":
            item.NotUsefulCount++
        case "unresolved":
            item.UnresolvedCount++
        }
        evidence[key][record.EvidenceDigest] = struct{}{}
    }

    result := make([]candidate, 0, len(grouped))
    for key, item := range grouped {
        for digest := range evidence[key] {
            item.EvidenceDigests = append(item.EvidenceDigests, digest)
        }
        sort.Strings(item.EvidenceDigests)
        if item.UsefulCount >= 2 && item.UnresolvedCount == 0 {
            item.NextAction = "REVIEW_INVESTMENT"
        }
        result = append(result, *item)
    }
    sort.Slice(result, func(i, j int) bool {
        if result[i].SourceDigest != result[j].SourceDigest {
            return result[i].SourceDigest < result[j].SourceDigest
        }
        if result[i].DeclarationDigest != result[j].DeclarationDigest {
            return result[i].DeclarationDigest < result[j].DeclarationDigest
        }
        return result[i].SuggestedQuery < result[j].SuggestedQuery
    })
    return result, nil
}

func main() {
    inputPath := flag.String("input", "-", "JSON feedback array path, or - for stdin")
    flag.Parse()

    var input io.Reader = os.Stdin
    var inputFile *os.File
    if *inputPath != "-" {
        var err error
        inputFile, err = os.Open(*inputPath)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        defer inputFile.Close()
        input = inputFile
    }

    records, err := decodeFeedback(input)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    candidates, err := aggregate(records)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    output := report{Version: "capability.feedback.investment.v1", Candidates: candidates}
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(output); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}