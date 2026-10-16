# 미등록 연산의 다음 단계를 Gooo가 설명하기

에이전트가 처음 보는 연산을 요청하면, 그 요청과 원래 소스의 기원을 남기고 다음에
무엇을 해야 할지 알려줘야 합니다. 현재 Gooo는 이 설명의 여섯 값을 한 레코드로
만드는 개발 소스를 갖고 있습니다.

```gooo
activity DescribeOperation(registered Boolean, originRecorded Boolean) -> OperationDisposition
    id "gooo://metric-operation-origin/describe-operation" computes `
return OperationDisposition{
    resolution: OperationResolution(registered, originRecorded),
    reason: OperationReason(registered, originRecorded),
    next_operation: OperationNextStep(registered, originRecorded),
    readiness_credit: false,
    repository_workspace_writes: false,
    promotion_authorized: false,
}
`
```

전체 필드 선언과 기대값은
[`policy.gooo`](../../internal/meta/metricprogram/unresolvedpolicy/policy.gooo)에 있습니다.
문자열 세 개와 불리언 세 개를 모두 원문이 정합니다. Go 연결기는 안정 필드 ID를
보고 네이티브 전송 구조체로 값을 옮깁니다.

## 현재 구분하는 입력

| 등록 여부 | 기원 기록의 필수 값이 있는가 | 해상도·이유 | 다음 단계 |
| --- | --- | --- | --- |
| 있음 | 있음 | `exact_operation` / 등록된 소스 연산 | 해당 연산 경로 계속 |
| 없음 | 있음 | `invariant_only` / 소스 등록에 없는 연산 | 소스 연산 선언 |
| 있음 또는 없음 | 없음 | `invariant_only` / 연산 기원 필요 | 기원 수집 |

기원 필수 값의 존재 여부는 현재 Go 경계에서 읽습니다. 그 값들이 진짜 요청자나
외부 사실을 인증한다는 뜻으로 사용하지 않습니다. 실제 미등록 연산의 거부 경로는
`registered=false`로 이 원문을 호출합니다. 결과는 요청한 작업의 설명 데이터입니다.
현재 원문은 준비도·쓰기·승격 값을 네 입력 조합에서 모두 거짓으로 지정합니다.

거부 레코드는 원래 요청, 등록 소스·Go 투영의 경로와 해시, 입력 바이트 해시,
설명 정책의 소스·투영 해시를 함께 운반합니다. 원래 연산이 없어도 무엇을 요청했고
어떤 소스를 보고 판단했는지 이어서 읽을 수 있습니다.

## 같은 사례의 프로필을 옮기기

기존 사례 `unresolved-operation-source-origin`은 같은 ID·경로를 유지합니다.
문자열·불리언 필드를 표현하는 명시적인 EntityFields V2 프로필로 옮겼습니다.
프로필 변경 명령이 만든 세 파일 후보를 검증한 뒤 개발 작업공간에 적용했습니다.
사례 수 131개와 기존 평가 분모 이력은 유지합니다.

이 단계의 V2는 등록된 소스의 왕복 검사 프로필입니다. 현재 본문 생성 입구는
기존 V3 본문 프로필로 이 필드 집합을 해석합니다. 두 입구의 실제 사용 범위를
각각의 관찰에 남깁니다.

[프로필 변경 명령](syntax-profile-migration.ko.md)은 변경 전후 원문을 보관하고,
네이티브 등록의 위치를 실제 `Cases` 표현식에서 찾습니다. 후보 생성과 검증은
저장소 적용 권한을 부여하지 않습니다. 이번 적용은 에이전트의 개발 변경이며,
완성된 소스의 원래 CI가 판단할 대상입니다.

## 원문에서 실행까지

1. 원문의 안정 활동 ID로 생성할 활동을 찾습니다.
2. `assembling`의 네 `value_case`가 두 불리언의 네 조합을 각각 한 번 담는지 확인합니다.
3. 고정 본문 하나를 검사하고, 네 입력·여섯 필드의 관찰을 남깁니다.
4. 타입이 확인된 Go 본문과 필드 ID 연결기를 생성합니다.
5. 파일별 원문·생성물 해시와 영역 ID를 manifest에 기록합니다.
6. 네이티브 거부 경로가 생성된 레코드의 여섯 값을 읽습니다.

이 원문에는 선택지가 없으므로 모델 순위를 요청하지 않습니다. 생성기 실행은 다음과
같으며, 마지막 인수는 해당 실행의 보고서를 저장할 때 지정합니다.

```sh
go run ./scripts/metric-unresolved-policy-codegen \
  internal/meta/metricprogram/unresolvedpolicy/policy.gooo \
  internal/meta/metricprogram/unresolvedpolicy /tmp/gooo-disposition-generation.json
```

생성 파일은 선언한 영역 단위로 최대 75줄씩 나뉩니다. 각 파일을 교체한 뒤 manifest를
마지막에 씁니다. 줄어든 파일 집합에서는 이전 manifest가 여전히 소유한 파일만
퇴역 대상으로 잡습니다. 이 절차는 파일 집합 전체의 전원 장애 트랜잭션을 보장하지
않으며, 중간 상태는 소스·파일 해시 검사가 수용하지 않습니다.

## 확인한 범위

[원래 관찰과 바이트 아카이브](../observations/source-disposition-record/observations.json)에
개발 중의 별도 단계 25개 새 사례를 보관합니다. 실제 정책의 생성된 Go 함수는 네
불리언 입력에서 24개 필드값이 맞았습니다. 두 별도 네이티브 변형은 다음을 확인했습니다.

- 활동·레코드·소스 필드·Go 필드의 표시 이름을 바꿔도 안정 ID로 값이 전달됩니다.
- Gooo의 세 불리언과 그 기대값을 바꾸면 실제 실행 값도 바뀝니다. Go 기본값으로
  채운 결과와 구별하기 위한 순수 데이터 변형이며, 작업 실행 권한을 주지 않습니다.

각 변형의 네 입력은 그 원문이 선언한 네 입력과 겹칩니다. 관찰한 영역의 필드
완전성을 설명하며, 새로운 입력에 대한 일반화나 전체 언어 완성을 뜻하지 않습니다.
실제 producer·consumer의 16개 기존 반례, 상위 미등록 요청, 기존 24개 의무와
23개 무관한 결과의 보존은 이 변경의 원래 CI에서 별도로 확인할 조건입니다.

arm64 개발 기계의 인라인 크기는 결과 구조체 56바이트, 전체 거부 구조체
304바이트였습니다. 문자열의 별도 저장 공간과 allocator·RSS를 제외한 크기입니다.
단일 소스 조립 관찰은 약 8.47ms였으며 전체 명령·네이티브 컴파일·p95/p99를 포함하지
않습니다. 이 수치로 전체 컴파일러의 지연 개선을 주장하지 않습니다.

다음 과제는 기원이 누락되거나 프로그램 일부만 만들 수 있는 경우에도, 원래
평가 대상을 유지하면서 가능한 부분과 막힌 부분을 함께 표현하는 것입니다.

확장 비용은 [최근 변경 분석](../research/extension-cost-20261016.ko.md)에 따로 기록합니다.
이 원문의 여섯 값을 옮기는 데 든 Go 구현과, 두 생성기가 공유하는 필드 ID 해석·파일
교체 비용을 구분합니다. 후속 연산에서 같은 연결 코드를 다시 작성하는지 확인합니다.
