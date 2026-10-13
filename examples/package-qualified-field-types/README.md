# 파일별 필드 타입과 파일 간 활동 연결

이 예제에서는 두 제공 패키지의 `Count`가 각각 정수와 불리언을 뜻한다.
앱의 두 파일은 자기 파일의 `data` import를 통해 필요한 타입을 읽는다.
`Pass.result -> Decide.input`은 같은 패키지의 두 파일에 나뉜 활동을 연결한다.

- `counts.gooo.fixture`: 정수 Count와 문자열 Label
- `flags.gooo.fixture`: 불리언 Count
- `ledger.gooo.fixture`: 정수 amount, 선택 문자열 note, Pass 활동
- `claim.gooo.fixture`: 불리언 eligible, Decide 활동과 연결
- `cases.json`: 입력 3개와 활동 출력 6개의 기대값
- `observation.json`: 내부 native 실행에서 얻은 원래 field270 관찰
- `actual-cli/`: 별도 field286 소스와 CLI 해석 관찰
- `unicode/`: 새 field352의 한글 파일명·별칭·타입 이름과 native 관찰
- `native-cli/`: 다른 field376 소스로 CLI 코드 생성·빌드·실행한 관찰

fixture는 Gooo 소스를 보관한다. 예제 파일의 `.fixture` 확장자와 외부 manifest는
게시용 포장이다. 원래 native 관찰에서는 테스트가 `.gooo` 파일명과 메모리 내
manifest를 만들었으며, 그 관찰의 producer·소스 digest는 `observation.json`에 남아 있다.
게시용 포장으로 원래 관찰을 다시 실행하거나 동일한 실행으로 표시하지 않았다.

별도 CLI 관찰은 `actual-cli` 디렉터리에서 다음 진입점을 사용했다.

```sh
gooo package resolve --json workspace.json
```

이 명령은 타입·연결·초기화 순서를 해석한다. 결과 image의
`packages[].sources[].field_types`에서 표기·제공 패키지·기본 타입 ID·소스 범위를,
`packages[].bindings[].source`에서 파일 간 연결의 범위와 활동 ID를 읽을 수 있다.

관찰에서 필드 3/3, 연결 1개가 유지됐고 입력 파일은 바뀌지 않았다.
활동 본문 실행과 모델 호출은 0회다. 프로세스 시작을 포함한 한 번의 시간은
585,400,583ns였고, 25,674,674바이트 임시 CLI는 hash를 기록한 뒤 제거했다.

native 관찰은 `ExecuteWorkspace`에서 생성·실행한 Go의 출력 6/6이 일치했으며,
반복 실행 결과도 같았다. 입력은 문자열 없음·빈 문자열·한글, 경계값 31과
큰 정수 `9007199254740995`를 포함한다. 두 관찰 모두 실제 제출·병합 커밋의 전체
호환성 검사를 대신하지 않으며 원래 producer에서 얻은 결과로 읽는다.

한글 field352 관찰은 별도 `959490f5` + 20 overlay에서 참조 범위 3/3, 연결 1개,
입력 2개의 출력 4/4와 반복 실행 일치를 확인했다. `자료.값`, `자료.문장`의 범위는
문자 수가 아닌 바이트 위치로 원본 문자열을 찾는다. 모델 호출은 0회다.
`unicode/`의 게시용 `.fixture` 파일명과 외부 manifest는 원래 테스트의 메모리 내
manifest와 구분해 읽는다. 원래 관찰은 테스트의 한글 `.gooo` 파일명에 묶여 있다.

`native-cli/`에서는 다음 명령을 실제 사용했다.

```sh
gooo package execute --json --cases cases.json workspace.json
```

새 field376의 참조 3/3와 연결 1개가 유지됐으며 입력 2개의 출력 4/4, native 2회의
실행 결과가 일치했다. 모델 호출은 0회다. 시작·생성·빌드·실행을 포함한 한 번의
시간은 1,439,670,917ns였다. 빌드 peak RSS는 87,195,648바이트, native 프로세스는
각각 4,947,968바이트·4,915,200바이트였다. 원래 입력 파일은 유지했고,
25,674,978바이트 임시 CLI를 hash 기록 뒤 제거했다. 이 관찰의 분모와 시간은 앞선
관찰에서 얻은 결과와 별도로 읽는다.

[언어 가이드](../../docs/language/package-qualified-field-types.ko.md)에 현재 범위와
진단, 비교 언어에서 가져온 관점을 정리했다.
