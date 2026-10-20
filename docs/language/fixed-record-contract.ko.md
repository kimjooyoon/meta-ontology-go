# 계산과 기대 결과를 같은 소스에 적기

배송 여부를 판단한다고 가정해 봅시다. 재고가 있고 결제가 끝났으면 한 개를
배송합니다. 결제가 끝나지 않았다면 결제를 기다리고, 결제만 끝났으면 재고를
기다립니다. Gooo의 [고정 본문 예제](../../examples/fixed-record-contract/main.gooo)는
이 순서를 본문에 쓰고, 같은 파일의 `value_case`에 기대 결과를 적습니다.

이미 작성한 본문 하나를 평가합니다. 후보 수는 하나이고 모델 호출은 0회입니다.
레코드에는 배송 가능 여부, 이유, 수량을 담으며 각 필드에 안정 ID를 붙입니다.

## 이름으로 읽고, 순서로 연결하기

```gooo
activity Decide(stocked Boolean, paid Boolean) -> ShippingDecision
```

본문은 `stocked`, `paid`를 사용합니다. 외부 평가 JSON은 선언 순서에 따라 입력을
연결합니다. [입력 이름 안내](named-inputs.ko.md)에서 이름과 위치, 타입 ID의 관계를
설명합니다.

| 소스 입력 이름 | 평가 JSON의 키 | 의미 |
| --- | --- | --- |
| `stocked` | `Decide.input0` | 재고 여부 |
| `paid` | `Decide.input1` | 결제 여부 |

소스의 `value_case` 입력 배열도 같은 순서입니다. `[false,true]`는 재고가 없고
결제가 끝난 경우이며, 결과는 배송 불가·재고 대기·수량 0입니다.
두 불리언 입력의 네 조합을 모두 적었습니다.

## 생성 결과에서 읽을 것

저장소 루트에서 개발 버전 컴파일러의 본문 생성 명령을 실행합니다.
이 진입점은 예제의 레코드 본문 프로필을 읽습니다.

```sh
go run ./cmd/gooo body-codegen --json --activity Decide examples/fixed-record-contract/main.gooo
```

`report.record_assembly`는 선언된 사례와 필드값을 각각 셉니다.

| 항목 | 모든 기대값이 맞을 때 | 기대한 이유 하나를 바꿨을 때 |
| --- | --- | --- |
| `status` | `COMPLETE_FINITE` | `PARTIAL_FINITE` |
| `passed / total` | 4 / 4 | 3 / 4 |
| `fields_passed / fields_total` | 12 / 12 | 11 / 12 |

하나의 사례에는 필드 세 개가 있습니다. 이유만 어긋나도 그 사례 전체는 통과하지
못하며, 일치한 나머지 필드값은 필드 집계에 남습니다. 입력·출력의 타입 오류는
진단 단계에서 처리합니다.

현재 명령은 일부 기대값만 맞는 본문도 생성해 성공 종료할 수 있습니다.
사용하는 쪽에서 필요한 완전성은 `status`와 두 집계를 읽어 판단합니다.
이 수치의 평가 영역은 해당 소스에 적힌 네 입력과 세 필드입니다.

## 생성한 프로그램에서도 대조하기

Go 1.27.3를 준비하고 결과를 저장할 새 경로를 지정합니다.

```sh
go run ./cmd/gooo body-compose --source examples/fixed-record-contract/main.gooo --entry Decide --cases examples/fixed-record-contract/evaluation.json --out /tmp/gooo-fixed-record-contract
```

이 명령은 생성된 프로그램을 컴파일하고 두 프로세스에서 실행합니다.
`runtime.finite_passed / finite_total`은 실제 출력과 기대값의 대조 결과이고,
`runtime.runs`에는 두 실행의 종료 상태와 시간이 남습니다.
`runtime_replayed`는 두 실행의 출력이 일치했는지 나타냅니다.

현재 평가 파일은 소스와 같은 네 입력을 사용합니다. 따라서 실행에서 확인한 입력과
소스 사례가 겹칩니다. 두 불리언의 가능한 조합이 모두 소스에 있어, 이 입력 타입
범위에서 소스 사례와 분리된 입력은 0개입니다. 다른 도메인으로 넓힐 때에는 입력과
기대값의 범위를 다시 정하고 `input_separation`의 구분 결과도 함께 읽어야 합니다.

결과 폴더에는 받은 원본, 선택된 소스, 생성 Go 코드, 실행 드라이버와 실행 기록이
있습니다. `original_source_sha256`과 `selected_source_sha256`은 각각의 바이트에
연결됩니다. 정규화된 선택 기록을 붙이면 두 소스의 해시가 달라질 수 있습니다.
필드의 안정 ID는 표시 이름과 함께 출처를 이어 읽는 기준으로 남습니다.

[관찰 기록](../observations/fixed-record-contract/observations.json)은 로컬 생성 검사,
원래 CI의 실행, 실패한 검사 조건과 후속 수정 상태를 각 출처에 연결합니다.
