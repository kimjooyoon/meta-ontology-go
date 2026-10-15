# 워크스페이스의 활동 이름 바꾸기

활동의 이름은 사람들이 읽는 표지이고, ID는 같은 부품을 계속 가리키는 번호다.
여러 파일이 그 부품을 쓰고 있다면 표지를 바꾸면서 연결도 함께 갱신해야 한다.
Gooo의 워크스페이스 rename은 패키지와 활동 ID로 대상을 고른 뒤,
워크스페이스에 선언한 모든 소스 파일에서 해석할 수 있는 참조를 갱신한다.

```sh
gooo package rename-activity --package review/rules --id urn:review:rule --to 검토 --json workspace.json
```

명령은 원본을 읽고 변경 제안을 JSON으로 출력한다. `result.files`에 변경된 소스,
편집 위치와 원본·결과 해시가 있으며, `updated_workspace`에는 변경 후 진입점과
기존 소스 경로가 있다. 기본 출력은 대상 ID, 참조 수와 파일 목록을 보여준다.
예제는 [workspace-activity-rename](../../examples/workspace-activity-rename/README.md)에 있다.

## 함께 갱신하는 내용

- 선언 이름과 직접 `computes` 호출. import는 각 파일의 별칭으로 해석한다.
- 고정 `criteria`의 활동, `observe`의 기준 활동과 `bind`의 양 끝점.
- 선언한 표현 대안과 checkpoint의 baseline 안에 있는 호출.
- 대상이 진입점이면 워크스페이스의 entry 이름.
- ID를 명시하지 않았던 대상이면 변경 전에 해석한 ID를 소스에 기록한다.

소스에 선언했지만 entry에서 호출하지 않는 소비자도 조사한다. 같은 이름의 다른
패키지 활동과, 같은 별칭으로 다른 패키지를 import한 파일은 대상 ID와 소유 패키지로 구분한다.
일반 문자열, 주석, intent, 선택지 ID와 평가 입력은 기존 내용을 유지한다.
quoted 본문의 escape와 raw 본문은 실제 소스 바이트 위치로 편집한다.

이 과정은 변경 전후의 export ID·타입 역할과 분류한 참조의 ID를 비교한다.
전체 PROV 그래프나 외부 소비자의 동작을 검증했다는 의미까지 포함하지는 않는다.
기존 실행 기록은 원래 소스에 연결해 두며, 변경된 프로그램의 관찰에는 새 소스 ID가 필요하다.

## import한 기준 활동에 질문하기

`observe`의 `oracle`도 다른 활동 호출처럼 한 파일의 import 범위를 사용한다.

```gooo
import rates "questions/rates"
activity Price(Integer) -> Integer computes "return input-31" assembling {
    choice "direction" operand_order at "0" intent "31에서 입력을 뺀다"
    case "31" -> "0"
    holdout_case "42" -> "-11"
    observe {
        input "-8"
        oracle "rates.Expected"
        max_candidates "2"
        max_rounds "1"
        resolve_unique "true"
    }
    attempts "2"
}
```

여기서 `Expected`는 import한 패키지의 고정 `Integer -> Integer` 활동이다.
한정 이름은 소스에서 해석한 뒤 내부 활동 이름으로 옮긴다. 추가 질문의 답은 이 기준
활동에서 얻고, holdout은 선택 후 별도로 평가한다. 모델과 후보가 기대값을 정하는 흐름은 없다.

관찰 기록의 `rounds`에는 질문을 고르는 순위 계산과 마지막 생존 후보 계산이 모두 들어간다.
따라서 순위 기록 수와 답한 질문 수를 따로 센다. 아래 새 예제의 질문 1회에는
순위 기록 2개와 기준 활동 평가 2회가 있었다. 평가는 처음 case 1개와 추가 질문 1개에 해당한다.

## 이번 개발에서 확인한 범위

| 새 소스 과제 | 갱신한 참조 / 해석한 대상 참조 | 추가로 확인한 내용 |
| --- | --- | --- |
| CLI, 3패키지·5파일 | 5 / 5 | 3파일 변경, entry 밖 호출·조건식·bind 갱신, 다른 패키지와 일반 문자열 보존 |
| import한 기준 활동, 2패키지·2파일 | 1 / 1 | 고정 ID 유지, 질문 1회로 생존 후보 1개, 선택 후 별도 입력 1/1 일치 |
| 이름 기반 ID인 entry, 1패키지·2파일 | 1 / 1 | 기존 ID를 선언에 기록, entry와 entry 밖 소비자를 함께 갱신 |
| 실제 CLI의 일반 출력, 1패키지·2파일 | 1 / 1 | 대상 ID·2파일·entry 변경을 표시하고 원본 파일을 보존 |

각 행은 다른 새 소스를 한 번 관찰한 기록이다. rename은 프로그램 실행과 모델 예측을
각각 0회 수행했다. 기준 활동 예제의 조건 계산은 유한 Go AST 평가이며 네이티브 실행은 0회다.
생산자 소스와 결과는 [관찰 파일](../../examples/workspace-activity-rename/observation.json)에 연결했다.
관찰은 수행 당시 개발 소스에 연결되며, 이후 공개 커밋의 CI 결과와도 구분한다.
일반 출력 예제는 별도 새 reward 소스에서 CLI 실행 파일을 한 번 호출한 결과다.
CLI 프로세스 1회와 Gooo 소스 프로그램 실행 0회를 구분해 기록한다.

앞선 검사에서는 이름 기반 ID 누락, 잘못 작성한 checkpoint, 한정 oracle 거절을 발견했다.
질문 수와 순위 기록 수를 혼동한 테스트도 있었다. 원래 실패 기록은 보존했고,
수정한 동작의 로컬 관찰에는 각각 다른 새 소스를 사용했다.

## 범위와 다음 과제

지원 범위는 명시한 워크스페이스의 소스와 현재 분류할 수 있는 직접 참조다.
직접 선언한 `source_fill` 후보는 빈자리에 넣은 전체 본문에서 호출을 해석하고
원래 소스 위치로 연결한다. [후보 참조 안내](workspace-source-fill-refactoring.ko.md)에
별도 새 관찰과 범위를 기록했다. 동적 함수 값, parenthesized 호출, legacy 값 프로그램,
현재 자동 `derive`의 값·조건 문법은 활동 호출과 지역 선언을 만들지 않는 성질을 따라
본문 참조를 갱신한다. [파생 경로 안내](workspace-derived-fill-refactoring.ko.md)에 관찰을 기록했다.
새 문법 버전과 `search`의 후보별 lexical scope는 추가 위치 정보가 필요해 오류로 남긴다. 해석 실패와 이름 충돌은
부분 제안을 반환하지 않는다. 여러 단계의 import selector도 받지 않는다.

명령은 소스와 manifest를 자동으로 덮어쓰지 않는다. 외부 API 소비자, case 파일의 이름 키,
실제 에디터 적용과 변경된 네이티브 프로그램은 별도 범위다. source/interface 해시는 바뀔 수 있다.
다음 사용성 과제는 이 제안을 에디터와 소스 적용 도구가 같은 ID·해시 규칙으로 소비하는 것이다.

## 관련 언어에서 참고한 원리

[OML의 ontology와 import](https://opencaesar.io/oml/)는 모듈 namespace와 참조 별칭을 명시한다.
이 방식을 읽고 Gooo에서 확인할 과제로 삼은 것은 파일별 별칭이 어떤 활동 ID를 가리키는지,
그리고 이름 변경 때 entry 밖 소비자까지 따라갈 수 있는지다. OML과 동일한 온톨로지 의미론을
구현했다는 주장은 이 작업에 포함되지 않는다.

[관련 언어 비교](../ontology-language-comparison.ko.md)에서는 TypeQL, Nemo, Scallop과 OML을
각자의 계산·규칙·확률 태그·모듈 표현에 따라 살펴본다. 공통 업무 과제에서 필요한 외부 코드,
갱신한 참조, 별도 입력의 결과를 기록하는 방향으로 비교를 이어간다.
