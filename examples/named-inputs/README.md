# 이름으로 읽고, 위치로 연결하는 입력

개발 소스에서 다음처럼 계산의 방향을 이름으로 적을 수 있습니다.

```gooo
activity Difference(minuend Integer, subtrahend Integer) -> Integer
  id "urn:gooo:example:named-inputs:difference" computes `
return minuend - subtrahend
`
```

실제 실행 파일은 [main.gooo](main.gooo)입니다. `Credit`이 원금에 23을
더하고, `Net`이 그 결과와 외부 공제액을 `Difference`에 전달합니다.

저장소 루트에서 실행합니다. 현재 설치된 릴리스에는 이 개발 문법이 없을 수
있으므로 아래 명령은 체크아웃의 컴파일러를 사용합니다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/named-inputs/cases.json examples/named-inputs/workspace.json
go run ./cmd/gooo body-origins --activity Difference --format text \
  examples/named-inputs/main.gooo
go run ./cmd/gooo package interface --json examples/named-inputs/workspace.json
```

첫 실행에서 별도 입력 사례 2개, 출력 조건 4개가 모두 맞았고 모델 호출은
0회였습니다. 로컬 명령 벽시계 시간은 1.109초였습니다. 이 사례에 관한
관찰이며 컴파일러 전체의 성능이나 정답률을 나타내지는 않습니다.

출처 보기에는 `input0`의 타입이 원문 7행 29열의 `Integer`, 본문 이름이
7행 21열의 `minuend`로 나타납니다. 타입이 같은 두 입력도 원래 위치를
각각 보관합니다.

`bind Credit.result -> Net.input0`과 외부 입력 키 `named/inputs:Net.input1`은
선언 순서로 연결됩니다. 본문 이름은 `원금`, `공제`입니다. 입력 이름을
바꾸는 작업이 활동의 안정된 ID나 연결 위치를 새로 만들지는 않습니다.

[문법과 지원 범위](../../docs/language/named-inputs.ko.md)를 참고하세요.
