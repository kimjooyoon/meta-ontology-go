# 이미 정한 계산에 기대 결과를 붙이기

재고와 결제 여부로 배송 결과를 만드는 작은 프로그램입니다. 판단 순서는
`Decide` 본문에 있고, 기대 결과는 같은 Gooo 파일의 `value_case`에 있습니다.
선택지 없이 작성한 본문 하나를 검사하고 Go 코드로 옮깁니다.

| 재고 | 결제 | 결과 |
| --- | --- | --- |
| 있음 | 완료 | 배송 가능, 1개 |
| 없음 | 완료 | 재고 대기, 0개 |
| 있음 | 미완료 | 결제 대기, 0개 |
| 없음 | 미완료 | 결제 대기, 0개 |

개발 버전 컴파일러의 코드 생성 명령입니다.

```sh
go run ./cmd/gooo body-codegen --json --activity Decide examples/fixed-record-contract/main.gooo
```

`report.record_assembly`에서 사례 4개와 필드값 12개의 결과를 각각 읽을 수 있습니다.
두 불리언 입력의 네 조합을 모두 적었으며, 이 입력 영역이 이 예제의 검사 범위입니다.
기대값이 일부 맞으면 `PARTIAL_FINITE`와 맞은 사례·필드 수를 남깁니다.
명령 경계의 새 검사에서는 기대한 이유 한 개를 바꿨을 때 사례 3/4와 필드값 11/12가
맞는 것으로 기록됐습니다. 받은 원본과 생성 단계의 정규화된 소스는 각각의 해시로 남습니다.

생성된 프로그램에서 같은 네 조합을 대조하는 명령도 준비했습니다. 결과 폴더는
새 경로로 지정합니다.

```sh
go run ./cmd/gooo body-compose --source examples/fixed-record-contract/main.gooo --entry Decide --cases examples/fixed-record-contract/evaluation.json --out /tmp/gooo-fixed-record-contract
```

본문의 단일 후보와 실제 프로그램의 네 입력 대조는 각각 기록합니다. 생성 프로그램은
두 프로세스에서 실행합니다. 원본의 해시와 필드 안정 ID는 생성 기록에 남습니다.
[관찰 기록](../../docs/observations/fixed-record-contract/observations.json)에서 실행한 소스와
검사 결과를 구분해 읽을 수 있습니다.

[결과 읽기 안내](../../docs/language/fixed-record-contract.ko.md)는 입력 이름과 평가 JSON의
연결, 부분 일치, 명령 종료 코드와 완전성 집계, 원본·선택 소스의 해시를 설명합니다.
