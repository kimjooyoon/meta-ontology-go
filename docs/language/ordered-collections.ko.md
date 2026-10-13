# 목록을 계산하고 그 기원을 이어 보기

개발 소스의 EntityFields V5 기능이다. 작은 주문 묶음, 이름 목록, 여러 판정값을
Gooo의 레코드 필드로 받고 순서대로 계산할 수 있다. 계산 방식은 `computes`에 쓰고,
바꿔 볼 조건식은 `source_fill`의 후보로 선언한다.

```gooo
package totals
namespace totals
entity Integer id "urn:gooo:type:integer"
entity Basket id "urn:totals:basket" fields {
    field amounts id "urn:totals:amounts" type integer required many
}
activity Sum(Basket) -> Integer id "urn:totals:sum" computes `
    let total = 0
    for amount in input.amounts {
        if amount >= 20 { total = total + amount }
    }
    return total
`
```

`many`는 입력 순서를 가진 목록이다. 위 프로그램에 `{"amounts":[21,-3,5]}`를
주면 21을 계산한다. 출처 조회에서는 `amounts` 선언, 반복문에서 꺼낸 값,
조건식과 합계를 이어 본다. 명세의 항목마다 원래 붙어 있던 꼬리표를 계산 과정에도
가져가는 방식이다.

## 사용할 수 있는 표현

| 표현 | 의미 |
| --- | --- |
| `type integer required many` | 반드시 제공하는 정수 목록 |
| `type string required many` | 순서를 유지하는 문자열 목록 |
| `type boolean optional many` | 생략하거나 빈 목록 또는 불리언 목록으로 제공하는 필드 |
| `for item in input.items { ... }` | 항목을 순서대로 읽으며 계산 |
| `len(input.items)` | 목록 길이. Integer 반환에는 `int64(...)`를 사용 |
| `input.items[index]` | 해당 위치의 값을 읽음 |
| `input.items[low:high]` | 순서를 유지하는 부분 목록. 한쪽 경계 생략 가능 |
| `Result{items: input.items}` | 원래 목록을 결과 레코드에 전달 |
| `[]int64{input.amount, 0}` | 정수 목록을 본문에서 구성 |
| `[]string{"기본", input.name}` / `[]bool{true, false}` | 문자열·불리언 목록을 구성 |

한 필드의 상한은 256개다. `[]`는 제공된 빈 목록이고, optional 필드를 생략하면
값이 없는 상태를 유지한다. `null`과 중첩 목록은 허용하지 않는다. 정수는 int64로
읽어 2⁵³을 넘는 값도 보존한다. 인덱스와 범위는 필요한 조건문으로 확인한다.

반복 변수는 읽기 전용이며, 합계 같은 `let` 지역 변수에는 대입할 수 있다.
원소 대입, 목록 증가, 중첩 반복은 현재 지원 범위에 포함되지 않는다.
목록 선언이 있는 소스는 V5를 사용하고, 기존 단일 값 소스는 기존 프로필을 유지한다.

## 조건식의 후보를 Gooo에 두기

[실행에 사용한 조립 예제](../../examples/ordered-collections/threshold.gooo.fixture)는
33 이상인 항목만 합산하고 원래 목록도 반환한다. 두 후보의 조건식과 선택용 사례,
별도 보류 사례가 같은 소스에 있다. 컴파일러는 타입을 확인한 후보를 유한한 입력으로
계산하고 결과를 조립한다. 모델 연결이 없으면 기존 결정론적 선택 경로를 사용한다.

개발 소스로 설치한 CLI에서 생성 결과를 읽을 수 있다.

```sh
gooo body-codegen --json --activity Compute \
  examples/ordered-collections/threshold.gooo.fixture
```

실행 가능한 본문은 `body-compose`와 `package execute`에서 사용한다.
일반 `check`와 `generate`도 목록 필드와 명시적인 bind를 처리한다.
Go 필드의 표현은 `[]T`, optional 필드는 `*[]T`이다.

## 출처와 관찰의 범위

`for … in`을 Go 반복문으로 옮길 때 코드 길이가 변한다. 조건식, 호출 이름,
레코드 생성자와 선택한 후보의 위치는 원본 Gooo 바이트 위치로 다시 연결한다.
한글 이름, 다른 패키지의 함수 호출과 이름 변경도 이 위치 연결을 사용한다.

값의 정적 기원 분석은 여러 반복의 원소를 합쳐 연결한다.
`COLLECTION_ELEMENTS_JOINED`와 `LOOP_EXIT_PATHS_NOT_EXPANDED`를 남기며,
입력별로 어느 원소가 어떤 분기를 지났는지는 별도의 실행 관찰에서 읽는다.

[실험 기록](../observations/ordered-collections/README.md)에는 후보 선택용 결과,
별도 보류 결과, 네이티브 실행 결과와 첫 실패를 각각 기록했다.
독립 사용자 경험, 전체 언어 완성도와 모델 효과는 별도의 측정 과제다.

## 작은 도구로 사용하기

[작업 예산 계획](../../examples/batch-planner/README.md)은 목록을 순서대로 읽고,
예산을 넘기거나 잘못된 시간이 나오면 부분 결과와 중단 이유를 반환한다.
소스에 적힌 비교식 후보와 사례를 이용해 경계값 조건을 조립한다.

[작업 목록 전달](../../examples/batch-handoff/README.md)은 계산한 경계로 목록을
자르고 `bind`를 통해 다음 활동에 넘긴다. 다음 활동에서 같은 목록을 집계하므로
선택 결과를 별도 Go 코드로 다시 구성하지 않아도 된다. 두 예제는 문법 검사 목록에
V5 사례로 등록해 서식과 의미의 왕복도 확인한다.

## 실제로 읽은 원소를 입력에 연결하기

개발 빌드의 `body-compose`와 `package execute`는 목록 인덱스와 `for … in`의 실제
읽기를 runtime의 `read_sites`, `read_inputs`, `traces[].reads`에 기록한다. 예를 들어
원본 목록을 `[1:4]`로 자른 뒤 1번 원소를 읽으면, 원본 입력의 2번 원소로 연결된다.
같은 값이 반복돼도 실제 위치를 유지한다. 분기에서 생략된 읽기는 기록되지 않는다.

새 목록 리터럴에는 목록·원소별 생성 표현과 실제 생성 번호를 기록한다.
생성 목록을 자르거나 다음 활동에 넘겨도 원소를 읽은 기록에서 생성 위치로 이어진다.
각 원소를 계산하는 동안 발생한 목록 읽기는 별도의 구간으로 연결한다.
앞서 지역 변수에 저장한 값의 전체 데이터 의존성은 추가 개발 과제다.
입력 원소 연결률과 예상 출력 일치율은 각각의 분모로 읽는다. 누락과 미연결 수는
`read_observation`에 있다. 한 번의 실행에서 기록할 읽기 상한은 65,536개다.

[실행 기원 관찰](../observations/native-read-origins/README.md)에 중복 값, 빈 구간,
문자열·불리언, 중첩 인덱스, optional 입력 및 저장된 기록을 해석하는 Gooo 프로그램이 있다.

[목록 생성 기원](../observations/construction-origins/README.md)은 중첩 호출, 빈 목록,
계산 중단, 후보로 채운 원소 표현과 Gooo가 자신의 생성 기록을 설명하는 예제를 담는다.
