# Bounded Text byte search in Gooo

Gooo로 문자열을 한 바이트씩 읽어 원하는 값의 위치를 찾는 예제입니다.
작은 검색 함수를 다른 Gooo 패키지에서 가져오고, 저장한 프로그램도 다시 실행합니다.

This example runs with [Gooo 0.6.31-dev](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.31-dev).
The [release guide](../../docs/releases/0.6.31-dev.md) identifies its source and
four native platform checks. Generated code execution needs Go 1.27.2.

`search.gooo.fixture` implements byte search in Gooo. `app.gooo.fixture` imports it
and passes a declared `Query` record. The loop counter is a local initialized to
zero, advances by one and has the literal bound 1,024. It checks the text length
before reading each byte. A missing byte or a needle outside 0..255 returns -1.

Byte positions count UTF-8 bytes. For example, `x` in `가x` starts at byte 3, while
the second byte of `가` has value 176 at position 1. The input contract accepts
valid UTF-8 Text up to 1,024 bytes, including NUL. It does not represent arbitrary
Go byte strings.

```sh
gooo package execute --json --cases cases.json gooo.workspace.json > execution.json
gooo package replay --json --receipt execution.json --cases cases.json gooo.workspace.json
```

The repeated cases check saved program replay. The result's source, package,
generated program and input identities are retained. No model is requested.

The first development validation compared the interpreter with Go
`strings.IndexByte` across five texts and 258 needle values (-1..256): 1,290 finite
combinations. Ten native importing-package cases and their saved replay passed.
These counts describe this authored example, rather than all string operations.
The native example includes empty text, absence, repeated bytes, NUL, multibyte
byte positions, the 1,024-byte boundary and invalid needle values.

The counted-loop profile requires `for i = 0; i < N; i = i + 1`, with an existing
local and literal `N` from 1 to 1,024. Its body can return early, read inputs and
update other locals. Counter mutation, counter shadowing, nested loops, dynamic
bounds and unbounded loops are rejected. Calls inside the loop contribute their
maximum repetitions to the existing 4,096-call closure bound.

Text indexing yields an unsigned byte. Convert it with `int64(text[offset])` when
returning an Integer. Byte arithmetic keeps Go's unsigned byte wrap behavior;
out-of-bounds evaluation is an error. More general iteration, byte-array outputs,
bitwise operations, hex and UUID parsing remain follow-up work.
