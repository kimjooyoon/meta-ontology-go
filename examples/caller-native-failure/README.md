# 호출할 때 드러나는 계산 실패도 조립 기록으로 남기기

`PlanBudget`은 지역 사례와 별도 홀드아웃을 통과하는 후보를 가지고 있습니다.
그러나 `used == limit`인 실제 호출에서 0으로 나눕니다. 앞의 다섯 후보와 모든
원래 기대값을 유지하고, 마지막 정상 후보 앞에 이 실패 후보를 넣었습니다.

이 예제는 0.6.15 이후의 개발 소스에서 실행합니다. 공개 0.6.15는 이 호출에서
종료됩니다. [원래 실패 관측](https://raw.githubusercontent.com/wiki/kimjooyoon/meta-ontology-go/observations/caller-native-failure-20261009/README.md)을 보존했습니다.

```sh
go build -trimpath -o .gooo ./cmd/gooo
./.gooo body-construct --source examples/caller-native-failure/source.gooo.fixture \
  --entry Main --construction-cases examples/caller-native-failure/construction-cases.json \
  --cases examples/caller-native-failure/evaluation-cases.json --attempts 6 --out /tmp/gooo-native-fault-demo
./.gooo body-construct --source /tmp/gooo-native-fault-demo/original.gooo \
  --construction /tmp/gooo-native-fault-demo/construction.json \
  --cases examples/caller-native-failure/evaluation-cases.json
```

출력 디렉터리는 새 경로여야 합니다. 생성한 프로그램 조합을 바로 두 번 실행하고,
실패도 같은 입력에서 반복되는지 확인합니다. 이 예제에서는 여섯 번째 후보까지
도달할 수 있습니다. 한도를 5로 줄이면 원래 한도 안의 부분 결과를 남깁니다.

## 기록하는 내용

- `rejection`: 타입·학습용 평가 단계에서 탈락한 후보. 호출 프로그램을 실행하지 않았습니다.
- `runtime.traces[].deliveries[].fault`: 실제 실행한 정수 나눗셈·나머지 연산의 0 제수.
  원래 활동, 실행 투영의 식 위치·해시, 두 피연산자를 남깁니다.
- `blocked_by`: 결과를 내지 못한 직접 생산자. 해당 결과가 필요한 단계는 호출하지 않습니다.
  앞서 끝난 활동과 독립 활동의 값은 그대로 보존합니다.
- `runtime.outcomes`: 사용자가 제공한 기대값을 일치·불일치·연산 실패·의존 결과 없음으로 셉니다.
  전체 실행 관측이 끝나지 않으면 이 집계는 제공하지 않습니다.

지역 학습 점수, 홀드아웃, 호출 결과는 서로 구분합니다. 연산 실패를 숫자 0이나
지역 평가 탈락으로 바꾸지 않습니다. 실패 이력이 있는 조립은 `joint-construction/v6`이며,
저장 재실행은 후보 순서·식·실패 입력·기대값·중단 이유를 확인하고 모델을 호출하지 않습니다.
기존 v1–v5 기록은 원래 의미를 재현합니다.

현재 관측하는 연산 실패는 `int64`의 `/`, `%`에서 제수가 0인 경우입니다. 실행 도구 오류,
취소·시간 초과·알 수 없는 프로세스 오류는 요청을 종료합니다. 정상 Gooo 소스와 저장된
순수 Go 투영은 그대로 두고, 실행 관측용 투영과 드라이버의 해시를 따로 기록합니다.

`mixed-model.gooo.fixture`는 같은 실패 후보와 기존 레코드 선택을 함께 조립합니다.
기존 자체 모델을 `--model`로 지정하면 첫 레코드 선택 순서에 사용하며, 빈칸 후보의
후속 순서는 소스 순서를 따릅니다. 최대 공간은 48개이므로 `--attempts 48`을 사용합니다.
선택에 쓰지 않은 최종 평가 파일은 `mixed-evaluation-cases.json`입니다.
