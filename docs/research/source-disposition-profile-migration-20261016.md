# 같은 프로그램의 문법 프로필을 바꾸기

2026-10-16, PR #1621 소스 `9c1b231522b01a1dbcb414591c184967f6f7a745`에서
후속 개발을 조사한 기록입니다. 아래 레코드 적용과 프로필 변경 연산은 구현 전입니다.

## 구체적인 문제

현재 미등록 연산의 해상도·이유·다음 단계는 Gooo 활동 세 개가 정합니다. 반환 기록의
준비도·쓰기·승격 필드는 Go 구조체의 거짓 기본값으로 채웁니다. 다음에는 이 여섯
값을 Gooo 레코드 생성식에서 함께 읽을 수 있게 하려 합니다.

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

여기서 `OperationDisposition`은 안정 ID가 있는 문자열 필드 세 개와 불리언 필드
세 개를 가진 원본 레코드입니다. 본문 생성기는 이 표현과 고정된 세 보조 활동을
이미 처리합니다. Go 연결기는 필드의 안정 ID·타입을 대응시켜 전송합니다.

## 실제로 확인한 경계

새 검사 `TestSourceDispositionProfileFresh1751`의 서로 다른 사례 세 개가 통과했습니다.
기존에 끝난 검사나 실행 집합은 재실행하지 않았습니다.

| 진입점 | 새 레코드를 넣은 입력의 결과 |
| --- | --- |
| 일반 파서 | `parse.entity-fields-deferred` 유지 |
| 명시적인 EntityFields V3 프로필 | 구문 해석 성공 |
| 본문 생성기 | 레코드 1개·필드 6개·고정된 보조 활동 3개, 타입 검사와 결정론적 재생성 성공 |

이것은 해당 소스 입력과 세 진입점의 확인입니다. 등록된 문법 사례의 프로필 변경,
실제 프로그램 실행, 독립 소비자의 변경된 기록 수용을 확인한 결과는 아닙니다.

현재 `unresolved-operation-source-origin` 사례는 일반 프로필로 등록돼 있습니다.
레코드를 붙인 소스를 이 사례로 검사하면 기존의 프로필 거부가 발생합니다.
`RegisterSyntaxCapability`는 새 ID·경로의 사례를 추가하고 평가 집합을 한 개
늘리는 연산입니다. 이미 등록된 ID·경로를 주면 `REGISTRATION_CASE_ALREADY_EXISTS`로
거부합니다. 기존 사례의 프로필 변경을 새 사례 추가로 처리할 근거가 없습니다.

## 후속 연산의 요구사항

프로필 변경에는 별도의 명시적 요청과 원본에 연결된 후보가 필요합니다.

1. 변경 전 사례의 ID·경로·판단 계열·메타 연산을 정확히 지정합니다.
2. 변경 전·후의 프로필 ID를 지정하고, 지원하는 프로필의 실제 소스 해석을 수행합니다.
3. 사례 ID 집합과 평가 대상 수는 유지합니다. 무관한 사례와 과거 분모 기록도 보존합니다.
4. 실제 원본, 원본 스냅샷, 요청, 생성 후보를 각각 바이트 해시로 연결합니다.
5. 후보는 JSON 등록과 네이티브 등록에서 같은 사례의 프로필만 변경합니다.
   소스 단위의 위치는 구문 구조에서 찾으며 특정 파일명이나 행 번호로 고정하지 않습니다.
6. 원본의 충돌·누락·낡은 요청·모호한 사례는 서로 다른 미해결 원인으로 남깁니다.
7. 준비도·저장소 쓰기·승격 권한은 부여하지 않습니다. 후보 검증과 실제 적용을 구분합니다.
8. 원래 CI가 프로필 변경의 실제 생산·소비와 기존 원본 집합 보존을 확인해야 합니다.

기존 신규 등록 연산의 9개 산출물 계약은 유지합니다. 프로필 변경은 자체 산출물과
보존 법칙을 정의해야 합니다. 프로필 선택을 위해 일반 파서의 기본 지원 범위를
넓히는 변경도 이 작업에서 자동으로 도출되지 않습니다.

이 경로를 구현한 뒤 원래 정책 파일에 결과 레코드를 적용합니다. 레코드의 표시 이름을
바꿔도 필드 ID로 연결되고, 독립 소비자가 source-owned 여섯 값을 확인하는 것이
실제 사용 경로의 완료 조건입니다.
