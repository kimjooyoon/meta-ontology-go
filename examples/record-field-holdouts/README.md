# 고른 코드가 다른 입력에서도 맞을까

`Choose`는 입력을 그대로 반환하거나 부호를 바꾸는 두 표현 중 하나를 고릅니다.
선택 사례가 0 하나면 어느 쪽도 구별되지 않습니다. 별도 평가 입력 5에서는 차이가
드러나지만, 이 점수로 다시 고르지는 않습니다. 처음의 선택 1/1과 별도 평가 0/1을
각각 남깁니다.

호출자인 `Main`의 요구까지 사용하면 부호를 바꾸는 쪽이 선택됩니다. 그 뒤 같은
평가 입력을 최종 바디에 적용한 결과는 1/1입니다. 두 시점의 바디와 평가를 따로
읽을 수 있습니다. 개발 소스에서 빌드한 CLI로 실행합니다.

```sh
gooo body-codegen --json --activity Choose examples/record-field-holdouts/source.gooo.fixture

gooo body-construct --source examples/record-field-holdouts/source.gooo.fixture \
  --entry Main --construction-cases examples/record-field-holdouts/construction.json \
  --cases examples/record-field-holdouts/evaluation.json \
  --attempts 2 --format text --out /tmp/gooo-record-held
```

평가 답을 바꾸면 평가 결과가 달라집니다. 선택에 사용하는 사례와 모델 입력은
그대로입니다. 레코드 전체가 맞은 수와 개별 필드가 맞은 수도 구분합니다.
호출자 입력이 2라면 `Choose`에는 5가 전달되므로, 이 입력은 호출자 조립에서
이미 사용됐다고 표시됩니다.

`continuation.gooo.fixture`는 `Seed`와 `Wrap`의 조립을 Gooo 정책으로 이어가는
예제입니다. 먼저 한 후보에서 저장한 결과를 이어 만들면 `Seed`가 바뀌고,
`Wrap`의 이전 후보를 새 보조 함수로 재검사합니다. 두 함수의 별도 평가가
각각 0/1에서 1/1로 바뀌며, 이전 조립 기록의 점수는 그대로 남습니다.
`continuation-cases.json`은 그 프로그램의 별도 실행 입력 두 개입니다.

[평가 범위와 저장 결과 규칙](../../docs/record-field-holdouts.md)을 참고하세요.

## 조건에 따라 레코드를 갱신하기

`refund.gooo.fixture`는 주문의 금액·설명·승인 여부를 담은 레코드를 사용합니다.
반품 승인이 나면 금액의 부호를 바꾸어야 하지만, 선택 사례는 금액 0 하나입니다.
이 사례로는 두 후보를 구별할 수 있어 첫 후보가 유지됩니다.

```sh
gooo body-codegen --json --activity Decide examples/record-field-holdouts/refund.gooo.fixture
```

별도 평가에는 승인된 30원과 승인되지 않은 40원 입력을 넣었습니다. 최초 바디는
승인 여부와 설명을 보존하지만 승인된 금액의 부호를 바꾸지 못합니다.
그 결과 전체 사례는 `1/2`, 개별 필드는 `5/6`입니다. 이 차이로 남은 문제를
읽을 수 있습니다. 선택 사례와 평가 사례의 정답은 Gooo 소스에 그대로 남습니다.
이 예제는 유한한 선언과 결과를 설명하며 실제 결제 처리 범위를 다루지 않습니다.

JSON의 `report.decision`은 코드 생성 단계의 결과입니다.
`report.record_assembly`의 점수는 선택 사례를,
`report.record_assembly.holdout`의 점수는 별도 평가 사례를 설명합니다.
생성된 프로그램을 실행한 결과까지 함께 읽으면 어떤 입력에서 문제가 남았는지 확인할 수 있습니다.

선택 규칙이나 평가 정답을 고친 경우에는 `gooo impact 이전.gooo 이후.gooo`로
조립 계약이 바뀐 선언을 찾을 수 있습니다. 이 예제의 평가 정답 하나를 바꾸면
`Decide`가 재확인 대상으로 표시되고, 선택 사례 1개와 평가 사례 2개가 구분되어 남습니다.
