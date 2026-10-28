# 반복 항목을 바인딩한 곳의 필드 기원

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
파생 반복 바인딩을 포함하는 표는 `v2`를 사용한다. 동적 필드·지역 변수·원소 읽기
일지의 wire와 실행 순서는 기존 형식을 유지한다.

## 이 예제의 확인 범위

| 항목 | 이전 | 파생 관계 적용 |
| --- | --- | --- |
| 독립 기대값과 맞는 계산 | 4/4 | 4/4 |
| 재생 | 일치 | 일치 |
| 필드 읽기 소스 위치 | 2/2 | 2/2 |
| 반복 바인딩 필드의 값·선언 위치 | 0/2 | 2/2 |
| 원소 읽기 | 4회 | 4회 |
| 필드 기록·사용 | 각 8회 | 각 8회 |

동일한 `{quantity:2, price:7}` 항목 두 개는 기존 원소 읽기 기록에서 원래 위치
`0`, `1`로 구별된다. 이 변경은 반복문의 **정적 소스 바인딩**을 채운다. 각 필드의
동적 대입과 원래 원소 읽기 사이의 입력 기원 연결은 후속 작업이다. 필드 일지의
범위 안에서 `COMPLETE`여도 전체 입력에서 결과까지의 출처가 완성됐다는 뜻으로
집계하지 않는다.

다음 단계는 이미 기록된 원소 읽기와 파생 필드 바인딩을 정확한 참조로 연결하는
것이다. 값의 동일성으로 위치를 추정하지 않는다. 정적 분석의 반복 범위와 실제
실행의 원소 번호도 구분한다. 현재 `Total`의 정적 `body-origins`는
`FLOW_REQUIRES_REQUIRED_SCALAR_FIELDS`로 미해결이며, [#1834](https://github.com/kimjooyoon/meta-ontology-go/issues/1834)의
소스 반복·인덱스 기원 작업에 포함된다.
