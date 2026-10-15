# 문법으로 파생하는 경로의 참조와 완전성

`source_fill`의 `derive`는 Gooo가 정한 작은 표현 문법에서 후보를 만든다.
모델이 새 코드를 자유롭게 써서 빈자리에 넣는 방식과는 구성 규칙이 다르다.
소스는 문법, 남길 표현 수, 전체 조합 수와 확인할 입력을 선언한다.

```gooo
source_fill intent "조건과 값을 함께 구성한다" {
    hole "condition"
    hole "value"
    derive assignments max_candidates "4" {
        hole "condition" grammar "integer-predicate/v1" max_expressions "2"
        hole "value" grammar "integer-offset-constant/v1" max_expressions "2"
    }
}
```

## 아직 채우지 않은 경로에서도 이름 바꾸기

현재 지원하는 16개 버전의 표현 문법은 상수, 입력·필드 참조와 값 연산자를 조합한다.
새 활동 호출이나 지역 변수 선언을 만들지 않는다. 따라서 활동 참조는 원래 본문의 범위와
소스 위치에 남는다. 이름 변경은 이 문법의 성질을 확인한 뒤 본문에서 활동 ID를 따라간다.
후보 열거·점수 계산·모델 예측·프로그램 실행은 필요하지 않다.

예를 들어 `return fees.Base(__GOOO_BODY_HOLE_value__)`의 `Base`는 빈자리의 후보가
어떤 값 표현으로 채워져도 같은 활동이다. import를 각 파일의 범위에서 해석하고,
소유 패키지와 ID로 이 이름을 바꾼다. 직접 선언한 호출 후보는
[실제 빈자리의 범위에서 해석하는 별도 처리](workspace-source-fill-refactoring.ko.md)를 사용한다.
새 문법 버전에는 후보 참조의 범위 정보가 필요하며, 기존 버전의 성질을 자동으로 물려주지 않는다.

## 완전성의 분모를 읽기

새 `derive129` 예제는 이름 변경 후 전용 fill API로 본문을 조립했다.

| 항목 | 실제 관찰 |
| --- | --- |
| 조건 문법의 표현 | 8개 중 2개 유지, 25% |
| 값 문법의 표현 | 5개 중 2개 유지, 40% |
| 두 문법의 표현 합계 | 13개 중 4개 유지, 약 30.8% |
| 유지한 표현들의 전체 조합 | 4개 중 4개 구성, 100% |
| 선택용 입력 | 실제 일치 1/1 |
| 선택 뒤 별도 입력 | 실제 일치 1/1 |
| 갱신한 활동 참조 | 본문·미사용 소비자 2/2, 파일 3개 |

여기서 조합 100%의 분모는 유지한 표현들로 만들 수 있는 4개다.
원래 두 문법 전체, 적지 않은 요구와 모든 입력에서의 동작까지 뜻하지 않는다.
관찰 기록의 `grammar_complete`는 `false`이고, 문법 표현 범위와 조합 범위가 함께 남는다.
선택 사례와 별도 사례도 각각의 분모로 기록한다.

## 실제 명령과 API 사용

```sh
gooo package rename-activity --package derive129/helpers --id urn:derive129:base --to 기본값 --json workspace.json
```

[예제](../../examples/workspace-derived-fill-rename/README.md)는 원본 소스와 manifest를 유지하고
변경 제안만 반환한다. 별도 새 공유 문법 `derive148` 소스는 실제 CLI에서 참조 1/1개,
파일 2개를 갱신하는 제안을 확인했다. 후보를 열거하지 않았고 모델 예측·점수 계산·소스 프로그램
실행은 각각 0회다. CLI 프로세스 1회와 소스 프로그램 실행은 구분한다.

Go 호출자는 `SourceAssembly`로 소스 계약을 읽은 뒤, fill은 `GenerateWithSourceIRBodyFill`,
search는 `GenerateWithSourceIRSearch`로 구성한다. `TypedPathGenerator.GenerateSourceAssembly`는
스칼라·레코드 선택을 다룬다. 여기에 fill/search를 전달하면 이제 해당 전용 API를 안내한다.
미완성 빈자리를 일반 본문으로 투영하기 전에 멈추며, 모델을 다른 모드에 묵시적으로 사용하지 않는다.

앞선 새 검사는 잘못된 소비 API를 사용해 실패했고, 또 다른 검사는 search 필드 순서를
잘못 적어 파싱에서 실패했다. 원래 소스·실행 기록은 유지했다. 수정된 과거 로컬 소스를
다시 실행하지 않고, 각각 다른 새 소스로 조립과 API 안내를 확인했다.

## 관찰 범위

관찰은 공개 컴파일러 `ae957835`와 각 작업의 원래 source overlay에 연결한다.
classifier의 16개 버전 검사는 문법 성질 분류의 계약 검사이며 모든 문법의 생성·실행 관찰은 아니다.
구성 동작은 새 per-hole 스칼라 소스에서, 공유 문법 rename은 별도 새 CLI 소스에서 확인했다.
구성의 case/holdout은 유한 AST 평가이고 네이티브 프로세스와 모델 예측은 0회다.
전체 PROV 그래프, 외부 소비자, 에디터 적용과 이름 변경 후 네이티브 동작은 별도 범위다.
`search` 후보의 리팩토링 범위는 후속 과제로 남아 있다.
