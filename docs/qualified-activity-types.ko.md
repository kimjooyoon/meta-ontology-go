# Import로 활동의 타입 고르기

두 패키지가 `Amount`라는 이름을 써도 서로 다른 타입일 수 있다.
활동의 입력과 결과에 import 별칭을 붙이면 어느 패키지의 타입인지 지정할 수 있다.

```gooo
package app
namespace qualified_app
import 청구 "example/left"
import 보조 "example/right"

activity Main(청구.Amount) -> 청구.Amount id "urn:qualified:entry" computes "return 청구.요금(input)"
```

이 문법은 이 변경을 포함한 개발 소스에서 사용할 수 있다.
`청구.Amount`는 해당 파일에서 `청구`로 import한 패키지의 entity를 가리킨다.
타입 이름 자체가 타입 ID가 되지는 않는다. 패키지에서 내보낸 안정된 entity ID를 사용한다.
입력과 결과 모두 `별칭.타입`을 지원하며, 별칭과 타입 이름에 Unicode를 쓸 수 있다.
별칭을 생략한 이름의 기존 해석 규칙도 유지한다.

참조를 다시 편집할 때에는 [전체 소스 포맷 경로](source-preserving-format.ko.md)가
import와 계산 본문을 함께 남긴다. 타입 해석과 포맷은 각자의 소스 관찰을 기록한다.

## 별칭의 범위는 파일

같은 패키지의 다른 파일에서 `청구`를 다른 패키지에 연결할 수 있다.
각 활동의 타입은 그 활동이 선언된 파일의 import를 따라 해석한다.
한 파일의 별칭을 다른 파일의 타입 환경에 덮어쓰지 않는다.
한 파일에서 같은 별칭을 서로 다른 경로에 연결하면 import 모호성 오류가 난다.

선언하지 않은 별칭은 `PACKAGE_TYPE_IMPORT_UNKNOWN`으로 거절한다.
별칭이 가리키는 패키지에서 entity를 찾지 못하면 `PACKAGE_IMPORTED_TYPE_INVALID`가
원래 참조와 패키지 경로를 포함한다. 별칭을 생략한 타입이 여러 ID와 일치하면
기존 `PACKAGE_TYPE_AMBIGUOUS` 규칙이 적용된다.

## 내부 변환에서도 ID 유지

패키지 분석은 선택한 entity ID를 활동의 입력·결과 타입으로 기록한다.
분석용 IR에서는 그 ID에 연결된 기존 entity를 재사용하거나 내부 이름을 부여한다.
이미 같은 ID를 가진 entity가 있으면 별칭마다 새 entity를 중복해서 만들지 않는다.
내부 이름의 충돌은 피하고, 원본 파일은 변경하지 않는다.

워크스페이스를 하나의 소스로 준비할 때에도 타입 ID를 따라 이름을 정리한다.
이름이 같은 record가 서로 다른 ID를 가지면 기존 record 이름 정규화가 둘을 구분한다.
활동의 ID와 record의 ID는 각각 그대로 유지한다.

## 실제로 남긴 소스 관찰

[예제](../examples/qualified-signature-types/README.md)는 세 패키지와 네 소스 파일을 사용한다.
앱의 두 파일은 `청구`를 각각 left와 right 패키지에 연결한다.

| 활동 | 선언된 입력·결과 | 해석된 타입 ID |
| --- | --- | --- |
| `Main` | `청구.Amount` | `urn:qualified:left:amount` |
| `Alternate` | `청구.Amount` | `urn:qualified:right:amount` |

새 예제를 한 번 준비한 결과, 위 두 활동의 입력·결과 참조 네 개가 해당 ID로 기록됐다.
entry의 한국어 호출 한 개는 `example/left:요금`으로 연결됐다.
두 record는 서로 다른 내부 이름과 원래 ID를 유지했다. 입력 파일은 변경되지 않았다.
개발용 바이너리, 원래 소스와 출력 해시는 예제의 관찰 파일에 남겼다.

프로그램 실행, 모델 예측, 학습은 수행하지 않았다. 실제 값의 전달과 실행 결과,
저장한 실행의 재사용, 에디터의 모듈 참조 기능은 별도의 관찰이 필요하다.
관찰 후의 Go map 복사 문법 정리와 최종 컴파일은 따로 기록했으며,
예제를 다시 실행하거나 관찰 바이너리의 소스 기준을 바꾸지 않았다.

## 다음 표현 범위

현재 지원 범위는 활동 서명의 `별칭.entity`다. 여러 단계의 점 경로,
entity 필드의 한정 타입 표기, 타입 자체의 재내보내기는 후속 과제다.
워크스페이스 전체 rename과 LSP의 import 타입 탐색도 아직 연결되지 않았다.
기존 LSP는 단일 파일의 의미 해석이 실패하면 rename을 거절한다.

[OML의 모듈 참조 규칙](https://opencaesar.io/oml/)은 import 범위를 명시하는 문제를
살펴보는 데 참고했다. Gooo에서는 활동 서명의 파일별 별칭과 기존 entity ID에 적용했다.
