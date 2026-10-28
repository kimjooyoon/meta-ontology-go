# 같은 요청으로 권한 모델 변경 살펴보기

모델을 바꾸면 같은 요청의 허용 여부도 달라질 수 있습니다. 이 예제는 저장된
관찰의 입력과 출처를 먼저 맞춘 뒤, 결과 변화와 의도한 변화를 Gooo로 비교합니다.
예를 들어 editor에게 viewer를 추가하는 변경에서 Alice의 요청이 허용되면
`EXPANSION / EXPECTED`가 됩니다. 관찰이 오류로 끝나면 `UNKNOWN`으로 남습니다.

기존 [OpenFGA](https://github.com/openfga/openfga) 메모리 저장소에서 수집한
5쌍의 관찰을 사용합니다. 같은 프로세스에서 관계 3개를 한 번 기록하고,
관계를 추가로 수정하지 않은 상태에서 두 모델에 각각 `Check`를 호출했습니다.

| 사례 | 이전 → 이후 | Gooo 결과 |
| --- | --- | --- |
| editor인 Alice의 viewer 요청 | 거부 → 허용 | EXPANSION / EXPECTED |
| 그룹 구성원 Bob | 허용 → 허용 | UNCHANGED_ALLOW / EXPECTED |
| 관계 없는 Carol | 거부 → 거부 | UNCHANGED_DENY / EXPECTED |
| 같은 contextual membership을 준 Carol | 허용 → 허용 | UNCHANGED_ALLOW / EXPECTED |
| 정의되지 않은 owner 관계 | 오류 → 오류 | UNKNOWN / UNASSESSED |

마지막 사례는 양쪽 모두 `Code(2000)`과 응답 부재를 기록했습니다.
`recorded`, `invoked`, `completed`, `error_present`, `response_present`를 따로
다루므로 기록 없음·호출 전·미완료·오류·응답 없음을 구분합니다.

## 실행

저장소 루트에서 실행합니다. 출력 폴더는 새 경로를 사용합니다.

```sh
go run ./cmd/gooo body-compose \
  --source examples/openfga-model-change/main.gooo --entry Compare \
  --cases examples/openfga-model-change/cases.json --out /tmp/gooo-fga-first

go run ./cmd/gooo body-compose \
  --source examples/openfga-model-change/main.gooo \
  --cases examples/openfga-model-change/cases.json \
  --composition /tmp/gooo-fga-first/composition.json --out /tmp/gooo-fga-replay
```

재생 단계는 저장된 진입점을 사용하므로 `--entry`를 생략합니다.
고정된 `computes` 본문으로 실행하며 모델 설정은 필요하지 않습니다.
`runtime.finite_passed/finite_total`은 같은 21개 사례를 분모로 합니다.
원래 관찰만 확인하려면 `observed-cases.json`을 사용하면 됩니다. 그 5개는
전체 21개 안에 포함됩니다.

회귀 검사는 같은 21개 입력을 해석 실행·네이티브 실행·저장된 구성 재생에서
확인합니다. 세 경로의 실행 수를 서로 다른 입력 수로 합산하지 않습니다.
원문 정규화 검사도 저장된 관찰을 읽으며 API를 다시 호출하지 않습니다.

```sh
go test ./internal/bodycodegen ./cmd/gooo -run OpenFGA
go test ./examples/openfga-model-change/normalize
```

## 의도와 출처

`main.gooo`가 관찰 상태, 입력·데이터·실행 범위의 일치 여부, 변화 종류와
의도 일치 여부를 결정합니다. Go 도구는 원문을 읽고 비교 입력을 구성합니다.

`case_id`는 관찰 파일의 표시 키입니다. 의도 계약은 모델 ID를 제외한 요청 전체를
정규화한 식별자와 두 모델의 소스 해시에 연결됩니다. 같은 입력의 표시 키를
바꾸어도 판단은 유지됩니다. 모델 ID의 요청 바인딩은 별도로 확인합니다.
contextual tuples, `null`과 기본 속성, 배열 순서는 비교 입력에 남습니다.

`cases.json`에는 실제 관찰 5개와 로컬 반례 16개가 있습니다. 입력·데이터·실행 범위
변경, 미실행과 관찰 누락, 의도하지 않은 확대·축소, 계약에 없는 입력 등을 다룹니다.
각 입력과 결과의 `origin_kind`, [origins.json](origins.json)의 사례별 출처로
수집한 관찰과 작성한 반례를 구분합니다.

## 원문과 재현

원문은 `raw-observations.json.gz`에 비손실로 보관했습니다. 15,130바이트를
1,936바이트로 저장하며, 압축을 풀면 원래 바이트의 SHA-256이 유지됩니다.
수집기 원문은 [origin/collector.go.txt](origin/collector.go.txt)에 있습니다.

```sh
go run ./examples/openfga-model-change/normalize \
  examples/openfga-model-change/raw-observations.json.gz /tmp/gooo-fga-normalized.json
```

`request_serialization_digest`는 수집기가 보고한 직렬화 해시로 보존합니다.
`request_input_digest`는 Go 도구가 비교용 요청에서 계산합니다. 수집은 HTTP
리스너 없이 프로세스 안에서 수행했습니다. 실행 범위 식별자는 그 통제된 기록을
묶으며, 같은 운영 저장소의 두 시점에 적용하려면 별도 스냅샷·실행 환경 관찰이
필요합니다. 현재 기대값 검증의 범위는 위 21개 입력입니다.

## 계산식의 원래 입력 따라가기

저장소 루트에서 같은 소스의 입력과 반환 필드 의존성을 조회합니다.
관찰 사례를 실행하거나 권한 판단을 다시 요청하지 않습니다.

```sh
go run ./cmd/gooo body-origins --activity Compare \
  examples/openfga-model-change/main.gooo
```

`--format text`로 원문 위치와 레코드 버전 연결을 읽을 수도 있습니다.
`before_input_digest`와 `after_input_digest`는 같은 자식 필드 선언을 사용하지만,
각각 `ComparisonPair.before`와 `ComparisonPair.after`의 입력 경로를 유지합니다.
조회 스키마는 `gooo/typed-value-origins/v2`입니다.
JSON의 `flow.record_layouts`, `record_values`, `record_versions`는 필드 기원과
복사·호출·반환 이력을 함께 보관합니다. 같은 필드 기원 배열만 공유하므로
두 입력의 값이 같다는 이유로 출처가 합쳐지지 않습니다.

확인할 지표는 그래프 노드 수, 선언된 입력·출력 수, 레코드 버전 수와 공유 배열 수,
실제 원문 위치가 매핑된 사건 수입니다. 공유 전후의 필드 참조 수는 같은 버전
집합을 분모로 비교합니다. 전체 메모리나 실행 속도 개선을 뜻하지는 않습니다.
그래프 한도는 512노드이며, 실행 중의 구체적인 경로나 중첩 목록 인덱스는
이 소스 조회의 범위 밖입니다.
