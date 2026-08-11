# zerolang 연구: graph-first 모델과 폐쇄루프 Agent workflow

> 조사 시점: 2026-08-12 (Asia/Seoul)
> 공식 저장소 snapshot: `vercel-labs/zerolang` `afcc72da649fe4d4c670ac1489c2197d37436051`
> 범위: graph-first 모델, projection/patch/query/check/test/run/import/export, conformance,
> Agent workflow, 그리고 현재 `meta-ontology-go` open PR review checklist

이 문서는 zerolang의 구현을 복제하자는 제안이 아니다. 현재 공식 저장소에서 실제로
검증되는 경계와 실패 조건을 추출하고, `.gooo`의 Business DSL·Semantic IR·Go
projection이 그 기준을 어떻게 넘어서야 하는지 제안한다. zerolang 자체도 실험 단계이며
보안 취약점과 breaking change를 전제로 한다.

## 요약

zerolang의 핵심은 “텍스트를 읽고 추측한 뒤 별도 도구로 고치는” Agent loop를
“graph를 조회하고, 확인한 graph 상태에 묶인 patch를 제출하고, compiler가 검증한 뒤
필요한 검증만 실행하는” loop로 바꾸는 것이다.

- graph-first 패키지의 compiler input은 `zero.graph`다. `.0`은 사람이 읽고 검토하는
  projection이며 일반적인 Agent authoring surface가 아니다.
- `zero query`/`inspect`는 node ID, graph hash, type/effect/ownership/capability,
  import/call edge, target fact를 좁은 범위로 제공한다.
- `zero patch`는 graph hash, node hash, expected field value, typed operation을
  precondition으로 받을 수 있고, stale·shape·type 오류를 store write 전에 거부한다.
- `zero import`, `zero export`, `zero verify-projection`은 graph와 projection의
  동기화를 명시적인 workflow로 만든다. 양쪽이 독립적으로 바뀌면 한쪽을 임의로
  선택하지 않고 충돌을 보고한다.
- conformance는 언어 fixture 하나가 아니라 graph input policy, provenance guardrail,
  graph smoke/parity, canonical text, examples gate, command contract, run fixture를
  함께 묶은다.

meta-ontology-go가 가져올 것은 graph를 유일한 업무 SSOT로 만드는 정책이 아니라,
checked semantic edit·명시적 projection sync·결정적 conformance·repair-oriented
diagnostic이라는 품질 기준이다. `.gooo` 선언은 Business intent의 권위로 유지하고,
Semantic IR은 정규화된 의미의 권위, handwritten Go는 환원 불가능한 구현의 권위로
분리해야 한다.

## 1. Graph-first 모델

### 1.1 관찰된 구조

| 층 | zerolang의 역할 | meta-ontology-go에 적용할 때의 해석 |
| --- | --- | --- |
| Graph store | `zero.graph`가 graph-first package의 checked compiler input | IR/Go projection보다 먼저 어떤 view가 compiler input인지 명시한다. `.gooo` business intent의 권위와 혼동하지 않는다. |
| Node handle | node ID는 조회한 graph의 편집 대상이다 | 안정된 semantic ID와 compiler-local node handle을 분리한다. handle은 stale edit 방지용이고 URI-like ID가 identity다. |
| Graph hash | Agent가 확인한 상태의 optimistic precondition | patch뿐 아니라 semantic delta reconciliation에도 inspected IR fingerprint를 요구한다. |
| Node/field hash | 동일 ID라도 내용이 바뀌었는지 확인 | Go symbol lifting은 symbol ID와 source span/expected fact를 함께 확인한다. |
| Typed facts | type, effect, ownership, capability, imports, calls, target facts | PROV fact와 implementation fact를 구분하고, 구현 detail을 business relation으로 자동 승격하지 않는다. |
| `.0` projection | 사람이 읽고 review하는 text view | `.gooo`는 business source, generated Go는 structural projection으로 권위를 별도로 선언한다. |
| Source map | graph node와 읽기 쉬운 source 위치 연결 | DSL span, IR fact, generated Go marker, handwritten slot을 모두 추적한다. |

graph-first라는 말은 “graph가 모든 의미의 SSOT”라는 뜻이 아니다. zerolang의 현재
문서도 graph를 compiler database로 쓰면서 `.0`을 human escape hatch로 남긴다. 이를
meta-ontology-go에 옮길 때는 다음 authority boundary를 지켜야 한다.

```text
Business DSL (.gooo) -> Semantic IR -> generated Go / query / docs / CI evidence
                          ^                    |
                          |                    v
                    registered Go symbols <- semantic delta
```

Go 분석 결과는 source-backed deterministic fact일 때만 IR delta가 될 수 있다. 일반
helper 호출은 call graph의 implementation detail이며, semantic namespace에 등록된
symbol을 참조했다는 이유만으로 `requires`, `delegatesTo`, `wasDerivedFrom` 같은
domain relation을 추론해서는 안 된다. 애매한 관찰은 candidate/evidence로 남긴다.

### 1.2 Locality와 identity

zerolang의 parity fixture는 local literal edit에서 graph hash와 node hash는 바뀌지만
해당 node ID는 유지되는지, declaration rename에서 unambiguous한 경우 ID가 유지되는지,
동일 shape가 충돌하면 어느 ID가 retire되는지를 확인한다. 여기서 중요한 교훈은
compiler node ID의 안정성은 조건부라는 점이다.

meta-ontology-go의 안정성 기준은 더 강해야 한다.

1. semantic ID(`billing://activity/pay-order`)는 display name/alias rename과 무관하게
   유지된다.
2. namespace가 다르면 같은 이름을 merge하지 않는다.
3. semantic ID와 source span이 없으면 semantic delta를 accept하지 않는다.
4. implementation-only Go edit는 unrelated declaration, generated region, handwritten
   slot을 다시 쓰지 않는다.
5. ambiguous match는 임의로 기존 identity를 빼앗지 않고 conflict/candidate로 보고한다.

## 2. 명령 surface별 계약

| surface | 공식 workflow에서 하는 일 | meta-ontology-go가 넘어야 할 acceptance |
| --- | --- | --- |
| `zero query` / `zero inspect` | `--fn`, `--find`, `--refs`, `--calls`로 필요한 graph slice만 조회한다. 기본 text와 machine용 JSON을 분리한다. | semantic ID, relation, source span, graph/IR fingerprint를 deterministic order로 조회한다. query가 없는 fact를 발명하지 않고, JSON field를 contract test로 고정한다. |
| `zero patch` | graph node/field/function body를 typed op로 수정한다. `--check-only`, `--dry-run`, graph hash, node hash, expected value를 지원한다. | stale IR/Go snapshot, mismatched expected value, invalid relation shape를 write 전에 거부한다. 실패는 transactional이어야 하고 partial graph/generated output이 남지 않아야 한다. |
| `zero check` | graph-backed input을 resolve/type/readiness 관점에서 확인하고, `--json`으로 code/span/expected/actual/repair 정보를 준다. | `.gooo -> IR`, IR validation, registered Go fact lifting, generated-region integrity를 하나의 deterministic check로 연결한다. candidate와 accepted truth를 구분한다. |
| `zero test` | graph-backed test block을 발견하고 `--filter`와 JSON 결과를 제공한다. expected-fail은 unexpected pass도 실패로 만든다. | declaration/IR/Go projection/test harness가 같은 semantic ID를 가리키는 fixture를 실행한다. expected failure, discovered/selected count, per-test evidence를 보존한다. |
| `zero run` | 현재 graph input으로 실행하고 `--` 뒤의 program args를 전달한다. examples gate는 exit code와 output/server probe까지 확인한다. | canonical billing example을 check→test→run까지 연결한다. input, output, exit status와 provenance/evidence를 build artifact에 연결한다. |
| `zero export` | graph store에서 최신 `.0` projection을 쓴다. 사람 review용이며 Agent의 기본 edit가 아니다. | IR에서 Go를 생성할 때 stable generated markers와 semantic source map을 만든다. 생성 영역 밖 handwritten text와 slot은 보존한다. |
| `zero import` | 사람이 의도적으로 바꾼 `.0`을 graph store로 재구성한다. format/target을 명시할 수 있다. | `.gooo`/Go 역방향 변환에서 representable fact만 accept한다. source span 없는 fact, ambiguous relation, identity 충돌은 reject/candidate로 남긴다. |
| `zero verify-projection` | write 없이 projection drift를 검사한다. graph가 최신인지, text가 최신인지, 양쪽이 독립 변경됐는지를 content hash로 판단한다. | generated Go freshness와 DSL/IR/Go semantic equivalence를 write 없이 확인한다. 양쪽 변경 시 자동 overwrite하지 않고 conflict와 repair 경로를 제시한다. |
| `zero status` / `zero explain` / `zero fix --plan` | projection state와 repair-oriented diagnostic을 제공한다. fix plan은 자동 수정 자체가 아니다. | 모든 실패가 “무엇이 기대되었고 무엇이 관찰되었으며 다음 안전한 action이 무엇인지”를 말해야 한다. Guardian이 재현 가능한 evidence를 얻어야 한다. |

projection의 핵심은 양방향 텍스트 편집 자체가 아니라 authority와 sync 시점을 숨기지
않는 것이다. meta-ontology-go에서는 `.gooo`가 business intent의 권위이므로
`import`/`export`가 Go 구현 detail을 업무 선언으로 조용히 덮어쓰지 않도록 더 엄격한
reconciliation policy가 필요하다.

## 3. Agent workflow

공식 workflow를 meta-ontology-go의 언어로 옮기면 다음과 같다.

```text
request
  -> version-matched skill/rules 읽기
  -> semantic query로 좁은 context와 fingerprint 확보
  -> check-only patch / semantic delta 계획
  -> precondition을 만족할 때만 적용
  -> check -> focused test -> run/build
  -> evidence와 semantic delta 기록
  -> 사람은 projection과 diff를 검토
```

### 읽기 경로

- compiler가 제공하는 version-matched skill/rules를 먼저 읽어 현재 문법과 repair
  contract를 맞춘다.
- 전체 source를 무작정 context에 넣기보다 symbol, diagnostic, call, capability,
  module, node ID 중심으로 query한다.
- JSON은 exact field가 필요한 editor/CI/Agent에만 사용하고, 사람 review는 concise text와
  projection을 사용한다.

### 쓰기 경로

- text search-and-replace 대신 semantic target, expected value, inspected fingerprint를
  patch에 넣는다.
- patch가 적용되었다는 사실만으로 user-visible behavior를 증명했다고 간주하지 않고,
  필요한 다음 check/test/run을 선택한다.
- projection은 사람이 요청했을 때 export하고, 사람이 직접 수정했을 때만 import한다.
- diagnosis의 repair ID는 Agent가 다음 query/patch를 선택하는 입력이지 자동 승인 토큰이
  아니다.

### 안전 경계

zerolang의 `AGENTS.md`는 compiler가 실험 단계임을 명시하고 isolated/disposable
workspace를 요구한다. meta-ontology-go는 여기에 Builder/Guardian/Gate 분리를 더한다.
Builder는 `.gooo`와 handwritten slot을 바꿀 수 있지만 ontology, verifier semantics,
CI policy를 바꾸지 못한다. Guardian은 semantic delta·evidence·freshness를 검증하지만
기능을 구현하지 않는다. Gate는 deterministic policy만 실행한다.

## 4. Conformance와 closed loop

### 4.1 Fixture taxonomy

공식 저장소는 다음처럼 실패 양상과 compiler 경계를 분리한다.

| 영역 | fixture/검증 대상 | meta-ontology-go의 대응 |
| --- | --- | --- |
| check | `conformance/check/pass`, `common/pass`, `common/fail` | valid/invalid `.gooo`, namespace, ID, relation, source span, error code fixture |
| projection/format | `conformance/format`, `.0`와 `.graph` sibling | DSL/IR/Go canonical text, marker integrity, round-trip/locality golden |
| program graph | `conformance/program-graph`, graph store validity, import/export | semantic IR store, canonical hash, atomic write, stale/conflict import/export |
| agent surface | `conformance/agent-surface`, classification metadata | deterministic fact/candidate/implementation classification과 allowed scope fixture |
| package/import | `conformance/packages`, manifest와 module edge | namespace-safe cross-package semantic references와 dependency provenance |
| run | `conformance/run/pass`, exit/output | `check → test → run` end-to-end billing example과 evidence receipt |
| diagnostics | `conformance/diagnostics` expected JSON | code, span, expected/actual, repair ID, candidate-vs-truth contract |
| reliability | golden, snapshot, fuzz, crasher regression | cache corruption, malformed DSL, generated marker corruption, evidence freshness |

### 4.2 Harness와 CI depth

현재 snapshot의 `scripts/validation-suite.mts` conformance phase는 다음을 포함한다.

```text
native-build
graph-input-policy
native-contracts
provenance-guardrails
type-core-smoke
mir-verifier-smoke
program-graph-smoke
program-graph-parity
canonical-text-smoke
examples-gate
conformance-run
```

`program-graph-smoke`는 checked-in store가 binary인지, projection에 graph authority가
있는지, status/check/run/verify-projection이 clean한지, corrupt store가 `RGP003`으로
거부되는지, atomic temp file이 남지 않는지를 확인한다. `program-graph-parity`는 source와
graph artifact의 실행 결과 일치, graph/node hash, patch precondition, node ID locality,
reconcile와 source map을 확인한다. 이것이 단순 unit test보다 중요한 이유는 graph-first
compiler의 실제 authority 경계를 테스트하기 때문이다.

CI는 PR fast path와 schedule/manual deep path를 분리한다. `agent:checks`는 isolated
`/tmp` workspace에서 conformance, native shards, sanitizer, command contracts,
workspace checks를 실행한다. meta-ontology-go는 외부 sandbox가 필요하지 않으므로 같은
구조를 Go 표준 라이브러리만으로 재현하되, 각 job이 동일한 input hash와 evidence schema를
사용하게 해야 한다.

### 4.3 Closed-loop의 최소 증거

```text
allowed intent/scope
  -> actual semantic delta
  -> projection/generated output
  -> check/test/run/build evidence
  -> fresh provenance receipt
  -> Guardian review
  -> deterministic Gate decision
```

실제 delta가 allowed scope를 벗어나면 파일 overlap이 없더라도 실패해야 한다. 반대로
같은 semantic ID의 implementation-only 변경은 unrelated region을 재생성하지 않아야 한다.
검증 결과는 append-only evidence로 남기고, evidence가 현재 source/IR/Go hash와 맞지 않으면
stale로 실패시킨다.

## 5. meta-ontology-go가 초과해야 할 품질 기준과 수용 테스트

다음 표는 zerolang에서 관찰한 기준을 그대로 복사하지 않고, `.gooo`의 PROV-inspired
IR·Go projection·provenance 요구를 추가한 제안이다. 각 행은 “통과했다”가 아니라
재현 가능한 acceptance test를 가져야 한다.

| 기준 | 최소 수용 테스트 | zerolang 대비 초과점 |
| --- | --- | --- |
| 권위 경계 | `.gooo` 선언을 변경하면 IR/Go/docs/query가 갱신되고, generated Go만 편집해서 business declaration이 생기지 않는다. | graph authority를 business authority와 분리하고 역방향 추론의 승격 조건을 명시한다. |
| 안정 identity | display rename, alias 추가, namespace 동명이인 fixture에서 stable semantic ID가 유지되고 merge되지 않는다. | compiler node handle보다 강한 URI-like semantic identity를 보장한다. |
| 결정적 lowering | 같은 DSL과 compiler version의 IR canonical bytes/hash가 동일하고 normalization이 idempotent다. | semantic fingerprint가 source ordering·whitespace에 흔들리지 않는다. |
| graph query | subject/predicate/object, inverse, namespace, bounded neighborhood query가 정렬된 결과와 JSON schema를 낸다. | 단순 `find/calls`를 넘어 PROV relation과 derived query 의미를 계약화한다. |
| checked patch | stale graph/IR hash, stale node hash, expected field mismatch, invalid ordered relation을 모두 write 전에 거부하고 store/Go가 byte-for-byte unchanged다. | patch precondition을 semantic delta와 provenance까지 확장한다. |
| transactional reconcile | 한 fact가 source span/evidence를 잃으면 전체 reconcile이 실패하고 기존 model과 evidence가 보존된다. | candidate·syntactic·deterministic fact를 분리하고 partial commit을 금지한다. |
| BX laws | `Get-Put`, `Put-Get`, `DSL→IR→Go→lifted IR`, normalization idempotence를 property/table test로 검증한다. | Go lifting이 implementation detail을 business truth로 만들지 않는 negative test를 포함한다. |
| generated locality | 한 activity의 input/slot만 바꾸면 해당 marker region만 바뀌고 unrelated ID, slot body, comments, file order가 유지된다. | stable generated-region marker와 semantic source map을 함께 검사한다. |
| projection sync | export는 no-op이면 no-write, import는 명시된 source만 반영, 양쪽 독립 변경은 conflict, `verify`는 write-free다. | DSL/IR/Go의 다중 projection에서 silent divergence를 금지한다. |
| diagnostics | invalid fixture마다 stable code, source span, expected/actual, repair ID, authority layer가 JSON/text 모두에 존재한다. | Agent가 repair 전에 semantic scope와 evidence requirement를 알 수 있다. |
| test semantics | test discovery/filter, expected-fail의 unexpected pass, per-test evidence, semantic ID mapping을 검증한다. | test 자체도 provenance graph의 검증 Activity/Entity로 남긴다. |
| runtime loop | canonical billing fixture에서 `check`, focused `test`, `run`의 exit/output이 일치하고 실행 결과가 semantic activity와 연결된다. | run 결과를 단순 stdout이 아니라 provenance receipt로 만든다. |
| cache correctness | cache key에 DSL/IR semantic hash, generator/verifier version, relevant imports가 포함되고 unrelated edit는 hit, relevant edit/corruption은 miss다. | reconstructable cache와 durable truth/evidence를 분리한다. |
| conformance breadth | pass/fail, round-trip, locality, marker corruption, stale/conflict, namespace, candidate promotion, cache, race fixture를 CI에서 모두 실행한다. | zerolang의 graph parity 범위를 PROV/evidence/BX까지 확장한다. |
| CI closed loop | allowed semantic scope와 actual delta를 비교하고, stale projection/generated output/evidence freshness가 merge를 막는다. Builder가 verifier/policy를 약화시키면 별도 보호된 Gate가 실패한다. | Agent의 intelligence와 safety를 분리하고 “검사 코드가 같은 PR에서 약화되는” 경로를 차단한다. |
| deterministic ops | `gofmt`, `go vet`, `go test`, `go test -race`, CLI check와 conformance가 clean이며 Go file/function cap도 검사한다. | 현재 저장소의 300/75 cap을 실제 gate와 연구 evidence에 포함한다. |

## 6. 현재 open PR review snapshot

2026-08-12에 `integration`을 base로 한 open PR 11개를 확인했다. GitHub 화면의 push/PR
check가 중복 표시되는 경우는 합쳐서 적었으며, 아래 평가는 구현 PR을 수정하지 않고
통합 전 Guardian이 확인할 위험과 수용 조건만 적은 것이다.

| PR | 범위 / 현재 신호 | graph-first·conformance·closed-loop review 포인트 |
| --- | --- | --- |
| [#1 generator](https://github.com/kimjooyoon/meta-ontology-go/pull/1) | `internal/generator/**`; draft. Go test/vet는 보이지만 Semantic conformance는 실패. | marker/slot/locality가 semantic kernel·CLI와 실제로 연결되는지 확인한다. 전체 conformance와 deterministic regeneration, unrelated region 보존이 green이어야 한다. |
| [#2 docs](https://github.com/kimjooyoon/meta-ontology-go/pull/2) | governance/conformance docs와 fixture; draft. Go test/vet는 보이지만 Semantic conformance는 실패. | 문서가 현재 gate보다 앞서 unsupported 기능을 주장하지 않는지, fixture의 `check`/generate가 실제 command contract로 통과하는지 확인한다. 이 연구 문서는 별도 `docs/research/**`라 구현 PR을 침범하지 않는다. |
| [#3 semantic](https://github.com/kimjooyoon/meta-ontology-go/pull/3) | identity/PROV-inspired IR/validation; draft. scoped Go test/vet는 보이지만 Semantic conformance는 실패. | namespace, stable ID, canonical hash, candidate/deterministic fact가 CLI와 bidir에 연결되어야 한다. source span 없는 delta와 cross-namespace merge negative test가 필요하다. |
| [#4 cli](https://github.com/kimjooyoon/meta-ontology-go/pull/4) | check/generate/query/analyze/LSP CLI; draft. gofmt·일부 CLI check 외 Go test/vet/race·CI policy가 실패. | 현재 가장 큰 closed-loop blocker다. `check`가 실제 semantic authority를 사용하고 query JSON, generator freshness, race와 policy gate를 모두 통과할 때까지 통합하지 않는다. |
| [#5 analyzer](https://github.com/kimjooyoon/meta-ontology-go/pull/5) | 등록 symbol 중심의 conservative Go lifting; draft. 서로 다른 실행의 test/vet 성공·실패와 conformance 실패/skip이 보임. | unknown/helper call을 semantic fact로 올리지 않는 negative test, ambiguous binding candidate, source span, deterministic ordering을 full integration에서 재확인한다. |
| [#7 bidir](https://github.com/kimjooyoon/meta-ontology-go/pull/7) | BX laws, delta reconcile, locality; draft. Go test/vet 실패, conformance skip. | Get-Put/Put-Get/round-trip가 현재 syntax·semantic·generator API에서 동시에 성립하는지, conflict 시 transactional인지 확인한다. “scoped test 통과”만으로 merge하지 않는다. |
| [#8 cache](https://github.com/kimjooyoon/meta-ontology-go/pull/8) | content-addressed cache; draft. test/vet 실패와 conformance 실패/skip이 보임. | cache hit가 semantic hash를 빼먹지 않는지, corruption/atomicity/parallel compute가 full suite에서 증명되는지 확인한다. stale evidence를 cache hit로 숨기지 않아야 한다. |
| [#9 go-version](https://github.com/kimjooyoon/meta-ontology-go/pull/9) | `go.mod` toolchain only; non-draft. Go/test/vet/race/conformance는 성공하지만 CI policy 실패. | #12의 branch exception이 실제로 이 PR의 `go`/`toolchain` directive만 허용하는지 재실행한다. policy를 우회한 상태로 ready merge하지 않는다. |
| [#10 lsp](https://github.com/kimjooyoon/meta-ontology-go/pull/10) | stdlib LSP transport/features; draft. scoped Go test/vet는 성공하지만 Semantic conformance 실패. | LSP가 독자적인 semantic authority가 되지 않고 parser/IR/diagnostic contract를 사용해야 한다. stale document, source span, completion/definition과 CLI check의 결과 일치를 확인한다. |
| [#11 syntax](https://github.com/kimjooyoon/meta-ontology-go/pull/11) | `.gooo` lexer/parser; draft. test/vet/race와 Semantic conformance는 성공하지만 CI policy 실패. | syntax 자체의 green을 semantic lowering/identity/conformance의 green으로 과대해석하지 않는다. branch policy와 source span/diagnostic compatibility를 확인한 뒤 후속 PR의 기반으로 삼는다. |
| [#12 ci-workflow](https://github.com/kimjooyoon/meta-ontology-go/pull/12) | branch policy, verifier, workflow; non-draft. 표시된 모든 check는 성공이나 merge state는 unknown. | 보호된 Gate를 변경하는 PR이므로 independent Guardian review가 필요하다. `agent/go-version`의 go.mod directive 예외만 허용하고, 다른 branch의 path/scope/DAMP/DRY/semantic gate를 약화하지 않는지 확인한다. |

### 통합 전 review checklist

- [ ] PR base가 `integration`, head가 `agent/*`, scope가 한 authority boundary 안이다.
- [ ] 변경된 `.gooo`/IR/Go semantic ID와 실제 semantic delta가 목록화되어 있다.
- [ ] generated region은 generator가 재생성한 결과이며, handwritten slot 외 수동 수정이 없다.
- [ ] graph/IR fingerprint와 source span/evidence가 모든 accepted delta에 있다.
- [ ] stale hash, expected-value conflict, ambiguous identity, invalid shape가 negative
  fixture에서 atomic하게 거부된다.
- [ ] `check`, focused `test`, `run`, projection verify 결과와 JSON contract가 일치한다.
- [ ] round-trip, locality, namespace, candidate promotion, generated freshness가 full
  integration branch에서 실행된다.
- [ ] `gofmt`, `go vet ./...`, `go test ./...`, `go test -race ./...`와 해당 CLI check가
  통과한다.
- [ ] conformance 실패/skip이 “기존 worktree 문제”로만 설명되지 않고 통합 기준에서 해소됐다.
- [ ] Builder가 verifier/ontology/policy를 함께 바꾸지 않았고, Gate가 같은 변경의 evidence를
  스스로 승인하지 않는다.

### 권장 통합 순서와 주의점

1. #12의 policy 변경은 별도 Guardian 검토 후 적용하고, 그 다음 #9를 정책 green으로
   재검증한다.
2. syntax/semantic 기반을 먼저 통합한 뒤 bidir/analyzer와 generator를 연결한다. 각 단계마다
   전체 branch에서 canonical hash·identity·round-trip을 다시 확인한다.
3. CLI는 그 실제 API 위에서 `check/query/generate`를 제공해야 한다. LSP는 CLI/IR을 우회하는
   두 번째 semantic implementation이 되지 않게 한다.
4. cache는 semantic/provenance contract가 안정된 뒤 붙이고, cache hit가 conformance나
   evidence freshness를 생략하지 않는지 검증한다.
5. #2의 governance/conformance 문서는 각 통합 단계의 실제 지원 범위와 동기화한다. 문서가
   실패한 gate를 green으로 표현해서는 안 된다.

## 7. 실험 backlog

다음 항목은 구현 완료를 주장하는 목록이 아니라, graph-first contract가 생겼을 때
바로 실행할 수 있도록 고정한 실험 설계다. 현재 저장소의 CLI는 일부 command가 아직
placeholder이므로, command가 없으면 실험을 pass로 표시하지 않고 `blocked`로 기록한다.

### 공통 실험 프로토콜

- fixture와 실행 결과에 repository commit, compiler/generator/verifier version, fixture
  hash, pre/post graph·IR·projection hash를 함께 기록한다.
- 각 run은 깨끗한 temporary workspace에서 수행하고 stdout/stderr, JSON output, changed
  path, changed generated marker, evidence ID를 보존한다.
- 같은 입력을 두 번 실행해 canonical output과 decision이 같은지 확인한다.
- 실패 시 큰 통합 테스트를 무작정 늘리지 않고, 최소 counterexample fixture와 다음 조치를
  backlog에 추가한다.

### EXP-GRAPH-01 — projection/patch/query contract matrix

**가설.** clean projection, checked patch, 좁은 query를 하나의 작은 fixture matrix로
묶으면 graph·projection·semantic IR 사이의 authority drift를 command 경계에서 잡을 수
있다.

**Fixture.** `conformance/graph-first/billing/`에 `main.gooo`, 기대 IR fact set,
generated Go marker view, handwritten slot, unrelated declaration을 둔다. 다음 상태를
각각 복제한다: clean, no-op export/import, one-local patch, stale graph hash, stale node
hash, expected field mismatch, projection-only edit, graph-only edit, 양쪽 독립 edit.

**측정값.** command별 exit code와 diagnostic code, pre/post semantic hash, graph hash,
changed path 수, changed marker 수, slot bytes, query result의 canonical hash/order,
evidence receipt 수와 wall-clock time을 기록한다.

**통과 기준.** clean fixture는 `query → patch --check-only → patch → check → test →
run`이 통과한다. no-op export/import와 `verify-projection`은 write가 0이고, matching
precondition patch는 의도한 region만 바꾼다. stale/mismatch/양쪽 edit는 write 없이
stable repair diagnostic으로 실패한다. 같은 query의 두 결과는 byte-identical이다.

**실패 시 다음 조치.** projection, patch, query 중 최초로 invariant가 깨진 층으로
fixture를 분리하고 해당 command contract를 먼저 고정한다. silent overwrite가 있으면
authority/reconciliation policy를 수정하기 전까지 downstream generator/test를 확장하지
않는다.

### EXP-GRAPH-02 — stale patch와 concurrent edit

**가설.** graph/IR fingerprint와 node/field expected value를 compare-and-swap처럼
사용하면 Agent A와 B의 stale semantic edit를 partial write 없이 검출할 수 있다.

**Fixture.** 하나의 base graph에서 A/B snapshot을 만든다. 같은 literal/관계에 대한
충돌 patch, 서로 다른 node에 대한 독립 patch, declaration rename 대 alias edit, graph hash는
같지만 node hash가 다른 patch, 이미 적용된 patch의 재적용을 준비한다.

**측정값.** accept/reject 수, false accept 수, store/Go byte diff, partial write 여부,
diagnostic code와 repair target, 재조회 후 recovery command 수, 충돌 후 semantic hash를
기록한다.

**통과 기준.** stale graph/node/field precondition은 100% reject하고 store·projection·
evidence를 변경하지 않는다. 독립 node patch는 정책이 허용하는 경우 모두 accept되며,
같은 semantic target은 deterministic conflict가 된다. 같은 요청을 재실행해도 결과가
달라지지 않는다.

**실패 시 다음 조치.** false accept면 precondition에 semantic node hash와 relevant
relation closure를 추가한다. partial write면 atomic staging/rollback을 우선 수정한다.
conflict code가 재현되지 않으면 query snapshot과 repair schema를 먼저 contract test로
고정한다.

### EXP-GRAPH-03 — query completeness와 결정성

**가설.** Agent가 필요한 neighborhood만 query해도 전체 IR을 읽은 reference evaluator와
동일한 semantic fact set을 얻고, inverse/namespace 경계가 결과를 오염시키지 않는다.

**Fixture.** `billing`, `fraud`, `settlement` 세 namespace에 같은 display name을 가진
Entity/Activity를 만들고 `used`, `wasGeneratedBy`, `invokes`, inverse와 bounded
neighborhood query를 준비한다. 명시 fact, candidate fact, derived fact를 구분하고
unrelated namespace를 섞는다.

**측정값.** query별 reference fact set과 실제 result set의 차집합, result ordering과
canonical JSON hash, bounded depth/size, query latency, candidate 누출·누락 수를 잰다.

**통과 기준.** 명시된 query semantics와 실제 결과가 exact match하고, 반복 실행 및
동일 fact의 입력 순서 변경에도 결과 bytes가 같다. namespace 밖 fact와 candidate가
허가 없이 섞이지 않으며, bounded query가 정해진 상한을 넘지 않는다.

**실패 시 다음 조치.** 누락이 vocabulary/derived-rule 정의 문제인지 query 구현 문제인지
분리한다. vocabulary가 모호하면 rule을 추가하기보다 query를 unsupported로 명시하고,
구현 오류면 해당 smallest graph를 regression fixture로 고정한다.

### EXP-LOCAL-01 — locality counterexamples

**가설.** unambiguous semantic edit는 해당 IR/Go/generated marker만 바꾸며, identity
matching이 모호한 경우에도 조용한 ID 탈취나 unrelated rewrite가 발생하지 않는다.

**Fixture.** 다음 변형을 각각 base fixture에서 만든다: 동일 shape declaration을 앞/뒤에
삽입, 같은 종류 statement reorder, activity rename, namespace 동명이인 추가, unrelated
file edit, handwritten slot edit, marker 중복·누락·범위 불일치, generated region 밖의
주석 edit.

**측정값.** 보존·retire된 semantic ID 수, changed generated marker 목록, unchanged
region/slot byte hash, source-map span 변화, semantic delta cardinality, conflict/candidate
수와 diagnostic을 비교한다.

**통과 기준.** 정상적인 단일 edit는 target closure만 변경하고 unrelated ID·slot·comment를
보존한다. ambiguous insertion은 기존 ID를 임의로 빼앗지 않고 retire/conflict/candidate
중 하나의 명시된 정책을 따른다. malformed marker는 regeneration 전에 실패하고
handwritten body를 잃지 않는다.

**실패 시 다음 조치.** 실패를 identity matcher, generator ordering, marker parser,
source-map 중 한 층의 최소 counterexample로 축소하고 그 fixture를 영구 conformance에
추가한다. locality assertion을 완화하거나 “전체 파일 재생성”으로 숨기지 않는다.

### EXP-EVIDENCE-01 — closed-loop Agent evidence

**가설.** Agent가 변경 전 scope와 변경 후 semantic delta/evidence를 구조화해 제출하게
하면 Builder의 추론 오류와 Gate의 안전성 판단을 분리할 수 있다.

**Fixture.** 다음 task transcript를 재생한다: 허용된 local activity edit, 허용 범위를
넘는 `invokes` edge 추가, stale patch 재시도, generated freshness 위반, expected-fail
test를 unexpected pass로 바꾸는 edit. 각 fixture에는 request ID, allowed semantic
scope fingerprint, base graph/IR hash, delta, projection hash, command results, evidence
hash, Builder/Guardian/Gate actor를 넣는다.

**측정값.** evidence field completeness, `actual delta ⊆ allowed scope` 정확도, stale
evidence 검출률, 같은 input에 대한 Gate decision 재현성, Builder self-approval 발생 수,
task당 command/turn 수를 기록한다.

**통과 기준.** accepted change는 100% complete·fresh evidence를 가지고, out-of-scope,
stale, generated-drift fixture는 merge 전에 reject된다. 같은 snapshot과 evidence를 두
번 평가하면 decision과 reason code가 같다. Builder는 verifier/policy를 바꿔 자기 변경을
승인할 수 없다.

**실패 시 다음 조치.** 누락된 field를 optional로 낮추지 말고 schema의 required field로
승격한다. stale evidence가 통과하면 source/IR/Go hash와 verifier version을 evidence와
cache key에 포함한다. 역할 분리가 깨지면 해당 Gate를 보호된 kernel 경계로 옮기고
independent Guardian 검토를 요구한다.

### EXP-EVIDENCE-02 — provenance와 cache freshness

**가설.** reconstructable cache는 재사용할 수 있지만 provenance/evidence는 append-only
durable record로 유지하면 cache hit가 오래된 검증을 새 변경의 증거처럼 재사용하지 않는다.

**Fixture.** 동일 input 반복, DSL whitespace-only edit, semantic ID rename, handwritten
logic edit, generator/verifier version 변경, relevant import 변경, cache entry corruption,
evidence field tamper를 각각 실행한다.

**측정값.** cache hit/miss와 invalidation reason, source/IR/Go/evidence hash 일치 여부,
evidence sequence length, tamper 검출률, 재계산 시간, durable record의 overwrite 여부를
측정한다.

**통과 기준.** whitespace-only와 unrelated edit만 허용된 범위에서 hit되고 semantic,
generator, verifier, import 변경은 miss된다. evidence hash가 현재 input과 다르면 Gate가
reject하고 새 receipt를 append한다. cache purge는 durable evidence를 삭제하지 않는다.

**실패 시 다음 조치.** 빠진 key를 invalidation contract에 추가하고, evidence overwrite가
발생하면 append-only storage/sequence 검사를 먼저 고친다. cache와 provenance의 수명이
불명확하면 두 저장소를 분리한 뒤 다시 측정한다.

### 실행 순서와 backlog 완료 조건

1. `EXP-GRAPH-01`로 command/result schema를 고정한다.
2. `EXP-GRAPH-02`와 `EXP-GRAPH-03`으로 patch precondition과 query semantics를 검증한다.
3. `EXP-LOCAL-01`로 identity·generated locality의 counterexample을 추가한다.
4. `EXP-EVIDENCE-01`과 `EXP-EVIDENCE-02`로 scope/evidence/cache의 폐쇄루프를 검증한다.

각 실험은 fixture 파일, machine-readable result, 명시된 threshold, 최소 하나의
regression case, 담당 역할(Builder/Guardian/Gate)을 갖기 전에는 완료로 표시하지 않는다.
command가 아직 구현되지 않은 경우도 실패로 위장하지 않고 `blocked: missing contract`
로 남긴다.

## 8. 출처와 재현 경로

모든 zerolang 관찰은 위 snapshot의 공식 저장소 파일을 기준으로 했다.

- [README.md](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/README.md): graph-first model, daily loop, command surface
- [CLI reference](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/docs/articles/cli-reference.md): query/patch/check/test/run/import/export contracts
- [Graph architecture](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/docs/articles/concepts/graph-architecture.md): graph editing loop and human review boundary
- [Projections](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/docs/articles/concepts/projections.md): explicit sync and no silent divergence
- [Testing](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/docs/articles/testing.md): graph-backed tests and expected failures
- [AGENTS.md](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/AGENTS.md): isolated workflow, conformance, and agent checks
- [validation-suite.mts](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/scripts/validation-suite.mts): named conformance phases and sharding
- [program-graph-smoke.mts](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/scripts/program-graph-smoke.mts): store/projection/atomicity smoke checks
- [program-graph-parity.mts](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/scripts/program-graph-parity.mts): graph hash, node identity, patch, reconcile, and locality checks
- [CI workflow](https://github.com/vercel-labs/zerolang/blob/afcc72da649fe4d4c670ac1489c2197d37436051/.github/workflows/ci.yml): fast/deep conformance and runtime validation

이 문서에서 `meta-ontology-go` PR 상태는 해당 날짜에 GitHub의 open PR/check를 조회한
snapshot이다. 상태가 바뀌면 표를 최신 check 결과로 갱신해야 하며, 구현 PR의 코드를 이
연구 branch에서 수정하지 않는다.
