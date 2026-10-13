# source_fill 후보의 참조를 함께 바꾸기

`source_fill`은 Gooo 소스가 본문의 빈자리와 허용할 표현을 정하는 방법이다.
모델은 이 후보의 순서를 제안할 수 있다. 이름을 바꾸는 도구는 후보 선택과 관계없이
소스에 적힌 모든 참조가 어떤 활동을 가리키는지 알아야 한다.

```gooo
activity Build(Integer) -> Integer computes "return fees.Boost(__GOOO_BODY_HOLE_base__)+__GOOO_BODY_HOLE_bias__" assembling {
    source_fill intent "기본값과 가산값을 구성한다" {
        hole "base"
        hole "bias"
        candidate "plain" { fill "base" "input" fill "bias" "0" }
        candidate "nested" { fill "base" "fees.Boost(input)" fill "bias" "fees.Boost(0)" }
    }
    case "5" -> "56"
}
```

위의 `Boost`는 본문에도 있고 후보 표현에도 있다. 표지판을 바꿀 때 완성된 길뿐 아니라
설계도에 적힌 다른 길의 표지판도 함께 바꾸는 셈이다.

## 어떤 위치에서 해석하는가

후보 표현만 따로 읽으면 `fees`가 import 별칭인지 지역 변수인지 알 수 없다.
이 구현은 직접 선언한 각 후보를 해당 빈자리에 넣은 전체 본문을 읽는다.
그때 찾은 활동 호출은 원래 본문 또는 fill 문자열의 바이트 위치에 연결한다.
본문에서 후보마다 반복해 보이는 같은 참조는 원래 위치 기준으로 한 번 센다.

```gooo
computes "if input<0 { let fees = input; _ = fees }; return __GOOO_BODY_HOLE_amount__"
```

이 본문의 뒤쪽 빈자리에서는 앞선 블록 안의 지역 변수 `fees`가 import 별칭을 가리지 않는다.
반대로 `let fees = input; return __GOOO_BODY_HOLE_amount__`의 빈자리에서
`fees.Boost(input)`을 사용하면 같은 범위의 지역 변수이므로 참조를 해석할 수 없다.
이때 오류를 반환하고 부분 변경 제안은 만들지 않는다.

원래 문자열의 escape, 한글 이름, 각 파일의 import 별칭을 따라 편집한다.
후보 이름, hole ID, intent와 평가 입력은 같은 내용으로 유지한다.
활동은 소유 패키지와 stable ID로 선택하고, 변경 전후에 해석된 참조의 ID를 비교한다.

## 사용과 관찰

```sh
gooo package rename-activity --package scope83/helpers --id urn:scope83:boost --to 가산 --json workspace.json
```

[새 예제](../../examples/workspace-source-fill-rename/README.md)는 3개 패키지의 4개 파일을 사용한다.
모든 입력 파일을 유지하면서 원래 본문, 두 후보의 표현과 호출되지 않는 소비자에서
대상 참조 6/6개를 갱신하는 제안을 확인했다. 본문 참조는 후보 3개를 읽어도 한 번 센다.
다른 패키지의 같은 이름 활동은 기존 소스 바이트를 유지했다.

별도 새 예제는 빈자리의 실제 범위에서 import 별칭을 가리는 지역 변수를 발견했다.
오류와 변경 제안 0개를 확인했다. 앞선 다른 예제는 manifest에 import가 빠져 시작 단계에서
실패했으며 그 기록도 보존했다. 수정한 과거 로컬 예제를 다시 실행하지 않았다.

관찰 시점의 컴파일러는 공개 소스 `ae957835`와 새 소스 overlay 6개다.
해석·소스 제안 과정의 모델 예측, 후보 점수 계산과 프로그램 실행은 각각 0회다.
이름 변경 뒤 각 후보의 실제 기능 일치 여부와 네이티브 실행은 별도 관찰 범위다.

또 다른 새 coupons 소스는 실제 CLI의 JSON 경로에서 본문·후보 참조 3/3개,
파일 2개를 갱신하는 제안을 확인했다. 원래 파일은 유지했다. CLI 프로세스 1회와
소스 프로그램 실행 0회를 구분해 기록한다. 전체 명령의 한 번 관찰은 약 0.67초였고
초기화와 파싱을 포함한다. CPU/RAM·비교 속도·추론 속도 측정은 수행하지 않았다.
임시 실행 파일은 해시를 남긴 뒤 삭제했다.

## 지원 범위

직접 선언한 2~16개 후보, 1~8개 hole과 hole마다 본문에 한 번 나타나는 기존 규칙을 사용한다.
현재 값·조건 문법의 자동 `derive`는 [버전별 참조 성질](workspace-derived-fill-refactoring.ko.md)을 따라
본문 참조를 갱신한다. 새 문법 버전, `search`, 동적 함수 값과 parenthesized 호출에는 각각의 후보 위치 정보가 필요하다.
이 작업은 분류할 수 있는 소스 참조와 export ID·타입 역할을 비교한다.
전체 PROV 그래프, 외부 API 소비자, 에디터 적용과 자동 디스크 저장은 별도 범위다.

## 비교에서 가져온 과제

[OML의 module/import 명세](https://opencaesar.io/oml/)는 namespace와 멤버 참조 범위를 명시한다.
여기서 Gooo로 이어 온 과제는 파일의 별칭을 활동 ID와 연결하고, 아직 선택하지 않은 표현도
같은 기준으로 추적하는 것이다. Gooo의 소스가 소유한 후보를 유지하면서 편집하는 데 적용했다.
관련 언어와의 공통 실행 속도나 정답률 비교는 아직 측정 자료가 없다.
