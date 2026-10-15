# Gooo가 만들고 있는 언어와 작은 모델

2026-10-15 개발 소스 기준. 공개 바이너리의 지원 범위는 각
[릴리스 안내](releases/0.6.34-dev.md)에서 확인할 수 있습니다.

Gooo는 계산의 의미, 허용한 구현 선택, 그 결과의 출처를 함께 다루는 언어를
만들고 있습니다. 코드를 고친 사람이 “이 값은 어디서 왔고, 왜 이 구현을
골랐나?”를 소스에서 실행까지 따라갈 수 있게 하는 것이 중심 과제입니다.

## 어떤 작업을 쉽게 만들고 싶은가

환불 규칙을 예로 들면 금액 계산, 승인 조건, 설명 보존을 각각 작성하고
그 조건을 만족하는 구현을 구성합니다. 결과가 바뀌면 어떤 입력·규칙·선택이
영향을 주었는지 살펴봅니다. 작업 지시서와 부품, 조립 기록을 함께 두는 모습에
가깝습니다.

일반 계산에는 `computes`를 씁니다. 구현을 구성할 때는 `assembling`에 후보와
검사할 사례·조건·예산을 선언합니다. 현재 지원하는 타입과 조립 프로필 안에서
작동하며, 미지원 표현이나 선택지로 해결하지 못한 요구도 결과에 남깁니다.

| 사용자가 표현하는 것 | 현재 입구 |
| --- | --- |
| 변수, 조건, 계산과 호출 | [첫 프로그램](getting-started.md), [일반 계산](../examples/everyday-program/README.md) |
| 후보가 만족해야 할 관계 | [소스의 판단 조건](../examples/source-behavior-criteria/README.md) |
| 확인할 입력과 별도 평가 | [추가 관측](../examples/source-record-observation/README.md), [별도 평가](../examples/record-field-holdouts/README.md) |
| 반환값까지 이어지는 소스 | [읽기 쉬운 기원 조회](language/readable-source-origins.ko.md) |
| 실행 기록을 해석하는 메타프로그램 | [기원 읽기 프로젝트](language/project-starters.md#start-an-origin-reader) |
| 선언·호출·값 버전을 구분하는 메타프로그램 | [지역 값 버전과 생성 관계](../examples/local-version-origins/README.md) |
| 값 버전의 생성·읽기, 짧은 설명과 RDF 교환 | [기록된 사용 관계](../examples/recorded-use-origins/README.md) |
| 받은 바이트와 해석한 문자열의 기원 | [Text 입력 규칙](../examples/text-input-origins/README.md) |

같은 정수 타입이라도 역할을 이름과 ID로 구분할 수 있습니다. 다음 계산은
요청의 금액에 조정액을 더하고, 지정한 경우 두 배로 만듭니다.

```gooo
package adjustment
namespace adjustment

entity Integer id "urn:gooo:type:integer"
entity Request id "urn:adjustment:request" fields {
    field amount id "urn:adjustment:amount" type integer required one
    field adjustment id "urn:adjustment:delta" type integer required one
    field doubled id "urn:adjustment:doubled" type boolean required one
}

activity Apply(Request) -> Integer id "urn:adjustment:apply" computes `
let amount = input.amount
amount += input.adjustment
if input.doubled {
    amount *= 2
}
return amount
`
```

`amount`와 `adjustment`는 같은 타입을 쓰면서 서로 다른 필드 기원을 갖습니다.
이처럼 지금 있는 표현으로 읽기 쉬운 프로그램을 만들고, 반복되는 불편을
확인한 뒤 문법을 넓힙니다.
[이 예제의 소스 조회](observations/readable-origins/input-role-source-view.txt)에서 세
필드와 갱신·조건의 연결을 읽을 수 있습니다. [관찰 기록](observations/readable-origins/input-role-source-observation.json)은
소스 분석의 범위와 시간을 남깁니다.

## 한 번의 조립을 따라가 보면

1. 작성자가 `.gooo`에 개념 ID, 함수, 허용할 후보와 판단 조건을 적습니다.
2. 컴파일러가 타입과 연결을 확인하고 허용된 후보를 구성합니다.
3. 고정 순서 또는 작은 결정 모델의 순위로 후보를 시도합니다.
4. 선택에 사용한 사례와 선택 뒤의 별도 평가를 각각 기록합니다.
5. 선택한 프로그램을 저장해 새 입력에 사용하고, 기록을 Gooo 프로그램으로 읽습니다.

소스가 허용한 선택과 검사가 작업의 범위를 정합니다. 자체 모델은 이 범위에서
먼저 시도할 경로를 고릅니다. 모델이 없으면 결정론적인 순서를 사용합니다.
현재 선택 후의 완성과 모델의 첫 선택 성공률은 각자 분모를 가진 관찰입니다.

## 언어·모델·실행기가 맡는 일

| 부분 | 맡는 일 |
| --- | --- |
| Gooo 소스 | 개념과 안정 ID, 계산, 선택지, 판단 조건을 소유합니다. |
| 컴파일러와 IR | 작성한 의미를 보존하면서 타입을 확인하고 실행 표현으로 옮깁니다. |
| 자체 소형 모델 | 지원하는 유한한 코드 경로의 시도 순서를 제안합니다. |
| Go 실행기 | 구성한 프로그램을 실행하고 관찰 범위와 출처를 전달합니다. |
| Gooo 메타프로그램 | 전달받은 기록을 자신의 조건으로 분류하고 다음 작업에 사용합니다. |

모델 개발은 개인 PC에서 다룰 수 있는 작은 결정 구조를 중심으로 이어갑니다.
학습 결과는 사용한 소스·분할·입력 범위와 함께
[모델별 안내](source-relative-v3-contract-model.md)에서 설명합니다.

## 차별점으로 발전시키려는 부분

Gooo의 가치는 개념의 출처가 구현과 실행을 거쳐서도 읽힌다는 데 있습니다.
표시 이름과 의미 ID를 구별하고, 원래 선언과 선택된 본문, 실행 중의 값 버전을
연결합니다. 같은 숫자가 나왔다는 이유로 같은 기원으로 판단하지 않습니다.

[PROV-O](https://www.w3.org/TR/prov-o/)의 Entity·Activity·Agent와 생성·파생 관계를
공통 어휘로 참고합니다. 현재 소스 의존성 조회와 실제 실행 기록은 각각의 범위를
가집니다. [출처 안내](language/source-origins.ko.md)에서 연결되는 좌표와 남은 틈을
확인할 수 있습니다.

최근 개발 소스에는 정수 지역 변수의 `+=`, `-=`, `*=`, `/=`와 텍스트 기원 조회를
추가했습니다. 표시 이름을 정하는 정책도 Gooo로 작성했습니다. 이 기능의 병합과
배포 상태는 [개발 릴리스 안내](releases/0.6.35-dev.md)와 해당 PR에서 확인합니다.

이어진 [조건 분기 판독기](../examples/source-control-reader/README.md)는
소스에 보관된 상수 조건과 실행 입력에 달린 조건을 Gooo의 조건식으로 읽습니다.
출처 식별자는 그대로 전달하고, 참·거짓·알 수 없음과 집계 범위를 Gooo가
구성합니다. [기록된 읽기의 설명](../examples/recorded-use-origins/README.md)과
함께, 언어가 자신의 구성 과정과 값의 기원을 읽는 작은 프로그램을 늘려갑니다.

[줄바꿈 조건문](../examples/else-layout/README.md)에서는 컴파일러가 허용할 표현도
Gooo 규칙으로 작성했습니다. 이름과 의미 ID가 있는 레코드로 토큰 관찰값을
구분하고, Go 호스트와의 연결도 Gooo에서 조립합니다. 표현을 더 편하게 쓰면서
원래 조건의 위치를 유지하는 작은 범위부터 발전시키고 있습니다.

[읽기 사건의 RDF 뷰](../examples/recorded-use-origins/README.md#읽기-사건을-rdf로-내보내기)는
Gooo가 기록을 확인하고 호출·읽기·값 버전의 관계를 직접 구성하는 예제입니다.
그 결과를 표준 RDF 도구로 읽을 수 있게 하여, 내부 기록을 외부에서 활용하는
입구를 만듭니다. 원래 문서의 해시와 좌표가 각 이름에 남고, Gooo 조건식과
출력 문장을 함께 읽으며 의미를 확인할 수 있습니다.

## 다음 개발의 우선순위

1. **읽을 수 있는 입력과 계산.** 현재 여러 인자는 `input0`, `input1`로 받습니다.
   본문에서 `let adjustment = input1`처럼 이름을 붙이거나 안정 ID가 있는 레코드
   필드를 사용할 수 있습니다. 이후 입력 이름과 포트의 의미를 함께 보존하는
   표현을 설계합니다. 같은 타입의 두 인자와 이름 변경을 구분해야 합니다.
2. **출처를 읽고 다시 구성하는 메타프로그래밍.** 소스의 가능한 연결과 실제
   실행에서 관찰한 갱신을 이어 읽습니다. 저장한 기록에 없는 연결과 기록 한도에
   걸린 부분은 미확정으로 유지합니다.
3. **작은 작업을 끝까지 쓰는 경험.** 설치, 작성, 실행, 결과 설명, 수정 후 재사용을
   한 예제로 이어갑니다. 실제 사용에서 드러난 표현의 반복을 언어 기능으로 줄입니다.

입력 역할을 표현하는 후속 설계에는 PROV-O의
[qualifiedUsage와 hadRole](https://www.w3.org/TR/prov-o/#qualifiedUsage)을 참고할 수
있습니다. 이 어휘의 커널 지원은 후속 설계 범위입니다. 먼저 작성자가 구분한
입력 역할과 실행 회차의 의미를 정리할 계획입니다.

## 발전을 확인하는 기준

- 처음 보는 개발자가 한 프로그램을 설치·작성·실행할 수 있는가.
- 소스에 적은 조건 중 무엇을 만족했고 무엇을 아직 확인하지 못했는가.
- 원래 선언에서 선택과 실행까지 출처를 이어 읽을 수 있는가.
- 모델을 교체하거나 생략해도 선택지와 판단 기준을 유지하는가.
- 변경된 의미의 소비자를 찾고 필요한 검사를 설명할 수 있는가.

CI 시간은 [검증 작업 도메인](ci-work-domain.ko.md)에서 따로 관찰합니다.
3분은 개선 목표이고 요청별 실행 예산은 명시합니다. 현재 검증 의무를 유지하면서
소비자 관계와 실제 비용을 다음 계획에 연결하는 단계입니다.

## 상세 기록과 연구

[앞선 개발 기록](language-direction-history-2026-10-15.ko.md)에 기존 안내의 내용과
모델별 수치, 예제, 실패, 참고 연구를 보존했습니다. 당시의 “현재”와 “다음”은
각 기록의 시점에서 읽어야 합니다. 후속 개발의 구체적인 표현 문제는
[언어의 정체성과 우선 과제](language-identity-problems.ko.md)에 모았습니다.
