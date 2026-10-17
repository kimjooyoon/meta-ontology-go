# 두 메모리 기록의 측정 범위 설명하기

개발 소스용 예제다. 공개 릴리스의 지원 범위는 해당 버전의 안내를 따른다.
원본 CI에서 이 소스의 실행과 저장 기록 재실행을 별도로 확인하도록 구성했다.

두 숫자를 비교하기 전에 같은 것을 측정했는지 읽는 작은 Gooo 프로그램이다.
예를 들어 빌드·테스트 자식 프로세스의 메모리 최댓값과 설명 도구 자신의 메모리
최댓값에는 서로 다른 이름표가 필요하다. 계산 규칙은 `main.gooo`의 조건문에 있다.

해당 개발 소스로 준비한 `gooo`를 사용해, 저장소의 `examples/measurement-metadata`
디렉터리에서 실행한다.

```sh
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
```

첫 응답의 `decision`이 `PASS`이고 `result.runtime`의 `finite_passed`와 `finite_total`이
모두 14인지 확인한 뒤 저장한 구성으로 다시 실행한다.

```sh
gooo package replay --json --receipt execution.json --cases cases.json gooo.workspace.json
```

`package replay`는 저장한 구성과 새 입력으로 프로그램을 실행한다.
기본 경로는 모델 없이 결정론적으로 동작한다.
이 예제는 사용자가 별도로 요청하는 설명이다. 컴파일러나 고객 CI의 실행 조건에
연결하지 않는다. 설명을 만들지 못한 경우 원래 작업의 결과와 출처를 그대로 둔다.

## 입력과 출력

두 `MeasurementMetadata` 레코드를 `input0`, `input1`에 전달한다. 단위는 RSS KiB로
고정하며 숫자 자체는 입력받지 않는다. 시간·CPU·서로 다른 단위의 변환은 이 예제의
입력이 아니다. 모르는 문자열은 빈 값으로 보존한다.

| 입력 필드 | 담을 내용 |
| --- | --- |
| `subject` | 실제로 측정한 프로세스 또는 프로세스 집합 |
| `interval` | 그 관측의 시간 범위. 근처에 적힌 실행 타이머에서 추정하지 않는다. |
| `statistic` | 하나의 프로세스 최고치, 여러 자식 최고치 중 최댓값 등 집계 방식 |
| `measurement_method` | 실제 수집 방식과 의미 |
| `collector_revision` | 측정 수집기의 고정된 기원. 관측 대상 프로그램의 버전과 구별한다. |
| `origin_ref` | 원본 산출물·레코드·필드를 가리키는 위치 |
| `value_status` | 명시적으로 보존한 `OBSERVED`, `UNKNOWN`, `SENTINEL` 상태 |

과거 기록에 숫자 `1`만 있으면 실제 관측인지 대체 숫자인지 복원할 수 없다.
그때는 상태를 `UNKNOWN`으로 둔다. 수집기 버전을 모를 때 컴파일러 SHA로 대신
채우지 않는다. 원본 필드와 생산자 소스에서 해석한 설명도 구별해 보관해야 한다.

| 출력 필드 | 의미 |
| --- | --- |
| `domain_relation` | 알려진 측정 범위의 차이, 선언한 네 필드의 일치, 또는 정보 부족 |
| `origin_status` | 관측 상태와 원본 위치·수집기 정보의 존재 여부 |
| `reason` | 먼저 확인할 한 가지 이유. 위 두 필드의 독립적인 결과도 함께 읽는다. |
| `claim_limit` | 항상 `NO_IMPROVEMENT_CLAIM` |

네 범위 필드 중 알려진 차이가 하나라도 있으면 다른 정보가 빠져 있어도 `DIFFERENT`를
보존한다. 모두 채워져 같으면 `MATCHED_DECLARED_METADATA`를 반환한다. 원본 위치가
서로 달라도 각각 존재할 수 있다. 프로그램은 그 위치를 가져오거나 진위를 확인하지
않으며, 문자열의 일치만으로 실행 환경이나 수집기 변경의 동등성을 증명하지 않는다.

상위 실행 보고서의 `PASS`는 사례의 기대값과 실행 결과가 일치했다는 뜻이다.
개별 설명이 `DIFFERENT`나 `UNKNOWN`이어도 그것이 기대한 결과이면 검사는 통과한다.
에이전트는 `result.runtime.traces[].deliveries[].actual` 안의 네 필드를 읽어 설명을
사용한다. 검사 성공을 메모리 수치의 비교 가능성이나 개선 성공으로 바꾸지 않는다.

## 보존된 실제 기록을 읽는 예

[2026-10-17 원본 실행](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/39022807095)의
`public-self-improvement-test-reuse-04ffab2c3f56372f662d14b60667acd841a5c332` 산출물
(`12096564398`)에는 다음 두 값이 있다. 실행을 다시 하지 않고 저장된 JSON을 읽었다.

| 원본 파일의 `peak_rss_kib` | 값 | 생산자 소스에서 확인한 수집 경로 |
| --- | ---: | --- |
| `test-reuse-baseline-report.json` | 90,392 | `/usr/bin/time -f %M`으로 읽은 build·test 값 중 최댓값 |
| `test-reuse-replay-report.json` | 16,180 | 판정 도구의 `/proc/self/status`에서 읽은 `VmHWM` |

[해당 버전의 생산자](https://github.com/kimjooyoon/meta-ontology-go/blob/04ffab2c3f56372f662d14b60667acd841a5c332/scripts/self-improvement-public-test-reuse/run.go)의
`baselineReport`, `runTimed`, `runReuse`, `currentPeakRSSKib`을 함께 읽어 대상과 수집 방식을
해석했다. 숫자는 원본 필드이고 이 설명은 생산자 소스에서 해석한 정보다. 각 `origin_ref`에
실행·산출물·파일·필드 위치를 남기고, 설명을 얻은 생산자 파일과 버전도 함께 보관한다.
보고서의 `wall_ms`로 RSS 측정 구간을 채우거나 컴파일러 SHA로 수집기 버전을 대신하지 않는다.
모르는 입력은 빈 값으로 유지한다. 이 쌍에서는 측정 대상의 차이를 먼저 설명하고 절감률
판단은 보류한다. 이 절은 실제 기록의 해석 예이며 자동 보고서 변환기는 아직 구현하지 않았다.

## 독립 기대값과 사용 범위

`cases.json`의 14행은 명시적으로 만든 가상 기록이다. 범위 일치, 대상·구간·집계·방법의
차이, 빠진 범위·수집기·원본 위치, 지원하지 않는 상태, 대체 숫자, 수집기 변경,
모든 정보의 누락, 차이와 대체 상태의 동시 존재, 관측 상태의 소실을 포함한다.
각 행의 네 출력 기대값을 별도로 적었다. `finite_total`은 `expected`의 출력 항목을
세므로 여기서는 레코드 비교 14개다. 선언한 필드값 56개와 입력 행 14개는 별도로
구분한다. 레코드의 어느 필드가 달라도 그 레코드 기대값은 불일치가 된다. Go 확인 코드는 기대값과 실제 결과·출처를
읽으며 분류 규칙을 다시 구현하지 않는다. 원본 CI 결과는 해당 소스 변경안에서 확인한다.

이 자료로 메모리 절감률·통계적 유의성·검사 재사용 권한을 만들지 않는다.
실제 보고서에 연결할 때는 각 입력이 어느 원본 필드나 생산자 소스에서 왔는지 먼저
연결해야 한다. 숫자의 출처와 측정 대상을 읽을 수 있게 하는 것이 이 예제의 역할이다.
