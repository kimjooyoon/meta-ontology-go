# 결정과 이유를 함께 조립하기

재고 요청을 판단할 때 `true` 하나만 맞추면 필요한 정보가 빠질 수 있다.
승인·거절 이유와 어떤 요청에 대한 결과인지도 같은 결과로 전달해야 한다.
이 예제는 그 요구를 Gooo의 서로 다른 `Request`·`Decision` entity ID로 표현한다.

원래 소스에는 수량과 재고를 비교하는 두 조건이 있다.
양수가 아닌 수량은 `invalid_quantity`, 재고보다 큰 수량은 `insufficient_stock`을 반환한다.
나머지 분기는 소스가 선언한 세 후보 중에서 조립한다.

```gooo
source_fill intent "Keep the reservation decision, declared reason and request reference together." {
 hole "result"
 candidate "allowed" { fill "result" "계약.Decision{approved: true, reason: \"capacity_available\", reference: input.reference}" }
 candidate "unexplained" { fill "result" "계약.Decision{approved: true, reason: \"unknown\", reference: input.reference}" }
 candidate "declined" { fill "result" "계약.Decision{approved: false, reason: \"capacity_available\", reference: input.reference}" }
}
```

선택 사례는 승인 여부, 이유와 요청 번호를 모두 요구한다. 세 후보는 모두 타입 검사를 통과했고,
`allowed`는 선택 사례 1/1, 나머지 두 후보는 각각 0/1이었다.
`unexplained`는 승인 여부만 맞춰도 이유가 다르므로 탈락했다.
타입에 맞는 결과를 만들 수 있는지와 요구를 얼마나 만족했는지를 따로 읽을 수 있다.

## 실제 관찰의 범위

| 확인 | 결과와 분모 |
| --- | --- |
| 후보 선택 | 후보 3개, 선택 사례 1개. `allowed` 1/1. |
| 별도 평가 | 선택 입력과 다른 사례 1/1. |
| 실제 호출 | 정상·재고와 같은 수량·0·재고 부족·음수 재고·음수 수량, 입력 6개. 두 활동의 출력 12/12. |
| 결정 레코드 | 위 출력 중 `Decision` 6개가 전체 일치. 그 안의 승인·이유·요청 번호 값 18/18. |
| 입력 분리 | 실제 호출의 서로 다른 입력 6개, 선택·별도 평가와 겹친 입력 0개. |
| 실행 반복 | 두 native 실행의 결과 일치. |
| 모델과 provider | 각각 0회. |

관찰의 생산자는 실제 `bbfcaebb39a62fd4c7bc1b0f6a0907b4d246d0e8`이며 overlay는 없다.
새 CLI 호출은 1회였다. 호출 wall은 1,527,537,750ns, 자식 CPU 합계는
user 0.179873초/system 0.186325초였다. 시작·구성·빌드·두 native 실행을 포함한다.
전체 PC의 사용률이나 다른 언어 대비 성능을 측정한 값은 아니다.

CLI는 종료 코드 0과 PASS를 반환했다. 뒤의 기록 도구가 소스 해시 필드 이름을 잘못 읽어
멈췄고, 원래 저장된 stdout에서 기록을 복구했다. 소스와 사례를 다시 실행하지 않았다.
입력 파일은 보존했고 임시 CLI는 해시를 남긴 뒤 제거했다.
[전체 입력과 관찰](../../examples/source-decision-reasons/README.md)을 함께 제공한다.

## 다음에 필요한 설명

현재 이유는 소스에서 직접 선언한 값이다. 실행 기록은 활동 ID, 필드 ID,
입력·출력과 후보 선택 결과를 연결한다. 개별 분기에서 어느 조건이 성립했는지 자동으로
나타내는 근거 경로는 후속 요구다. [Mangle의 provenance](mangle-and-source-decisions.ko.md)를
참고해 실제 평가값, 소스 조건과 결과 ID를 연결하는 작은 설명 단위부터 설계한다.

재고의 실시간 변경, 동시에 들어온 요청, 누락되거나 오래된 데이터는 별도 업무 요구다.
이번 사례의 결과와 분모를 이 요구들에 확대하지 않는다.
