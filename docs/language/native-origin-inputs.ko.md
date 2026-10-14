# 기원을 다음 Gooo 프로그램의 입력으로

개발 소스의 `package execute --origin-receipt`는 저장된 네이티브 실행 기록에서
기원 행을 골라 워크스페이스의 진입 활동으로 전달한다. 기록의 해석 규칙을 Gooo로
작성하는 진입점이다. 예를 들어 원소 계산의 중단을 분류하거나, 원래 입력과 새로
생성한 목록 사이의 연결을 설명할 수 있다.

[실행 가능한 입력 예제](../../examples/native-origin-reader/README.md)에서 시작한다.

## 입력과 순서

`--cases`, `--inputs`, `--construction-receipt`, `--origin-receipt` 중 하나를 선택한다.
기원 입력은 기존 `gooo/body-composition-inputs/v1` 형식으로 전달되며 기대값은 없다.
행의 키는 워크스페이스의 `package_path:activity`다. 각 예제의 입력 entity가 실제
스칼라 필드 전체를 선언한다. 종류를 바꾸면 그에 맞는 entity와 활동을 선택해야 한다.

- `reads`: 사례, 읽기 위치, 입력 또는 생성 원소의 기원, 원본 소스 범위.
- `constructions`: 사례 안의 생성 인스턴스, 선언한 원소 수, 시도한 수, 빈 목록과 완료 여부.
- `elements`: 생성 인스턴스 안에서 시도한 원소와 계산 중의 읽기 구간.
- `links`: 사례, 읽기 인덱스, 생성 인스턴스 인덱스, 원소 인덱스.

순서는 저장된 사례 → 기록 순서다. `elements`는 생성 인스턴스 순서 아래에 원소가
이어진다. 중첩 생성의 행은 인스턴스별로 묶이며, 원래 시간 관계는 읽기 구간과 링크로
읽는다. 인스턴스와 읽기 인덱스는 사례마다 다시 시작한다.

## 페이지와 부분 기록

`--origin-kind` 기본값은 `elements`, offset은 0, limit는 32다. offset은 일반 종류에서
0..65536, `local-fill-fragments`에서는 0..2147483647이다. limit는 1..128이다.
선택한 페이지의 JSON 입력 한도는 32KiB다. 넘으면 limit를
줄여야 한다. 내용이나 표현식을 임의로 자르지 않는다. 빈 페이지에는 kind, offset,
전체 행 수를 포함한 오류를 반환한다.

결과의 `origin_input`에는 읽은 파일 전체의 SHA-256, 원래 선택 소스와 사례 묶음의
해시, 전체·반환 행 수, 다음 offset과 원래 관찰 요약이 들어간다. `truncated`는
현재 페이지 뒤에 남은 행이 있다는 뜻이다. 앞에서 건너뛴 행 수는 `query.offset`이다.
누락된 읽기·등록·생성 수와 중단 수는 원래 요약의 수치를 유지한다.

## 조립 기원을 배열 입력으로 전달하기

기본 `--origin-shape rows`는 행마다 새 입력을 만든다. `local-fill-columns`는
`local-fill-fragments` 페이지를 같은 길이의 배열로 묶어 입력 하나를 만든다.
kind를 생략하면 이 조각 종류를 선택하며, 다른 kind와 함께 요청하면 오류를 반환한다.

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt saved-execution.json --origin-shape local-fill-columns \
  examples/derived-origin/coverage.workspace.json
```

입력에는 원래 선택 소스 해시, 기록된 대입 수, 페이지·기록 완료 여부와 여섯 배열
`case_indexes`, `assignments`, `kinds`, `source_shas`, `starts`, `ends`가 들어간다.
Go는 행을 배열로 옮긴다. [Gooo 프로그램](../../examples/derived-origin/coverage.gooo)이
같은 대입을 묶고 조립 기원 연결 비율을 계산한다.

`origin_input.total_rows`와 `returned_rows`는 원래 조각 행의 수를 유지한다.
이 형태는 `input_shape=gooo/local-fill-columns/v1`, `consumer_rows=1`을 덧붙인다.
중간·마지막 페이지는 전체 비율을 미확정으로 남긴다. offset 0에서 조각이 하나도
없으면 빈 배열 입력 하나를 전달한다. 범위를 벗어난 offset은 오류다.
기본 행 형태의 빈 페이지 오류 동작은 유지한다.

분모는 기록된 스칼라 대입, 분자는 조립 기원 조각에 연결된 대입이다. 직접 작성한
대입은 일반 소스 위치를 가지면서도 조립 기원 조각은 없을 수 있다. 입력 예제와
[실제 관찰](../observations/origin-column-input/README.md)에 이 범위를 구분해 적었다.

## 받아들이는 저장 형식

직접 저장한 `CompositionRuntime` v1/v2/v3, `body-compose` 결과,
`body-construct` 결과, `JointEvaluation`, `package execute` 결과를 읽는다.
이력과 후보 시도 목록이 있는 경우 최신 `runtime`만 선택한다. 읽기와 생성 관련
집계·참조가 맞아야 하며, 지원하지 않는 형식은 오류로 알린다.
생성 기록이 없는 v1 읽기 요약에는 `reads`를 사용한다.

파일을 읽는 동안 원래 프로그램·후보·모델을 실행하는 횟수는 0이다. 새 워크스페이스는
기존 실행 경로를 사용한다. 이 예제들은 모델 설정 없이 결정론적으로 실행된다.

기록의 내부 참조 검사는 보존된 기록을 읽기 위한 검사다. 원래 실행의 진위나 전체
프로그램의 정확성까지 새로 판정하는 범위는 포함하지 않는다. 원소의 읽기 구간도
그 계산 중 발생한 읽기를 나타낸다. 앞서 계산된 지역 변수의 전체 영향 연결은
후속 분석 범위로 남아 있다.

[이번 관찰과 초기 실패](../observations/native-origin-inputs/README.md)도 함께 공개한다.
