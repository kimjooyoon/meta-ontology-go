# 입력과 출력의 Text 크기

개발 중인 EntityFields V6 예제입니다. `Prefix`는 입력 앞에 `prefix`를 붙입니다.
입력 Text는 기존 1,024바이트 범위를 사용하고, 반환 레코드의 `value`는 원문에서
1,030바이트까지 허용합니다.

```gooo
field value id "gooo://examples/bounded-text/expanded/value"
    type string required one max_utf8_bytes "1030"
```

V6를 지원하는 컴파일러로 저장소 루트에서 실행합니다.

```sh
gooo package execute --json --inputs examples/bounded-text/inputs.json \
  examples/bounded-text/workspace.json
```

크기는 JSON을 푼 뒤의 UTF-8 바이트입니다. 한글 한 글자는 보통 3바이트이며
`\u0000`은 전송할 때 6바이트, 값에서는 1바이트입니다. 원문에 선언한 조건이
준비·관측·재실행 경로의 필드 식별자와 함께 전달됩니다.

- 조건을 생략하면 기존 실행 프로필의 제한을 유지합니다.
- `max_utf8_bytes "0"`은 빈 Text만 허용합니다. 미선언과 구분합니다.
- `optional` 필드는 생략할 수 있습니다. `many`에서는 각 원소를 검사합니다.
- 조건은 레코드가 Gooo 실행 경계로 들어오거나 결과로 나올 때 검사합니다.
- 전체 JSON·출력·기록의 자원 예산도 각각 적용됩니다. 큰 필드 선언만으로
  전체 문서의 예산이 늘어나지는 않습니다.

Go로 투영한 구조체의 태그는 조건을 설명합니다. 실제 값의 수용 여부는 Gooo
실행 경로에서 판정합니다. 표시용 렌즈나 용량 힌트는 조건을 바꾸지 않습니다.
