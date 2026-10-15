# 줄을 바꿔 쓰는 조건문과 원래 위치

개발 소스에서는 닫는 중괄호와 `else` 사이에 공백·탭·줄바꿈을 둘 수 있습니다.
예제는 한글 변수 이름과 두 조건을 사용합니다.

```gooo
let 조정값 = input.amount
if input.enabled {
    조정값 = 조정값 + 7
}
else {
    조정값 = 조정값 - 5
}
```

```sh
gooo package execute --cases examples/else-layout/cases.json \
  examples/else-layout/gooo.workspace.json
gooo body-origins --activity Adjust --format text examples/else-layout/main.gooo
```

컴파일러는 공백 구간의 줄바꿈을 Go 파서가 받을 수 있게 바꾸고 원래 바이트
위치를 유지합니다. 원고를 고쳐 읽더라도 원래 문장의 위치로 돌아갈 수 있게
하는 작은 변환입니다. 문자열·주석 내용과 작성자가 넣은 세미콜론은 보존합니다.
주석이 중괄호와 `else` 사이에 있는 경우는 기존 같은 줄 표현을 사용합니다.

허용 조건은 [Gooo 정책](../../internal/meta/bodysurface/else.gooo)에 있습니다.
`LayoutFacts`는 토큰 종류, 자동 삽입된 줄 구분, 닫는 중괄호, 공백 구간을
이름과 의미 ID가 있는 네 필드로 구분합니다. `ClassifyElse`가 판단하고,
`JoinElse`가 Go 호스트의 관찰값을 Gooo 레코드로 조립합니다.
실행 파일에는 그 정책에서 생성한 Go 코드가 들어갑니다.

## 첫 관찰의 범위

| 항목 | 결과와 분모 |
| --- | --- |
| 이름 있는 정책 입력 | Boolean 필드 4개의 조합 16/16 |
| 예제의 기능 | 새 입력 8/8, 경계값 41 포함 |
| 원래 조건 위치 | 분기 노드 4/4가 원래 UTF-8 소스와 일치 |
| 전체 노드의 원래 표현 위치 | 28/30; 암묵적인 입력 포트 2개는 `INPUT_PORT_IDENTITY` 유지 |
| 정적으로 참·거짓이 정해진 분기 | 0/4; 나머지 4개는 실행 입력에 따라 달라짐 |

각 첫 관찰의 모델 호출은 0회입니다. 위 숫자는 서로 다른 범위를 설명합니다.
이 예제는 주어진 입력에서 값을 계산하면서, 실행 전에 모르는 조건은 그대로
남기는 흐름을 보여줍니다.

Gooo가 이 프로그램의 그래프를 읽는 사례도 있습니다.

```sh
gooo package execute --cases examples/else-layout/source-control-case.json \
  examples/source-control-reader/gooo.workspace.json
```

이 파일은 첫 소스 조회의 입력을 보관합니다. CI가 현재 예제 소스와 그래프를
다시 대조합니다. 새 소스 조회를 연결하려면 다음 명령을 사용합니다.

```sh
gooo body-origins --activity Adjust --inputs-for source/control:Inspect \
  --inputs-shape source-control-columns examples/else-layout/main.gooo > /tmp/else-origins.json
gooo package execute --inputs /tmp/else-origins.json \
  examples/source-control-reader/gooo.workspace.json
```

[첫 관찰 기록](../../docs/observations/else-layout/README.md)에서 소스·생성 코드·실행
범위를 확인할 수 있습니다. 이 기능은 개발 버전이며 공개 0.6.34 설치 안내와
구분합니다.
