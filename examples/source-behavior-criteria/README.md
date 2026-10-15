# 반품 처리의 조건을 Gooo에 적기

반품 처리는 금액의 부호, 승인 여부, 설명 보존을 함께 만족해야 합니다.
이 예제는 각 조건을 Gooo 활동으로 적고, `criteria`에서 조립할 바디와 연결합니다.
정답 레코드를 손으로 나열하지 않아도 원래 입력과 결과의 관계를 검사할 수 있습니다.

개발 버전 컴파일러로 실행합니다.

```sh
go run ./cmd/gooo body-codegen --json --activity Decide examples/source-behavior-criteria/main.gooo.fixture
go run ./cmd/gooo body-compose --source examples/source-behavior-criteria/main.gooo.fixture --entry Decide --cases examples/source-behavior-criteria/evaluation.json --out /tmp/gooo-behavior-criteria-example
```

선택 입력은 승인된 금액 37이고, 조건은 세 개입니다. 후보는 금액·승인·설명의
두 가지 표현을 조합한 여덟 개입니다. 한 후보가 일부 조건만 만족해도 각 결과를
남깁니다. 계산할 수 없었던 조건은 `unknown`으로 구분합니다.

금액 73은 후보 선택 뒤 평가합니다. 별도 파일의 금액 113과 127은 생성된 Go를
실행하는 사례이며, 소스 내부 조건 검사와 다른 분모로 보고합니다.

기존에 사용자가 직접 작성해야 했던 평가 조합과 결과 집계가 `criteria` 선언으로
옮겨졌습니다. 조건을 어떻게 정의했는지는 소스에서 읽고 고칠 수 있습니다.
이 예제의 성공 수는 이 입력과 조건에 한정됩니다.

자세한 의미와 한도는 [기능 문서](../../docs/source-behavior-criteria.md)에 있습니다.

첫 개발 실행에서 후보별 충족 수는 `0, 1, 1, 2, 1, 2, 2, 3 / 3`이었습니다.
최종 후보는 선택 조건 3/3, 별도 입력 조건 3/3, 추가 네이티브 출력 2/2를
만족했습니다. 모델 호출은 0회입니다. [실행 요약](observation.json)에 소스와
개발 컴파일러의 식별값을 남겼습니다. 이 실행에서 지연시간·CPU·모델 정확도를
측정하지 않았습니다.
