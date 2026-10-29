# 선언한 테스트와 관찰한 실행 읽기

실험에서 기대한 테스트 이름·결과·소스 버전과 실제로 관찰한 실행을 연결한다.
`main.gooo`가 판단 규칙을 소유하고, Go는 기존 이벤트 해독기와 관찰 운반을 맡는다.
에이전트는 프로세스 종료, 수집한 기록의 완결성, 선언한 기대의 충족을 각각 읽을 수 있다.

| 관찰 | Gooo의 판단 |
| --- | --- |
| 선언한 테스트 집합·결과·종료 코드가 일치 | `MATCHED_EXPECTATION` |
| 명시적으로 기대한 실패와 종료 코드 1이 일치 | `MATCHED_EXPECTATION` |
| 종료 코드 0이며 이름 있는 테스트 기록이 없음 | `NO_TESTS_OBSERVED` |
| 테스트 누락·추가·다른 결과 | `TEST_SET_MISMATCH` |
| 같은 실험 이름에 다른 소스 버전의 결과가 연결됨 | `INVALID_OBSERVATION` |
| 프로세스 또는 이벤트 수집의 완료가 확인되지 않음 | `INCOMPLETE_OBSERVATION` |

`pass`, `fail`, `skip`은 관찰 가능한 결과다. 선언한 `skip`은 일치할 수 있고,
`pass`를 기대한 테스트가 `skip`이면 기대와 다른 결과로 남는다.
허용한 결과 집합은 Gooo의 `KnownOutcome` 한 곳에서 정의한다.
`process_outcome`은 `MATCH`, `MISMATCH`, `UNASSESSED`로 구분한다.
불완전한 기록에서 발견한 실패도 `failure_observed`에 남긴다. 계획과 관찰의
실험·소스 ID를 각각 보존하므로 다른 소스의 실패가 계획의 실패로 바뀌지 않는다.

```sh
go test ./internal/testevents ./examples/ci-strategy/declared-test-observation
go run ./cmd/gooo body-compose \
  --source examples/ci-strategy/declared-test-observation/main.gooo \
  --entry Assess --cases examples/ci-strategy/declared-test-observation/cases.json \
  --out /tmp/gooo-declared-tests-fresh
go run ./examples/record-origin-reader --activity Assess \
  --field urn:declared-tests:assessment/observation-source \
  examples/ci-strategy/declared-test-observation/main.gooo
```

실행 출력 폴더는 매번 새 경로를 사용한다. 모델 없이 결정론적으로 실행한다.
정규화된 합성 관찰에는 독립적으로 작성한 기대값 21개가 있다. 준비된 실행과
독립 native 실행·저장 replay가 같은 사례들을 확인하며 분모를 합치지 않는다.
별도의 실제 Go 프로세스 5개가 통과·의도한 실패·실행 누락·skip·중첩 테스트 출력을 만든다.
손상된 접미부 검사는 그중 실패 출력을 변형한 반례이며 추가 실행 횟수로 세지 않는다.

`Assess`는 호출자가 보관한 `Capture`를 변경하지 않는다. 해독 오류가 있으면
기록된 접두부로 계산한 Gooo 결과와 오류를 함께 반환한다. `Capture.Completed`는
0이 아닌 종료 코드를 포함한 프로세스 종료를 뜻하고, `StdoutComplete`는 호출자가
관찰한 수집 상태다. `Arguments`는 제공된 인자 벡터이며, 원래 인자의 관찰 범위를
`argv_captured`에 별도로 남긴다. API가 호출자의 관찰을 추가로 인증하지는 않는다.

공통 해독기의 범위는 한 패키지의 이름 있는 Go 테스트다. 패키지 수준 결과,
진단 메시지와 고루틴 정리 의무는 호출자가 보존한 원래 출력의 별도 관찰 영역이다.
부모와 leaf를 포함한 모든 named node를 보존한다. leaf 하나를 선언하고 실제로
부모·leaf 두 개를 관찰하면 집합 불일치다. 부모를 포함한 두 이름을 명시하면
같은 원문으로 두 노드의 일치를 확인한다. 이 재평가는 추가 프로세스 실행으로 세지 않는다.
leaf 전용 렌즈는 원래 capture와 `sha256.Sum256(captured.Stdout)`, 기대 inventory,
선택 근거에 연결해야 한다. 이 운반기는 선택하지 않은 노드나 실패를 걸러내지 않는다.
`failure_observed=false`는 제공된 이름 있는 이벤트에서 실패를 보지 못했다는 뜻이다.
외부 라이브러리가 포착 전에 숨긴 오류의 원인은 이 입력으로 복원할 수 없다.

기존 callback 추출기의 지정 테스트 통과 조건과 실패 접두부 보존은 유지한다.
이 예제는 CI 상태를 쓰거나 병합을 승인하지 않는다. 합성 관찰의 시작점은
[19개 사례의 공개 초안](https://gist.github.com/kimjooyoon/91a20b136ceb318c5457b3f57ef9f20c/fecd5f244c9b8d0c9b055b3dd7dec8d94a3c0ca5)이며,
여기서는 실제 이벤트 연결과 skip 사례를 더했다.


## 실제 외부 TDD 기록

같은 Gooo 소스로 OpenFGA의 보관된 실패 단계와 수정 후 기록을 읽었다.
사전에 작성된 목록은 15개 leaf와 부모 4개를 포함한 이름 있는 노드 19개다.

| 기록 | 기대 결과와 일치한 노드 | Gooo 판정 |
| --- | --- | --- |
| 실패 단계 | 9/19 | `TEST_SET_MISMATCH` |
| 수정 후 | 19/19 | `MATCHED_EXPECTATION` |

네이티브 실행은 독립적으로 작성한 판정 레코드 2/2를 확인했다.
저장 결과 재실행도 같은 두 입력을 확인했으며, 모델 호출은 0회다.
이 수치는 위의 합성 사례 21개나 현재 OpenFGA의 전체 테스트 수와 합산하지 않는다.
외부 프로젝트의 실제 채택이나 언어 전체의 정답률로 해석하지 않는다.

원본 수집의 종료·대기·출력 해시와 명시한 부분 트리를 확인했다.
요청한 인자 벡터는 실제 argv 포착과 구분해 `argv_captured=false`로 두었다.
4,336개 노드가 있는 별도 두 기록은 관찰된 기준 목록이며 현재 입력 상한도 넘는다.
이들을 선언한 의도로 바꾸거나 잘라서 평가하지 않았다.

[Gooo 소스·정규화 입력·기대값·범위를 고정한 재현 자료](https://gist.github.com/kimjooyoon/639d92615ff95d0aab4c7d15a9018249/5baf38756e2f6baf26df9abb99da05c89ee39de8)에서
생산자 버전과 각 분모를 확인할 수 있다.
