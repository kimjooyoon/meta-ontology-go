# 파일별 import로 필드 타입을 읽기

프로그램을 여러 파일로 나누면 `Count` 같은 이름을 서로 다른 모듈에서 쓸 수 있다.
이름을 찾을 때 어느 파일의 import를 읽었는지 잃으면 정수와 불리언을 혼동하거나,
실행 가능한 소스를 만들면서 타입 이름이 충돌한다. Gooo에서는 소스에 적은 참조와
해석된 타입 ID를 함께 보관하는 방식으로 이 문제를 다룬다.

## 작은 프로그램

`counts.gooo`는 기본 정수와 문자열에 사용할 이름을 내보낸다.

```gooo
package counts
namespace field270_counts
entity Count id "urn:gooo:type:integer"
entity Label id "urn:gooo:type:string"
```

`ledger.gooo`는 이 파일에서 선언한 import를 이용해 필드를 구성한다.

```gooo
package app
namespace field270_app
import data "field270/counts"
entity Ledger id "urn:field270:ledger" fields {
 field amount id "urn:field270:amount" type data.Count required one
 field note id "urn:field270:note" type data.Label optional one
}
activity Pass(Ledger) -> Ledger id "urn:field270:pass" computes "return input"
```

다른 파일에서는 같은 `data` 별칭을 다른 패키지에 연결할 수 있다.
예제의 `claim.gooo`는 `field270/flags`를 가져오며, 그곳의 `Count`는
`urn:gooo:type:boolean`이다. 이 파일의 `data.Count`는 불리언으로 해석된다.
서로 다른 파일의 별칭을 합쳐 읽지 않는다.

같은 패키지 안에서 활동을 별도 파일에 선언해도 연결할 수 있다.

```gooo
activity Decide(Ledger) -> Claim id "urn:field270:decide" computes "return Claim{eligible: input.amount >= 31}"
bind Pass.result -> Decide.input
```

생산자의 결과 타입과 소비자의 입력 타입이 같아야 한다. 파일을 넘는 연결에는
원래 `bind`의 파일·범위와 두 활동의 ID가 함께 남는다.

## 소스에서 실행까지 남기는 정보

| 단계 | 보관하거나 사용하는 정보 |
| --- | --- |
| 원본 소스 | `data.Count`라는 표기와 파일별 import |
| 패키지 해석 | 소유 엔티티 ID, 필드 ID, 제공 패키지, 기본 타입 ID, 원본 범위 |
| 임시 lowering 뷰 | 기존 V4가 이해하는 정수·문자열·불리언 ID |
| 실행 소스 구성 | 같은 표시 이름을 가진 서로 다른 타입 ID의 별도 내부 이름 |
| 연결 근거 | `bind`의 원본 범위, 생산자 ID, 소비자 ID, 전달 엔티티 ID |

이 임시 뷰는 전달받은 원본 소스 바이트를 바꾸지 않는다. 실행 준비 시 원본 digest와
참조 범위를 확인한 뒤 해석 정보를 사용한다. 이 검사는 내부 Build 결과의 일관성을
다루며, 임의의 외부 image 전체를 인증하는 기능의 관찰과는 범위가 다르다.

## 현재 표현 범위

- 필드 타입에 한 단계 `별칭.멤버`를 사용할 수 있다.
- 대상은 기존 V4의 정수·문자열·불리언 기본 타입이다.
- 기존 `required one`, `optional one` 필드 규칙을 따른다.
- 같은 패키지의 파일 간 활동 연결을 패키지 export 목록에서 해석한다.
- 중첩 레코드와 `many`를 실행하는 범위는 추가 설계와 관찰이 필요하다.

존재하지 않는 별칭·멤버는 `PACKAGE_FIELD_TYPE_IMPORTED_INVALID`, 현재 필드
프로필 밖의 타입은 `PACKAGE_FIELD_TYPE_PROFILE_UNSUPPORTED` 진단으로 반환한다.
진단에는 파일과 필드가 나온다. 파싱 진단에도 원래 소스 위치를 포함한다.

## 새로 관찰한 결과

[예제와 관찰 파일](../../examples/package-qualified-field-types/README.md)을 함께 읽을 수 있다.
관찰의 producer는 `c34bae45` 소스와 19개 Go overlay다. 이후 공개 커밋이나 병합
커밋으로 이 producer를 바꿔 적지 않는다.

별도로 작성한 field352는 `959490f5` 소스와 20개 overlay에서 한글 별칭·파일명과
타입 이름을 확인했다. 이 두 producer의 관찰은 각각 원래 소스에 묶어 읽는다.

| 관찰 | 분모와 결과 |
| --- | --- |
| 필드 참조 해석 | 3개 중 3개: 정수·문자열·불리언 |
| 파일 간 연결 | 1개: 원본 범위와 생산자·소비자 ID 유지 |
| 같은 표시 이름의 구별 | `Count` 2개: 정수 ID와 불리언 ID의 내부 이름이 다름 |
| native 실행 | 입력 3개, 활동 출력 6개 중 6개 일치 |
| 입력 경계 | 문자열 없음, 빈 문자열, 한글 문자열, `9007199254740995` 정수 |
| 별도 CLI 해석 | 새 field286 소스 1회, 참조 3/3와 연결 1개 유지 |
| 한글 별칭·파일명 | 새 field352 소스: 바이트 범위 3/3, 연결 1개, 입력 2개의 출력 4/4 |
| CLI 코드 생성·실행 | 새 field376 소스: 참조 3/3, 연결 1개, 입력 2개의 출력 4/4, native 2회 결과 일치 |
| 모델 사용 | 각 관찰에서 호출 0회 |

native 관찰은 내부 `ExecuteWorkspace` 경로에서 생성한 Go를 실행하고 반복 결과를
비교했다. CLI 관찰은 `package resolve`로 계약과 초기화 순서를 읽었으며 활동 본문을
실행하지 않았다. CLI 한 번의 시간은 프로세스 시작을 포함해 585,400,583ns였다.
이 수치는 반복 성능 측정이나 모델 추론 속도의 통계가 아니다.

이전에 작성한 새 fixture의 파싱·빌드 실패와 파일 간 연결 실패도 producer별로
남겼다. 수정한 과거 fixture를 로컬에서 다시 실행하지 않았고, 마지막 결과는 별도로
작성한 field270·field286 소스에서 얻었다. 전체 기존 테스트와 호환성은 제출 소스의
CI 결과를 읽어 판단한다.

한글 관찰에서는 `자료.값`, `자료.문장`과 `계수.gooo`, `상태.gooo`, `원장.gooo`,
`판정.gooo`를 사용했다. 바이트 범위가 원래 참조 문자열을 정확히 찾는지 확인했고,
기준 43의 직전 값과 일치 값에서 생성된 Go 실행 결과를 읽었다. 이 관찰의 입력 수와
출력 수를 앞선 field270 관찰의 분모에 더하지 않는다.

별도 field376 CLI 관찰은 실제 `package execute --json --cases cases.json workspace.json`으로
Go 생성·빌드·실행을 진행했다. 프로세스 시작부터 결과까지 한 번의 시간은
1,439,670,917ns였고, 하위 프로세스의 CPU 시간은 사용자 0.166825초·시스템 0.163154초다.
빌드 프로세스의 peak RSS는 87,195,648바이트, 두 native 실행은 각각 4,947,968바이트와
4,915,200바이트였다. 각 프로세스의 관찰값이며 PC 전체 사용률이나 전체 메모리 합계로
읽지 않는다. 원래 `959490f5` + 20 overlay producer와 빌드·실행 hash를 기록했다.

## 다른 언어에서 가져온 관점

[OML](https://opencaesar.io/oml/)은 온톨로지 import의 로컬 prefix와 멤버의 IRI를
명세로 정리한다. Gooo에 적용한 판단은 소스의 이름과 해석된 ID, 참조의 출처를
각각 보관하는 것이다. 이번 변경에서 파일별 import와 기본 타입의 연결을 직접 다뤘다.

[Nemo](https://knowsys.github.io/nemo-doc/intro/tour/)의 관계·규칙,
[Scallop](https://www.scallop-lang.org/doc/probabilistic/provenance.html)의 태그 전파,
[egglog](https://egraphs-good.github.io/egglog-tutorial/01-basics.html)의 동치 표현 탐색은
별도 비교 과제다. 다음에는 같은 작은 업무 요구에서 표현한 요구 수, 추가 구현량,
변경의 영향 범위와 실제 입력 결과를 각각 기록한다. 현재 경쟁 언어를 같은 환경에서
실행한 속도·메모리 순위는 없다.
