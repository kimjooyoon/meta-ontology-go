# 원래 파일 위치를 설명하는 Gooo 프로그램

`local-input-locations` 행을 받아 파일명·실행 사건·기원 종류를 설명한다.
원래 파일 위치가 없는 행은 그 이유를 반환한다. 읽기 규칙은
[reader.gooo](reader.gooo)의 `Describe` 활동에 있다.

먼저 [작은 생산 프로그램](../../workspace-local-origins/README.md)을 실행한다.
그 저장 기록을 `--origin-receipt`로 이 폴더의 `reader.workspace.json`에 전달한다.
원래 생산 프로그램은 다시 실행하지 않는다.

입력은 정수·문자열 23필드다. `source_sha256`와 `source_start/end`는 원래
워크스페이스 파일을 가리킨다. 기존 `local-writes`와 `local-uses`의 같은 이름 필드는
선택된 합성 Gooo 소스를 가리키므로, 입력 종류마다 맞는 프로그램을 사용한다.
`origin_input.selected_source_sha256`에는 선택된 합성 소스의 해시가 따로 남는다.

`declaration`, `value`, `use`는 정확한 소스 구간이다. `declared_candidate`,
`derived_grammar`, `retained_template`는 그 계산을 구성한 문맥의 구간일 수 있다.
문법에서 생성한 식은 문법 선언에 연결하며 원문에 없는 식의 위치는 만들지 않는다.
예전 기록에 파일 정보가 없으면 `UNMAPPED:ORIGINAL_FILE_ANNOTATION_ABSENT`를 반환한다.

새 생산 입력·설명 출력의 대조와 실제 범위는
[관찰 기록](../../../docs/observations/workspace-local-file-origins/README.md)에 있다.
