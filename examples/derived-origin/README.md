# 조립한 식의 출처를 읽고, 연결된 비율 계산하기

[main.gooo](main.gooo)는 작은 정수 문법과 두 입출력 사례에서 보정식을 고른다.
이번 관찰에서는 `input + 557`이 선택됐다. 모델을 설정하지 않은 결정론 경로다.
후보 6개를 모두 살펴봤다는 기록과, 선택에 쓰지 않은 입력 3개의 실행 결과가
따로 남는다.

저장소 루트에서 실행한다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/derived-origin/cases.json \
  examples/derived-origin/gooo.workspace.json > derived-execution.json

go run ./cmd/gooo package execute --json \
  --origin-receipt derived-execution.json --origin-kind local-fill-fragments \
  --origin-limit 6 examples/native-origin-reader/local-fill/reader.workspace.json
```

첫 대입의 조각은 `grammar`, 복사 대입의 조각은 `template`이다. 문법에서 만든 식에는
원문에 적힌 후보 표현의 위치를 붙이지 않는다. 대신 `integer-offset-constant/v1`을
선언한 위치가 남는다. 읽기 단계에서 원래 생산 프로그램과 모델을 호출한 횟수는 0이다.

[coverage.gooo](coverage.gooo)는 저장된 조각을 열 형태로 받아 출처 연결 비율을
계산한다. 계산 규칙도 Gooo의 조건문, 반복문, 변수와 배열 접근으로 작성했다.
`case_index`와 `assignment_index`가 같으면 여러 조각이 있어도 대입은 한 번 센다.

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/derived-origin/coverage-cases.json \
  examples/derived-origin/coverage.workspace.json
```

제공한 열 자료는 위 실험의 저장 기록에서 추출한 입력과, 별도로 표시한 모순·누락
입력이다. 이 예제는 이 입력 자료를 계산하며, 새로운 실행 기록을 자동으로 열로
바꾸는 범용 변환기는 포함하지 않는다.

| 입력 범위 | 관찰한 출력 |
| --- | --- |
| 실제 저장 기록의 전체 페이지 | 대입 6개 중 6개 연결, 10000 basis points(100%) |
| 같은 기록의 앞 3개 조각 | `partial-page`, 전체 비율은 `-1`(미확정) |
| 출처 하나를 비운 입력 | 대입 6개 중 5개 연결, 8333 basis points(약 83.33%) |
| 같은 대입의 조각을 한 번 더 준 입력 | 조각은 7개, 연결된 대입은 여전히 6개 |
| 길이가 다르거나 순서가 맞지 않는 열 | `unusable-input` |

분모는 저장된 대입이다. 비율은 계산 결과의 정답률이나 자연어 의도 전체의 충족률을
뜻하지 않는다. 완전한 페이지와 완전한 기록을 모두 받았을 때만 전체 연결 비율을
계산한다. [관찰 자료와 범위](../../docs/observations/derived-origin-example/README.md)에
소스, 실행 결과, 입력 추출 출처를 남겼다.
