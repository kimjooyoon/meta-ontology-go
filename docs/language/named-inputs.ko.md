# 활동 입력에 이름 붙이기

`Difference(Integer, Integer)`만 보면 어느 값을 빼는지 본문까지 내려가서
`input0`과 `input1`을 확인해야 합니다. 개발 소스는 각 타입 앞에 본문에서
사용할 이름을 적는 문법을 제공합니다.

```gooo
activity Difference(minuend Integer, subtrahend Integer) -> Integer
  id "urn:example:difference" computes `return minuend - subtrahend`
```

한국어 이름과 파일의 import 별칭도 사용할 수 있습니다.
`Quote(원금 money.Amount, 할인 Integer)`에서는 이름과 타입을 각각 읽습니다.
이름이 없는 자리도 함께 쓸 수 있어서 `Scale(Amount, factor Integer)`의
본문에서는 `input0`과 `factor`를 사용합니다.

## 서로 다른 세 가지 정보

| 정보 | 예 | 보관하는 의미 |
| --- | --- | --- |
| 본문 이름 | `minuend` | 해당 본문에서 값을 가리키는 이름 |
| 입력 위치 | `input0` | 활동 연결과 외부 입력의 선언 순서 |
| 타입 ID | `urn:gooo:type:integer` | 그 입력에 허용된 타입의 정체성 |

이름을 붙인 입력은 본문에서 그 이름을 사용합니다. 같은 자리를 부르는
두 번째 본문 별칭을 자동으로 추가하지 않습니다. 한 입력의 이름으로 다른
자리의 `input0`, `input1`을 빌리거나 이름을 중복하면 진단합니다.
입력은 읽기 전용이며, 계산 중 변경할 값은 `let`으로 선언합니다.

활동 사이의 `bind`와 JSON 입력 키에는 기존 위치 이름을 사용합니다.
입력이 하나면 `input`, 여러 개면 `input0`, `input1` 순서입니다.
`bind Producer.result -> Difference.minuend` 형태는 현재 지원 범위 밖입니다.

## 출처가 이름을 따라가는 방법

컴파일러는 이름과 타입의 UTF-8 바이트 범위를 각각 원문에 연결합니다.
`body-origins` JSON에는 기존 입력 위치와 함께 `declaration.name`,
`name_start_byte`, `name_end_byte`가 추가됩니다. 텍스트 보기에는 두 위치를
따로 표시합니다. 활동을 한국어로 이름 변경한 뒤에도 현재 파일의 해시와
새 위치를 읽습니다.

패키지 인터페이스의 `input_names`는 선언한 이름의 순서입니다. 혼합 문법의
이름 없는 자리는 빈 문자열이며, 전부 이름이 없으면 이 필드는 생략됩니다.
기존 `input-declaration-columns/v1` 내보내기는 타입과 위치를 다루는 11개
열을 유지합니다. 이름의 원문 정보는 `body-origins` JSON에서 읽습니다.

이름은 PROV-O의 안정된 역할 ID를 자동으로 선언하지 않습니다. 현재 기원
정보는 원문, 활동 ID, 입력 순서, 타입과 필드 ID를 연결합니다. 역할을 언어에
도입할 때에는 역할의 선언, 참조, 변경 규칙을 별도로 정해야 합니다.

## 현재 확인한 범위

- 파싱·서식·복사·의미 IR 왕복에서 이름과 원문 위치를 보존합니다.
- Go 본문 생성, 다른 패키지의 활동 호출, 위치 기반 `bind`에 연결합니다.
- 같은 타입의 두 입력, 한국어 이름, 레코드 입력, 혼합 문법과 조건 분기를 다룹니다.
- 입력 이름이 함수나 import 별칭을 가리면 해당 호출을 거절합니다.
- 이름을 사용하는 본문 변경은 영향 분석에 나타납니다. 안정된 활동 ID는 유지됩니다.

[실행 예제](../../examples/named-inputs/README.md)와
[첫 관찰 기록](../observations/named-inputs/observations.json)을 제공합니다.
이 기능은 개발 소스의 구현이며, 배포된 릴리스 전체의 지원을 주장하지 않습니다.

## 선언 사실을 Gooo에서 사용하기

`body-origins --inputs-shape named-input-declaration-columns`는 작성한 이름,
타입, 입력 순서와 각각의 원본 바이트 범위를 Gooo 프로그램의 입력으로
내보냅니다. [DescribeInputs 예제](../../examples/source-control-reader/README.md#gooo로-입력-설명-만들기)는
이 사실로 사람이 읽는 입력 설명과 이름의 존재 비율을 조립합니다.
기존 `input-declaration-columns`의 11개 열은 v1으로 유지하고,
이름을 포함하는 14개 열은 v2로 구분합니다.
