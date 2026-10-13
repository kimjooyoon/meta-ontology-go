# 가져온 업무 타입으로 값을 만들기

가격과 수량 패키지가 모두 `Value`라는 이름을 사용할 수 있다.
사용하는 파일의 import가 어느 패키지의 타입인지 정한다.
활동의 입력·결과 서명 다음으로, 실제 값을 만드는 본문에도 같은 해석이 필요하다.

```gooo
import 대상 "constructor716/money"
import 수량 "constructor716/count"
activity Quote(대상.Value, 수량.Value) -> 대상.Value id "urn:constructor716:quote" computes "return 대상.Value{value: input0.value * input1.value, note: input0.note}"
```

앞선 구현에서는 서명의 `대상.Value`를 해석해도 본문 생성자의 별칭이 실행 그래프에 남았다.
새 수정안은 원래 파일의 import와 제공 패키지의 entity ID를 따라 생성자 타입을 해석한다.
같은 별칭을 다른 파일에서 다른 패키지에 사용해도 각각의 범위를 읽는다.
원본 소스는 유지하고 임시 실행 소스에서만 내부 타입 이름을 사용한다.

## 후보 안에서도 같은 타입을 사용한다

값 생성자는 `computes`, 조립 baseline, 직접 선언한 `source_fill` 후보와 외부 fill plan에서 처리한다.
실제로 사용할 빈자리의 지역 변수 범위도 읽는다. 후보 표현만 따로 읽으면 자유로운 import 별칭처럼
보이던 이름이 그 빈자리에서는 지역 값일 수 있기 때문이다.
다른 블록에서 선언한 지역 값은 해당 빈자리의 import를 가리지 않는다.

다음은 [전체 CLI 예제](../../examples/package-qualified-record-constructors/README.md)의 후보 선언이다.

```gooo
source_fill intent "Use the source-declared product of price and adjusted quantity." {
 hole "result"
 candidate "product" { fill "result" "대상.Value{value: input0.value * input1.value, note: input0.note}" }
 candidate "sum" { fill "result" "대상.Value{value: input0.value + input1.value, note: input0.note}" }
}
```

소스가 곱셈과 덧셈 경로를 선언하고, 선택용 `value_case`가 필요한 결과를 정한다.
모델을 연결하지 않은 이번 예제는 결정론적으로 후보를 평가해 곱셈을 선택했다.
`holdout_value_case`와 실제 호출 입력은 각자의 결과로 기록한다.
선택할 부품 목록과 조립 위치가 도면에 있고, 계산 결과는 조립 후 확인하는 방식이다.

## 새 관찰

| 관찰 | 확인한 범위 |
| --- | --- |
| 원래 constructor612 | 서명은 해석됐지만 본문 생성자에 `price.Value`가 남음. 실행 전에 테스트 실패. |
| 새 constructor641 | 두 파일의 같은 별칭·서로 다른 업무 ID를 유지. 생성자 위치 3개, 입력 2개의 출력 8/8, 잘못된 범위 3개 거절, 부분 프로그램 수락 0. 문자열·주석·원본 보존. |
| 별도 constructor687 | 소스 후보 2개 중 선택 사례 1/1, 별도 평가 1/1, 다른 실제 입력 2개의 출력 2/2. 소스·외부 후보의 빈자리에서 별칭을 가리는 지역 값 2개 거절. 다른 블록의 지역 값 보존. 원본·외부 계획 보존. |
| 새 CLI constructor716 | 후보 조립과 가져온 레코드 생성자를 함께 처리. 선택 1/1, 별도 평가 1/1, 실제 입력 2개의 활동 출력 8/8, 내부 두 실행의 결과 일치. |

각 관찰의 모델 예측·외부 provider 호출은 0회다.
constructor641은 실제 `45439469`와 Go overlay 7개, constructor687과 CLI716은 같은 base와
overlay 8개를 생산자로 기록했다. 향후 제출·병합 커밋으로 이 생산자를 바꾸지 않는다.
원래 실패한 소스와 기존 관찰을 로컬에서 다시 실행하지 않았다.

CLI 한 번의 시작·준비·구성·빌드·실행 시간은 1,500,878,000ns였다.
해당 호출의 자식 프로세스 CPU 합계는 user 0.176860초, system 0.171698초다.
전체 PC 사용률이나 다른 언어와의 속도 비교는 별도 측정 대상이다.
25,709,506바이트 임시 CLI는 해시를 기록한 뒤 제거했다.

## 현재 표현 범위

생성자 타입은 한 import 별칭과 한 entity 멤버로 지정한다.
기존 프로필에서 각 필드에 정수·문자열·불리언 값 하나를 담는 레코드를 사용한다.
필드는 여러 개일 수 있으며 중첩 레코드와 다중 값 실행은 후속 과제다.
알 수 없는 별칭·멤버와 빈자리에서 지역 값에 가려진 별칭은 결과 생성 전에 거절한다.
이름이 같은 업무 타입의 ID 구별과 실제 숫자의 단위·출처 선언은 각각의 요구로 다룬다.

관련 기록: [업무 타입 ID](domain-record-identity.ko.md),
[필드의 한정 타입](package-qualified-field-types.ko.md),
[다른 언어와 비교할 과제](../ontology-language-comparison.ko.md).
