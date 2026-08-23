package languagegointeroperation

func go127Fixture(id string) ([]byte, bool) {
	source, found := go127Sources[id]
	return []byte(source), found
}

var go127Sources = map[string]string{
	"generic-method": `package interop
type Rand struct{}
func (Rand) N[Int ~int](value Int) Int { return value }
`,
	"generic-receiver-method": `package interop
type Box[T any] struct { Value T }
func (box Box[T]) Map[U any](convert func(T) U) U { return convert(box.Value) }
`,
	"generic-alias": `package interop
type Sequence[T any] = []T
func Keep[T any](value Sequence[T]) Sequence[T] { return value }
`,
	"assignment-inference": `package interop
func Identity[T any](value T) T { return value }
var StringIdentity func(string) string = Identity
`,
	"constrained-method": `package interop
type Number interface { ~int | ~int64 }
type Math struct{}
func (Math) Twice[N Number](value N) N { return value + value }
`,
	"generic-pair-method": `package interop
type Pair[A, B any] struct { First A; Second B }
type Pairer struct{}
func (Pairer) Pair[A, B any](a A, b B) Pair[A, B] { return Pair[A, B]{First: a, Second: b} }
`,
	"generic-codec-method": `package interop
type Codec struct{}
func (Codec) Decode[T ~string](value T) string { return string(value) }
`,
	"alias-to-generic-type": `package interop
type Cell[T any] struct { Value T }
type CellAlias[T any] = Cell[T]
func Wrap[T any](value T) CellAlias[T] { return Cell[T]{Value: value} }
`,
}

func guardrailFixture(id string) ([]byte, bool) {
	source, found := guardrailSources[id]
	return []byte(source), found
}

var guardrailSources = map[string]string{
	"parse-error":           "package interop\nfunc",
	"type-mismatch":         "package interop\nfunc Broken() int { return \"x\" }\n",
	"duplicate-declaration": "package interop\ntype Duplicate struct{}\ntype Duplicate int\n",
	"undefined-identifier":  "package interop\nfunc Broken() Missing { return Missing{} }\n",
	"import-authority":      "package interop\nimport \"fmt\"\nfunc Print() { fmt.Println() }\n",
	"unexported-api":        "package interop\ntype hidden struct{}\nfunc use(hidden) {}\n",
	"constraint-violation":  "package interop\nfunc NeedInt[T ~int](value T) T { return value }\nvar Broken = NeedInt(\"x\")\n",
}
