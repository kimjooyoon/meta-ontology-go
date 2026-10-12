# Gooo 활동 이름 바꾸기

Gooo에서는 활동 이름과 의미의 ID를 따로 다룹니다. `Calculate`를 `요금`으로
바꿀 때는 같은 활동이라는 ID를 유지하고, 해당 파일이 소유한 호출과 조건도
새 이름을 가리키게 해야 합니다.

개발 소스에 `refactor rename-activity` 명령을 추가했습니다.

```sh
gooo refactor rename-activity --from Calculate --to 요금 examples/activity-rename/main.gooo.fixture > renamed.gooo
gooo refactor rename-activity --from Calculate --to 요금 --json examples/activity-rename/main.gooo.fixture
```

첫 번째 명령은 바뀐 소스를 출력합니다. 두 번째는 변경 기록과 소스를 JSON으로
출력합니다. **출력을 입력과 다른 파일에 저장하세요.** 같은 파일로 리다이렉션하면
셸이 먼저 입력 파일을 비웁니다.

## 함께 따라가는 것

- 기존 ID가 명시되지 않았다면, 현재 해석된 ID를 선언에 적어 둡니다.
- 본문의 직접 호출과 선언된 대안 표현식, 체크포인트 기준 본문을 바꿉니다.
- 파일 안의 바인딩, `criteria` 조건 활동, `observe` 관찰 활동 참조를 바꿉니다.
- 조건의 ID, 후보 선택 ID, 선택 라벨과 관찰 입력은 유지합니다.
- 주석과 설명 문장, 문자열 값, 필드 이름과 지역 변수 이름은 유지합니다.

같은 활동에 새 이름표를 붙이는 셈입니다. 새 이름이 지역 변수에 가려지거나
다른 선언과 충돌하면 변경안을 반환하지 않습니다. 결과를 다시 읽고 의미의 ID,
PROV 관계 ID, 실행 바인딩의 연결이 유지됐는지도 확인합니다.

## 변경 기록 읽기

`activity_id`는 그대로이고 `from`, `to`가 이름 변경을 설명합니다.
`input_sha256`, `output_sha256`은 서로 다른 두 소스를 구분합니다.
`edits`에는 바뀐 바이트 범위와 전후 내용이 있습니다.

기존 실행 기록은 이전 소스에 계속 연결됩니다. 이름을 바꾼 프로그램의 실행 결과는
새 소스에 대한 별도 관찰로 남겨야 합니다. 이 명령은 소스 변환 기록을 만듭니다.

## 현재 범위

한 파일, import 없는 소스, 입력 128 KiB까지 지원합니다. 일반적인 statement 본문과
명시적으로 선언한 조립 선택을 다룹니다. 생성형 `search`, `source_fill` 후보나
`computes "int.add:1"` 같은 이전 프로필의 값 프로그램에는 추가 매핑이 필요합니다.
대안 표현식의 지역 변수 범위를 확정할 수 없는 경우에도 멈춥니다.

다른 파일의 호출자, 워크스페이스 진입점, 케이스 파일의 입력 키와 외부 API 소비자는
별도로 연결해야 합니다. 실험적 LSP에서 활동 선언의 rename도 같은 변환을 사용합니다.
본문 호출에서 활동 ID를 확정하지 못하면 선언 위치에서 시작하도록 안내합니다.
본문 호출에서 시작하는 탐색과 워크스페이스 전체 rename은 다음 범위입니다.

[작은 예제와 실제 변경 기록](../../examples/activity-rename/README.md),
[영문 참조 문서](activity-refactoring.md)에 현재 근거를 남겼습니다.
한글 이름은 공개 결정 런타임
[v0.2.48-experimental](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.48-experimental)로
이어집니다. 모델이 활동 분리나 이름 변경을 제안할 때도 이처럼 Gooo가 변경 범위와
연결 관계를 확인하는 소스 연산이 바탕이 됩니다.
