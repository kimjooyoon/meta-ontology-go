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

## 7. 출처와 재현 경로

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
