# 첫 후보의 원본 CI

후보 `4be83252b631cd92e5c74ea12faf88535279e4b8`, PR1433 → dev `035e4447`.
CI37990072227/attempt1/suite102937583222가 정상 종료했습니다. 전체 결론은 FAILURE입니다.
의미 일관성 job114021669927은 도움말 선언의 `NO_SAFE_DECLARATION_CAPACITY`로 실패했습니다.
원본 로그의 CLI 처리 함수에는 `CALLEE_EFFECTS_UNPROVEN`도 남습니다.
네 플랫폼 준비 검사37990072175/attempt1/suite102937583111은 SUCCESS였습니다.
CI evidence job114029619353은 이 실패한 의미 검사를 canonical 성공 증거로 받아들이지 않아
`canonical CI job "Semantic conformance" is missing or mismatched`로 종료했습니다.
관측한 main 보호 설정은 여섯 자동 검사, 필수 승인0명이며 Guardian 체크는 없습니다.

원본 실행을 취소·재시작·재실행하지 않았습니다. 최초 PR 본문도 그대로 보존합니다.
도움말·CLI 처리를 짧게 나눈 수정본과 실제 로드 시간을 기록하는 수정은 별도 생산자3def의
관측으로 남깁니다. 후속 CI는 새 소스의 별도 원본 실행으로 확인합니다.
이 실패를 성공으로 재표기하거나 통과 기준을 완화하지 않습니다.

각 원본은 압축 해제 후 바이트를 비교했습니다. `FILES.sha256`은 공개 바이트를 묶습니다.
