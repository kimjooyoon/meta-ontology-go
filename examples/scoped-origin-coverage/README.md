# 기원 조각의 연결 범위를 Gooo로 집계하기

`coverage.gooo`는 `for let row=0;row<256;row++`로 저장된 기원 조각을 읽습니다.
템플릿·선택 후보·유도 문법의 수와 위치가 연결된 행의 비율을 Gooo에서 계산합니다.
원래 업무 프로그램을 다시 실행하는 횟수는 0입니다.

저장소 루트에서 Go 1.27.3를 사용합니다.

```sh
saved="$(mktemp)"
gzip -dc examples/scoped-origin-coverage/recorded-runtime.json.gz > "$saved"
go run ./cmd/gooo package execute --json \
  --origin-receipt "$saved" --origin-shape local-fill-columns \
  --origin-limit 128 examples/scoped-origin-coverage/gooo.workspace.json
rm "$saved"
```

결과의 `result.runtime.traces[0].deliveries[0].actual`은 기록된 40행,
연결된 40행, 후보 12행, 템플릿 28행과 `RECORDED_FRAGMENTS`를 보여 줍니다.
`row_mapping_basis_points=10000`은 선택된 조각 행의 위치 연결 비율입니다.
언어 전체나 업무 계산의 완전성 비율로 사용하지 않습니다.

`--origin-offset 2`를 추가하면 38행과 `PARTIAL_PAGE`를 얻습니다.
`whole_recording=false`를 유지하므로 잘 연결된 일부 페이지를 전체 기록의
결과로 읽지 않습니다. 입력만 받은 소비자의 `finite_total`은 0입니다.

이 기록의 원래 생산 실험은 조립 선택용 사례 하나와 별도의 실행 입력 두 개를
가집니다. 현재 예제는 이미 저장된 결과만 읽습니다. 모델 설정 없이 동작합니다.

## 문법과 수정

카운터는 반복문 안에서만 유효합니다. `len`과 비교하는 부분의
`int64(len(input.kinds))`는 현재 정수 표현 차이를 명시합니다.
종류를 세는 규칙과 출력 상태는 이 Gooo 파일에서 수정할 수 있습니다.

[언어 범위](../../docs/language/scoped-counted-loops.ko.md) ·
[실행·초기 실패 기록](../../docs/observations/scoped-counted-loops/README.md)
