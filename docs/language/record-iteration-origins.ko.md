# 반복 항목의 필드와 입력 위치 연결

`for item in input.items`는 각 항목의 필드를 지역 레코드 `item`에 바인딩한다.
실제 Gooo 원문에는 `item.quantity = ...` 같은 대입문이 없다. 출처를 설명할 때는
목록 표현식 `input.items`, 변수 선언 `item`, 선언된 필드의 안정 ID를 함께 사용한다.

[레코드 목록 예제](../../examples/record-collections/main.gooo)의 `Total`이 이 경우다.

```gooo
for item in input.items {
    total = total + item.quantity * item.price
}
```

네이티브 실행의 `field_sources`는 이 바인딩의 필드를 `record_iteration_field`
파생 관계로 표시한다. `derivation.local`은 기존 컴파일러가 보존한 활동·지역 변수
번호, 반복문 값 표현식과 선언 범위를 갖는다. `source.value`는 목록 표현식,
`source.target`은 실제 변수 토큰의 원본 UTF-8 범위다. 형식·필드 ID는 원래 선언에
연결한다. 사용자가 직접 작성한 필드 식은 기존 `origin` 관계를 사용한다.

변수 이름을 `row`로 바꾸거나 파일을 옮기면 해당 소스 범위와 내용 해시가 바뀐다.
활동·레코드·필드의 안정 ID와 변수 바인딩 번호는 각각의 역할을 유지한다.

저장소 루트에서 다음과 같이 확인한다.
출력 폴더는 아직 없는 경로를 지정한다.

```sh
go run ./cmd/gooo body-compose --source examples/record-collections/main.gooo \
  --entry Total --cases examples/record-collections/cases.json --out /tmp/record-origin-run
```

원본 source-only 필드 기록은 `gooo/native-field-sources/v1`을 사용한다.
파생 반복 바인딩을 포함하는 표는 `v2`를 사용한다.
과거 생성 기록에 지역 변수의 선택적 원문 주석이 없으면 계산과 재생을 유지하고,
새 파생 바인딩을 만들지 않는다. 조회는 보존된 관측 범위와 함께 사용한다.

실행 중에는 이미 기록된 원소 읽기의 **사례 안 번호**를 필드 대입에 연결한다.
네이티브 생성기는 반복문의 목록·인덱스 바인딩이 맞는 읽기 호출 바로 뒤에 필드
바인딩을 배치한다. 바인딩된 레코드의 각 필드는 같은 원소 읽기를 가리킨다.
다른 반복이나 복사 대입은 각각의 관계를 유지한다.

예를 들어 같은 값의 두 번째 항목은 다음 경로로 이어진다.

```text
input.items의 원래 위치 1
  → 실제 원소 읽기 1
    → item.quantity 필드 대입
    → item.price 필드 대입
```

연결은 기존 field/local bridge의 `v2`에 `element_reads` 정수 벡터로 저장한다.
필드·지역 변수·원소 읽기 일지의 기존 wire와 각 일지의 사건 순서는 유지한다.
벡터는 필드 대입과 같은 순서이며, 연결할 대상이 없는 대입은 `-1`, 한도 때문에
읽기를 기록하지 못한 반복 바인딩은 `-2`다. 네이티브의 `int32` 벡터는 기존
필드 대입 한도인 4,096개 안에 머문다. 값과 입력 배열은 이 벡터에 복제하지 않는다.

`field-writes`와 `field-uses` 조회는 `element_read_index` 및 기존 읽기 사실 형식인
`element`를 제공한다. 바깥 `field_id`는 선택한 레코드 멤버이고,
`element.field_id`는 원래 입력 목록의 필드다. `element.origin_index`가 원래 위치다.
저장된 기록을 조회할 때 생산 프로그램이나 모델은 실행하지 않는다.

이 사실을 받는 원시 행의 입력 설명은 `gooo/native-origin-input/v2`를 사용한다.
`field-lineage-columns/v1`과 `field-local-lineage-columns/v1`은 원소 연결이 있는
사례를 명시적으로 거절한다. 원시 행과 `field-local-columns-v2` 형식이 이 연결을
운반한다. [Gooo 경로 조회](../../examples/local-field-bridge/indexed.workspace.json)는
원래 32개 열을 `history`, 원소 정보를 `elements` 레코드로 받아 공통 탐색 활동을
사용한다. 기존 v1 기록의 조회도 지원한다.

## 이 예제의 확인 범위

| 항목 | 이전 | 파생 관계 적용 |
| --- | --- | --- |
| 독립 기대값과 맞는 계산 | 4/4 | 4/4 |
| 재생 | 일치 | 일치 |
| 필드 읽기 소스 위치 | 2/2 | 2/2 |
| 반복 바인딩 필드의 값·선언 위치 | 0/2 | 2/2 |
| 실제 필드 대입에서 원소 읽기로의 연결 | 0/8 | 8/8 |
| 원소 읽기 | 4회 | 4회 |
| 필드 기록·사용 | 각 8회 | 각 8회 |

동일한 `{quantity:2, price:7}` 항목 두 개는 기존 원소 읽기 기록에서 원래 위치
`0`, `1`로 구별된다. 중첩 반복의 독립 사례는 같은 값이 든 두 목록을 사용하며,
16개 필드 대입 중 반복 바인딩 12개는 정확한 목록 읽기, 복사 대입 4개는 기존
필드 부모를 가리킨다. 다른 경로를 같은 값으로 합치지 않는다.

한도 사례는 65,536개 읽기를 먼저 기록한 뒤 레코드 목록을 읽는다. 계산 결과
`14`와 재생은 유지되고, 이후 필드 대입 2개는 `READ_LIMIT`로 남는다. 이 경우
요약은 `PARTIAL`이다. 기록되지 않은 읽기를 이전 사건으로 대체하지 않는다.

[Gooo 소비자](../../internal/bodyexecution/testdata/record-element-locate.gooo)는
기록의 상태와 목록의 안정 ID를 보고 위치를 선택한다. 검사는 실제 조회한 8개
행과 별도로 작성한 미해결·다른 목록 거절 2개를 구분한다. 네이티브 실행 10개
기대값 비교는 생산 프로그램의 4개 계산 사례와 별도 분모다.

현재 `Total`의 정적 `body-origins`는 `FLOW_REQUIRES_REQUIRED_SCALAR_FIELDS`로
미해결이다. 다음 단계는 정적 반복의 상징적 항목·누적 값과 실제 실행의 원소 번호를
구분해 표현하고, 생성된 목록과 큰 실행 기록을 Gooo 경로 조회에 연결하는 것이다.
이 작업은 [#1834](https://github.com/kimjooyoon/meta-ontology-go/issues/1834)에 포함된다.
필드 일지의 `COMPLETE`는 해당 기록 범위의 완전성을 나타낸다.
