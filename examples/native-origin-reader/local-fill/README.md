# 대입 표현의 조립 출처 읽기

이 Gooo 프로그램은 저장된 대입 기록의 출처 종류를 읽는다. 직접 선언한 후보는
`candidate`, 문법으로 조립한 표현은 `grammar`, 유지한 템플릿은 `template`으로
구분한다. 분류 규칙은 [reader.gooo](reader.gooo)에 있다.

저장소 루트에서 실행한다.

```sh
go run ./cmd/gooo package execute --json \
  --origin-receipt examples/native-origin-reader/local-fill11591-runtime.json \
  --origin-kind local-fill-fragments --origin-limit 2 \
  examples/native-origin-reader/local-fill/reader.workspace.json
```

예제의 첫 두 행은 `candidate`, `template`이다. 입력으로 삼은 기록은 새 프로그램의
별도 실행 입력 2개에서 얻었으며, 읽는 동안 원래 생산 프로그램을 다시 실행하지
않는다. CLI 읽기는 기대값을 전달하지 않은 실행이어서 유한 정답률 분모는 0이다.
위 두 출력은 별도의 CLI 검사에서 직접 대조했다.

`case_index`와 `assignment_index`가 어느 실행 사례의 대입인지를 가리킨다.
`completed`는 대입 완료 여부다. `source_start/end`는 조립 전 입력 문서에 있는
후보·문법·템플릿 조각의 위치이며, `selected_decoded_start/end`는 조립된 본문의
대입 우변 위치다. 템플릿 조각은 대입 우변보다 넓을 수 있다.

현재 CLI에서 관찰한 종류는 후보와 템플릿이다. 문법 분류 분기는 입력 규칙으로
제공되며, 실제 문법 조립 경로는 별도 사례로 확인할 범위다.
