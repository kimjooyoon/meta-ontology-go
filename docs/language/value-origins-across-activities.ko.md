# 함수 호출과 패키지 연결을 따라 값의 기원 읽기

Gooo에서는 계산과 함께 그 계산을 구성한 선언과 후보를 살펴볼 수 있다. 예를 들어
외부 입력에 보정치를 더하고, 다른 패키지에 전달해 결과를 만들었다면 입력·보정 후보·
전달 경로를 원본 파일까지 이어 본다. 개발 소스의 `package origins --follow`가 이 조회를 제공한다.

## 저장한 실행 기록에서 시작하기

```sh
gooo package origins --receipt execution.json --follow
gooo package origins --receipt execution.json --follow --candidate selected-rule --json
gooo package origins --receipt execution.json --follow --activity urn:example:calculate
```

`--follow`는 저장된 최종 소스를 읽어 함수 호출과 `bind`를 가로지르는 정적 경로를 만든다.
기본 출발점은 워크스페이스의 진입 activity다. `--activity`에는 출발점의 안정적인 ID를
넣는다. `--follow`가 없는 조회는 기존처럼 조립한 activity의 저장된 그래프를 읽는다.

JSON의 `report.rows`는 기존 후보 사용처이며, `report.value_flow`는 함수 사이의 기원이다.
각 보고서의 상태와 잘림 여부를 따로 읽는다. 기본 한도는 각각 32행이고 최대 128행이다.

## 경로에 남는 연결

| 표시 | 읽는 방법 |
| --- | --- |
| `input` | 현재 계산에 사용한 앞선 값이나 대입으로 이동 |
| `control` | 계산이 속한 분기를 정한 조건으로 이동 |
| `call_return` | 호출한 함수의 반환식으로 이동 |
| `call_argument` | 그 호출에 전달한 인자로 돌아옴 |
| `bind` | 현재 입력을 제공한 다른 activity의 반환식으로 이동 |

같은 함수를 두 번 호출해도 각 호출의 인자를 구분한다. 첫 인자만 반환하는 함수라면
둘째 인자는 반환값의 기원에 포함하지 않는다. 둘째 인자의 식이 실행 중 평가되는지는
별도의 문제이며, 이 조회는 반환값에 대한 의존 관계를 설명한다.

선택한 후보에는 조립 입력과 최초 패키지 파일의 위치가 남는다. 외부 입력에는 원래
activity 선언의 위치가 붙는다. 파일 해시, 안정적인 ID, 바이트 범위로 같은 출처를
다시 찾을 수 있다. 조건문이 없는 반환식도 최초 파일의 후보 조각과 연결된다.

## 여러 정수 입력을 받는 후보 작성

`value_case`의 입력 배열은 선언 순서대로 대응한다. 다음 형태는 기존 정수 평가기를
사용하며 모델을 설정하지 않아도 후보의 선언 순서와 사례 결과로 구성을 정한다.

```gooo
package difference
namespace difference6763
entity Integer id "urn:gooo:type:integer"
activity Difference(Integer,Integer) -> Integer id "urn:difference6763:diff" computes "return __GOOO_BODY_HOLE_expr__" assembling {
 source_fill intent "첫 수에서 둘째 수를 뺀다" {
  hole "expr"
  candidate "sum6763" { fill "expr" "input0 + input1" }
  candidate "difference6763" { fill "expr" "input0 - input1" }
 }
 value_case "[2,8]" -> "-6"
 holdout_value_case "[100,4]" -> "96"
}
```

현재 이 경로는 정수 입력 1~16개, 정수 출력, 명시적으로 적은 후보를 지원한다.
값은 부호 있는 64비트 정수로 읽는다. 입력 개수 오류, 소수, null, 범위 초과는 진단한다.
입력 배열을 이용해 후보 문법을 자동으로 파생하는 기능은 후속 범위다.

이 프로그램의 관찰은 선택 사례 1/1, 별도 보류 사례 1/1, 추가 실행 입력 2/2였다.
서로 다른 용도의 사례를 따로 집계한다. [원본과 결과](../observations/value-transfers/difference6763.json)를
함께 보관한다.

## Gooo 안에서 출처를 설명하기

`ProgramOriginFacts`는 경로와 출처 위치의 쌍을 스칼라 필드가 있는 레코드로 전달한다.
`origin_kind`, `candidate_id`, `call_returns`, `bind_steps`, `source_view`와 원본 위치를
Gooo 활동에서 읽을 수 있다. `call_returns`와 `bind_steps`는 표시한 경로의 구조를 센다.

[실제 소비 프로그램](../observations/value-transfers/consumer6781.json)은 저장된 기원 3행의
위치 4건을 읽어 “후보가 패키지 연결을 거쳐 전달됨”, “현재 활동의 외부 입력”처럼 설명한다.
네 결과가 예상과 일치했다. 원래 프로그램은 재실행하지 않았다.
보고서 전체의 `analysis`, `issues`, `truncated`는 사실 레코드와 함께 해석해야 한다.

## 분석 범위

조회는 선택된 소스의 가능한 의존 관계를 읽고 분기 양쪽을 합친다. 실제 입력이 지나간
경로는 실행 기록에 별도로 남는다. 하나의 출발점과 기원마다 최단 경로 하나를 보여준다.
재귀, 일부 반복·종료 경로, 펼치지 못한 호출은 `PARTIAL`과 이유를 남긴다.

소스 128 KiB, activity 128개, 전체 노드와 방문 8,192개, 호출 깊이 16개,
출력 경로 합계 2,048단계로 제한한다. 과거 기록에 원본 파일 정보가 없으면 그 위치는
생략한다. 기원 수집이 한도를 넘으면 실행 기록의 `fill_origin_issue`에 이유가 남는다.

이 기능의 [실행 관찰과 초기 실패](../observations/value-transfers/README.md)는 범위별로
읽을 수 있다. 작은 결정 모델의 제공자·모델 파일 기원 연결, 컬렉션과 중첩 값의 표현,
제3자의 첫 사용 경험은 이어서 다룰 과제다.
