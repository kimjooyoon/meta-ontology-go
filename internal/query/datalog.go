package query

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	DefaultDatalogLimit        = 100
	MaxDatalogLimit            = MaxEnvelopeLimit
	DefaultDatalogDerivedLimit = 10000
	MaxDatalogRules            = 128
	MaxDatalogBodyAtoms        = 32
)

var (
	ErrInvalidDatalogQuery = errors.New("invalid Datalog query")
	ErrDatalogBudget       = errors.New("Datalog evaluation budget exceeded")
)

// DatalogTerm is either a stable ID constant or a variable. Variables are
// canonicalized with a leading '?' but callers may pass either "name" or
// "?name" to Variable.
type DatalogTerm struct {
	Variable string
	Constant ID
}

// Term is the concise spelling used by parser-neutral callers.
type Term = DatalogTerm

func Variable(name string) DatalogTerm { return DatalogTerm{Variable: name} }

func Constant(id ID) DatalogTerm { return DatalogTerm{Constant: id} }

// DatalogAtom is a positive binary triple pattern. Negation, aggregation, and
// unbounded path operators are intentionally not part of this query boundary.
type DatalogAtom struct {
	Predicate string
	Subject   DatalogTerm
	Object    DatalogTerm
}

type Atom = DatalogAtom

func Triple(predicate string, subject, object DatalogTerm) DatalogAtom {
	return DatalogAtom{Predicate: predicate, Subject: subject, Object: object}
}

// DatalogRule is a positive implication. Every variable in Head must occur in
// Body, which makes rule evaluation finite over the selected fact universe.
type DatalogRule struct {
	ID   string
	Head DatalogAtom
	Body []DatalogAtom
}

type Rule = DatalogRule

// DatalogQuery evaluates a conjunction of patterns. Candidates are visible to
// matching only when IncludeCandidates is true; they are never rule inputs.
type DatalogQuery struct {
	Patterns          []DatalogAtom
	Rules             []DatalogRule
	IncludeCandidates bool
	IncludeDerived    bool
	Limit             int
	MaxDerivedFacts   int
}

// DatalogOrigin identifies the authority boundary of a returned fact.
type DatalogOrigin uint8

const (
	DatalogDeclared DatalogOrigin = iota + 1
	DatalogCandidate
	DatalogDerived
)

func (origin DatalogOrigin) String() string {
	switch origin {
	case DatalogDeclared:
		return "declared"
	case DatalogCandidate:
		return "candidate"
	case DatalogDerived:
		return "derived"
	default:
		return "unknown"
	}
}

// DatalogFact is a read-only fact in the selected query universe. Derived
// facts carry one deterministic support proof; they never enter Graph.
type DatalogFact struct {
	Subject   ID
	Predicate string
	Object    ID
	Origin    DatalogOrigin
	RuleID    string
	Support   []DatalogFactKey
}

type DatalogFactKey struct {
	Subject   ID
	Predicate string
	Object    ID
}

func (fact DatalogFact) Key() DatalogFactKey {
	return DatalogFactKey{fact.Subject, fact.Predicate, fact.Object}
}

// DatalogRow contains one set-semantics binding and the facts that satisfied
// the query patterns. Binding names omit the canonical '?' prefix.
type DatalogRow struct {
	Bindings map[string]ID
	Facts    []DatalogFact
}

func (row DatalogRow) Value(name string) (ID, bool) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "?")
	value, ok := row.Bindings[name]
	return value, ok
}

// DatalogResult is deterministic by construction. Complete is false when the
// result limit trimmed rows; no partial result is ever reported as complete.
type DatalogResult struct {
	Rows     []DatalogRow
	Derived  []DatalogFact
	Complete bool
}

// EvaluateDatalog evaluates positive rules over deterministic graph facts and
// then matches the requested patterns. It is a read-only projection and does
// not promote candidates or alter the graph hash.
func (graph Graph) EvaluateDatalog(request DatalogQuery) (DatalogResult, error) {
	normalized, rules, err := normalizeDatalogQuery(request)
	if err != nil {
		return DatalogResult{}, err
	}

	declared := make([]DatalogFact, 0, len(graph.DeterministicFacts()))
	for _, fact := range graph.DeterministicFacts() {
		declared = append(declared, DatalogFact{
			Subject: fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
			Origin: DatalogDeclared,
		})
	}
	sortDatalogFacts(declared)
	var derived []DatalogFact
	if normalized.IncludeDerived {
		derived, err = deriveDatalog(declared, rules, normalized.MaxDerivedFacts)
		if err != nil {
			return DatalogResult{}, err
		}
	}

	universe := append([]DatalogFact(nil), declared...)
	if normalized.IncludeDerived {
		universe = append(universe, derived...)
	}
	if normalized.IncludeCandidates {
		for _, fact := range graph.CandidateFacts() {
			universe = append(universe, DatalogFact{
				Subject: fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
				Origin: DatalogCandidate,
			})
		}
	}
	sortDatalogFacts(universe)
	rows := matchDatalogPatterns(normalized.Patterns, universe)
	complete := true
	if len(rows) > normalized.Limit {
		rows = rows[:normalized.Limit]
		complete = false
	}
	return DatalogResult{Rows: rows, Derived: derived, Complete: complete}, nil
}

// QueryDatalog is an API synonym that reads naturally at call sites.
func (graph Graph) QueryDatalog(request DatalogQuery) (DatalogResult, error) {
	return graph.EvaluateDatalog(request)
}

func normalizeDatalogQuery(request DatalogQuery) (DatalogQuery, []DatalogRule, error) {
	if len(request.Patterns) == 0 || len(request.Patterns) > MaxDatalogBodyAtoms {
		return DatalogQuery{}, nil, datalogError("query requires 1..%d patterns", MaxDatalogBodyAtoms)
	}
	if len(request.Rules) > MaxDatalogRules {
		return DatalogQuery{}, nil, datalogError("rule count exceeds %d", MaxDatalogRules)
	}
	if request.Limit < 0 || request.Limit > MaxDatalogLimit {
		return DatalogQuery{}, nil, datalogError("limit must be 0..%d", MaxDatalogLimit)
	}
	if request.Limit == 0 {
		request.Limit = DefaultDatalogLimit
	}
	if request.MaxDerivedFacts < 0 || request.MaxDerivedFacts > DefaultDatalogDerivedLimit {
		return DatalogQuery{}, nil, datalogError("max derived facts must be 0..%d", DefaultDatalogDerivedLimit)
	}
	if request.MaxDerivedFacts == 0 {
		request.MaxDerivedFacts = DefaultDatalogDerivedLimit
	}

	patterns := make([]DatalogAtom, len(request.Patterns))
	for index, pattern := range request.Patterns {
		var err error
		patterns[index], err = normalizeDatalogAtom(pattern)
		if err != nil {
			return DatalogQuery{}, nil, err
		}
	}
	rules := make([]DatalogRule, len(request.Rules))
	knownHeads := make(map[string]struct{}, len(request.Rules))
	seenRules := make(map[string]struct{}, len(request.Rules))
	for index, rule := range request.Rules {
		normalized, err := normalizeDatalogRule(rule)
		if err != nil {
			return DatalogQuery{}, nil, err
		}
		if _, exists := seenRules[normalized.ID]; exists {
			return DatalogQuery{}, nil, datalogError("duplicate rule ID %q", normalized.ID)
		}
		seenRules[normalized.ID] = struct{}{}
		knownHeads[normalized.Head.Predicate] = struct{}{}
		rules[index] = normalized
	}
	known := map[string]struct{}{
		string(Used): {}, string(WasGeneratedBy): {}, string(WasDerivedFrom): {}, string(WasAssociatedWith): {},
	}
	for predicate := range knownHeads {
		known[predicate] = struct{}{}
	}
	for _, rule := range rules {
		for _, atom := range rule.Body {
			if _, exists := known[atom.Predicate]; !exists {
				return DatalogQuery{}, nil, datalogError("rule %q references unknown predicate %q", rule.ID, atom.Predicate)
			}
		}
	}
	for _, pattern := range patterns {
		if _, exists := known[pattern.Predicate]; !exists {
			return DatalogQuery{}, nil, datalogError("query references unknown predicate %q", pattern.Predicate)
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	request.Patterns, request.Rules = patterns, rules
	return request, rules, nil
}

func normalizeDatalogRule(rule DatalogRule) (DatalogRule, error) {
	rule.ID = strings.TrimSpace(rule.ID)
	if !validDatalogRuleID(rule.ID) {
		return DatalogRule{}, datalogError("rule ID %q is invalid", rule.ID)
	}
	if len(rule.Body) == 0 || len(rule.Body) > MaxDatalogBodyAtoms {
		return DatalogRule{}, datalogError("rule %q requires 1..%d body atoms", rule.ID, MaxDatalogBodyAtoms)
	}
	head, err := normalizeDatalogAtom(rule.Head)
	if err != nil {
		return DatalogRule{}, datalogError("rule %q head: %v", rule.ID, err)
	}
	body := make([]DatalogAtom, len(rule.Body))
	bound := make(map[string]struct{})
	for index, atom := range rule.Body {
		body[index], err = normalizeDatalogAtom(atom)
		if err != nil {
			return DatalogRule{}, datalogError("rule %q body atom %d: %v", rule.ID, index, err)
		}
		for _, term := range []DatalogTerm{body[index].Subject, body[index].Object} {
			if term.Variable != "" {
				bound[strings.TrimPrefix(term.Variable, "?")] = struct{}{}
			}
		}
	}
	for _, term := range []DatalogTerm{head.Subject, head.Object} {
		if term.Variable != "" {
			if _, exists := bound[strings.TrimPrefix(term.Variable, "?")]; !exists {
				return DatalogRule{}, datalogError("rule %q has unsafe head variable %q", rule.ID, term.Variable)
			}
		}
	}
	rule.Head, rule.Body = head, body
	return rule, nil
}

func normalizeDatalogAtom(atom DatalogAtom) (DatalogAtom, error) {
	predicate, err := normalizeDatalogPredicate(atom.Predicate)
	if err != nil {
		return DatalogAtom{}, err
	}
	subject, err := normalizeDatalogTerm(atom.Subject)
	if err != nil {
		return DatalogAtom{}, datalogError("subject: %v", err)
	}
	object, err := normalizeDatalogTerm(atom.Object)
	if err != nil {
		return DatalogAtom{}, datalogError("object: %v", err)
	}
	return DatalogAtom{Predicate: predicate, Subject: subject, Object: object}, nil
}

func normalizeDatalogTerm(term DatalogTerm) (DatalogTerm, error) {
	if term.Variable != "" && term.Constant != "" {
		return DatalogTerm{}, datalogError("term cannot be both variable and constant")
	}
	if term.Variable != "" {
		name := strings.TrimPrefix(strings.TrimSpace(term.Variable), "?")
		if !validDatalogIdentifier(name) {
			return DatalogTerm{}, datalogError("variable %q is invalid", term.Variable)
		}
		return DatalogTerm{Variable: "?" + name}, nil
	}
	if term.Constant == "" {
		return DatalogTerm{}, datalogError("term is empty")
	}
	id, err := ParseID(term.Constant.String())
	if err != nil {
		return DatalogTerm{}, datalogError("constant: %v", err)
	}
	return DatalogTerm{Constant: id}, nil
}

func normalizeDatalogPredicate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if relation, err := ParseRelation(Relation(raw)); err == nil {
		return string(relation), nil
	}
	if !validDatalogIdentifier(raw) {
		return "", datalogError("predicate %q is invalid", raw)
	}
	return raw, nil
}

func validDatalogIdentifier(value string) bool {
	if value == "" || !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z') || value[0] == '_') {
		return false
	}
	for _, character := range value[1:] {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func validDatalogRuleID(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !(character == '/' || character == ':' || character == '.' || character == '-' ||
			(character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}

func datalogError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidDatalogQuery, fmt.Sprintf(format, args...))
}

type datalogBinding map[string]ID

func deriveDatalog(base []DatalogFact, rules []DatalogRule, limit int) ([]DatalogFact, error) {
	byPredicate := make(map[string][]DatalogFact)
	known := make(map[DatalogFactKey]struct{})
	for _, fact := range base {
		byPredicate[fact.Predicate] = append(byPredicate[fact.Predicate], fact)
		known[fact.Key()] = struct{}{}
	}
	for predicate := range byPredicate {
		sortDatalogFacts(byPredicate[predicate])
	}
	derived := make([]DatalogFact, 0)
	for {
		changed := false
		for _, rule := range rules {
			for _, conclusion := range applyDatalogRule(rule, byPredicate) {
				if _, exists := known[conclusion.Key()]; exists {
					continue
				}
				if len(derived) >= limit {
					return nil, fmt.Errorf("%w: maximum derived facts %d", ErrDatalogBudget, limit)
				}
				known[conclusion.Key()] = struct{}{}
				derived = append(derived, conclusion)
				byPredicate[conclusion.Predicate] = append(byPredicate[conclusion.Predicate], conclusion)
				sortDatalogFacts(byPredicate[conclusion.Predicate])
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	sortDatalogFacts(derived)
	return derived, nil
}

func applyDatalogRule(rule DatalogRule, byPredicate map[string][]DatalogFact) []DatalogFact {
	results := make([]DatalogFact, 0)
	var join func(int, datalogBinding, []DatalogFact)
	join = func(index int, binding datalogBinding, support []DatalogFact) {
		if index == len(rule.Body) {
			subject := datalogTermValue(rule.Head.Subject, binding)
			object := datalogTermValue(rule.Head.Object, binding)
			results = append(results, DatalogFact{
				Subject: subject, Predicate: rule.Head.Predicate, Object: object,
				Origin: DatalogDerived, RuleID: rule.ID, Support: datalogSupport(support),
			})
			return
		}
		atom := rule.Body[index]
		for _, fact := range byPredicate[atom.Predicate] {
			next, ok := bindDatalogAtom(atom, fact, binding)
			if !ok {
				continue
			}
			join(index+1, next, append(support, fact))
		}
	}
	join(0, make(datalogBinding), nil)
	return results
}

func bindDatalogAtom(atom DatalogAtom, fact DatalogFact, binding datalogBinding) (datalogBinding, bool) {
	next := make(datalogBinding, len(binding)+2)
	for name, value := range binding {
		next[name] = value
	}
	for _, pair := range [][2]DatalogTerm{{atom.Subject, {Constant: fact.Subject}}, {atom.Object, {Constant: fact.Object}}} {
		if pair[0].Variable != "" {
			name := strings.TrimPrefix(pair[0].Variable, "?")
			if existing, exists := next[name]; exists && existing != pair[1].Constant {
				return nil, false
			}
			next[name] = pair[1].Constant
		} else if pair[0].Constant != pair[1].Constant {
			return nil, false
		}
	}
	return next, true
}

func datalogTermValue(term DatalogTerm, binding datalogBinding) ID {
	if term.Variable == "" {
		return term.Constant
	}
	return binding[strings.TrimPrefix(term.Variable, "?")]
}

func datalogSupport(support []DatalogFact) []DatalogFactKey {
	keys := make([]DatalogFactKey, len(support))
	for index, fact := range support {
		keys[index] = fact.Key()
	}
	return keys
}

func matchDatalogPatterns(patterns []DatalogAtom, facts []DatalogFact) []DatalogRow {
	byPredicate := make(map[string][]DatalogFact)
	for _, fact := range facts {
		byPredicate[fact.Predicate] = append(byPredicate[fact.Predicate], fact)
	}
	for predicate := range byPredicate {
		sortDatalogFacts(byPredicate[predicate])
	}
	rows := make([]DatalogRow, 0)
	seen := make(map[string]struct{})
	var join func(int, datalogBinding, []DatalogFact)
	join = func(index int, binding datalogBinding, matched []DatalogFact) {
		if index == len(patterns) {
			row := DatalogRow{Bindings: copyDatalogBinding(binding), Facts: append([]DatalogFact(nil), matched...)}
			key := datalogBindingCanonical(binding)
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				rows = append(rows, row)
			}
			return
		}
		pattern := patterns[index]
		for _, fact := range byPredicate[pattern.Predicate] {
			next, ok := bindDatalogAtom(pattern, fact, binding)
			if ok {
				join(index+1, next, append(matched, fact))
			}
		}
	}
	join(0, make(datalogBinding), nil)
	sort.Slice(rows, func(i, j int) bool { return datalogRowCanonical(rows[i]) < datalogRowCanonical(rows[j]) })
	return rows
}

func copyDatalogBinding(binding datalogBinding) map[string]ID {
	copy := make(map[string]ID, len(binding))
	for name, value := range binding {
		copy[name] = value
	}
	return copy
}

func datalogBindingCanonical(binding datalogBinding) string {
	names := make([]string, 0, len(binding))
	for name := range binding {
		names = append(names, name)
	}
	sort.Strings(names)
	var builder strings.Builder
	for _, name := range names {
		builder.WriteString(strconv.Quote(name))
		builder.WriteByte('=')
		builder.WriteString(strconv.Quote(binding[name].String()))
		builder.WriteByte(';')
	}
	return builder.String()
}

func datalogRowCanonical(row DatalogRow) string {
	var builder strings.Builder
	builder.WriteString(datalogBindingCanonical(row.Bindings))
	for _, fact := range row.Facts {
		builder.WriteString(strconv.Quote(fact.Key().String()))
	}
	return builder.String()
}

func sortDatalogFacts(facts []DatalogFact) {
	sort.Slice(facts, func(i, j int) bool {
		left, right := facts[i], facts[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.RuleID < right.RuleID
	})
}

func (key DatalogFactKey) String() string {
	return key.Subject.String() + "\x00" + key.Predicate + "\x00" + key.Object.String()
}

// Canonical provides a stable receipt for replay and permutation tests.
func (result DatalogResult) Canonical() string {
	var builder strings.Builder
	for _, row := range result.Rows {
		builder.WriteString(datalogRowCanonical(row))
		builder.WriteByte('\n')
	}
	for _, fact := range result.Derived {
		builder.WriteString(fact.Key().String())
		builder.WriteByte('\t')
		builder.WriteString(fact.RuleID)
		builder.WriteByte('\n')
	}
	if result.Complete {
		builder.WriteString("complete\n")
	} else {
		builder.WriteString("incomplete\n")
	}
	return builder.String()
}

func (result DatalogResult) StableHash() string {
	digest := sha256.Sum256([]byte(result.Canonical()))
	return hex.EncodeToString(digest[:])
}
