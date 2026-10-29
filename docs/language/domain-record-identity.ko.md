# 같은 숫자에 다른 의미를 붙이기

가격 29와 수량 4는 모두 정수지만, 프로그램에서 맡는 역할은 다르다.
Gooo에서는 한 필드 레코드에 서로 다른 entity ID를 선언해 이 경계를 표현할 수 있다.
이번 과제는 가격을 받는 포트에 수량의 결과를 연결했을 때 그 실수를 발견하는 것이다.

```gooo
entity Value id "urn:domain438:money" fields {
 field value id "urn:domain438:money-value" type "urn:gooo:type:integer" required one
}
activity Keep(Value) -> Value id "urn:domain438:keep-money" computes "return input"
```

수량 패키지도 같은 `Value` 이름과 `value` 정수 필드를 사용한다.
다만 entity ID는 `urn:domain438:quantity`다. 사용하는 파일에서는 import로 소유 패키지를 밝힌다.

```gooo
import price "domain438/money"
import quantity "domain438/count"
activity Multiply(price.Value, quantity.Value) -> Total id "urn:domain438:multiply" computes "return Total{value: input0.value * input1.value}"
bind price.Keep.result -> Multiply.input0
bind quantity.Keep.result -> Multiply.input1
```

정수 기본 타입의 별칭 두 개는 모두 같은 `urn:gooo:type:integer`를 가리킨다.
업무상 구별은 레코드의 명시한 entity ID로 선언한다.
표지판이 같아도 부품 번호가 다르면 서로 바꿔 끼우지 않는 방식에 가깝다.

## 새 소스에서 확인한 범위

| 항목 | 관찰 |
| --- | --- |
| 같은 이름·필드 구조의 레코드 | 서로 다른 업무 ID 2개 유지 |
| 명시한 활동 연결 | 2개를 다른 타입 ID로 유지 |
| 수량 결과 → 가격 입력 연결 | `PACKAGE_BINDING_TYPE_MISMATCH`로 1개 거절, 부분 image 수락 0 |
| 실행 | 입력 사례 2개에서 활동 출력 6/6 일치 |
| 내부 Go 이름 | 두 업무 ID가 서로 다른 이름으로 생성됨 |
| 반복 결과·원본 | 결과 일치, 원본 manifest 보존 |
| 모델 호출 | 0회 |

새 테스트의 실제 생산자는 `d8779545bf82b8688993d77ef50c1aa0feec2b82`에
테스트 파일 하나를 더한 overlay다. 구현 변경 없이 기존 레코드·활동·bind 동작을 확인했다.
[관찰 기록](../../examples/domain-record-identity/observation.json)은 이 생산자와 소스 해시를 보존한다.
이 결과는 선언한 두 사례와 한 잘못된 연결의 범위이며, 전체 업무 요구의 완성률은 별도로 정한다.

외부 JSON의 `{"value": 29}` 자체에는 가격의 출처나 통화가 들어 있지 않다.
현재 과제는 소스가 선언한 타입과 포트의 연결을 확인한다.
통화, 단위 변환, 출처, 값의 범위는 추가 선언과 조건·실행 관찰이 필요한 다음 요구다.
가격과 수량을 곱한다는 계산식도 소스가 직접 정한다.

## 비교에서 얻은 다음 기준

[TypeQL의 데이터와 계산 구분](https://typedb.com/docs/typeql-reference/functions/functions-vs-rules/)은
업무 개념과 계산 방향을 각각 읽는 데 참고한다.
[OML의 모듈과 멤버](https://opencaesar.io/oml/)는 어느 모듈의 개념을 사용하는지 밝히는 데 참고한다.
Gooo에서는 명시 ID, 계산 본문, 허용한 선택지와 실제 관찰을 이어가는 방향으로 개발한다.

다음 공통 과제는 가격·수량의 구분, 승인 이유, 모듈의 이름 변경이다.
각 요구를 표현한 선언, 외부 구현이 필요한 부분, 유지한 ID와 확인한 입력 사례를 센다.
관계 질의·중첩 레코드·단위 변환·에디터 연동도 지원 범위를 따로 기록한다.
동일한 업무 과제를 다른 언어에서 실행한 비교 자료는 아직 준비하지 않았다.

## 레코드 이름을 바꿀 때

다음 소스에서 `Line`의 표시 이름을 `Product`로 바꾸어도 품목의 ID는 유지한다.

```gooo
entity Line id "urn:record-items:line" fields {
 field quantity id "urn:record-items:quantity" type integer required one
}
entity Basket id "urn:record-items:basket" fields {
 field items id "urn:record-items:items" type Line required many
}
```

V7 레코드 프로필의 `PutWithEntityFieldsSupport`는 원본에서 해석한 대상 ID를 사용해,
직접 수정하지 않은 `type Line`을 `type Product`로 함께 갱신한다.
`type "totals:Line"`처럼 이름 공간을 적었다면 같은 이름 공간을 유지한다.
`type "urn:record-items:line"`처럼 ID로 적은 참조는 그대로 둔다.
다른 레코드에 예전 이름을 붙여도 기존 참조가 그 레코드로 옮겨가지 않는다.

필드의 ID·순서·필수 여부·개수와 원본의 세부 위치는 유지한다.
위치는 수정 전 원본의 기원을 가리키며, 새로 출력한 텍스트의 위치는 다시 파싱해서 얻는다.
필드 타입을 명시적으로 함께 수정했다면 그 수정은 별도로 검증한다.
이름 변경이 상충하는 타입 수정을 덮어쓰지 않으며, 실패한 갱신은 원본을 반환한다.

회귀 검사는 세 참조 표기와 예전 이름 재사용 여부의 6개 조합에서 네 필드 형태를 확인한다.
충돌·대상 삭제 등 5개 거절 사례와 일관된 타입 변경 1개는 별도로 확인한다.
소스 표기를 다시 파싱한 결과와 BX 법칙도 검사한다.
이 범위는 parser-neutral 문서 갱신 API이며, 편집기나 CLI 이름 변경 명령은 별도 구현이 필요하다.
