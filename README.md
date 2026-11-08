# FuckLang

> A lightweight, Turing-complete scripting language where the keywords are English profanity.

---

## Overview

FuckLang is a small, cross-platform interpreted language. It was built with a simple goal: prove that a language doesn't need a massive ecosystem or complex syntax to be Turing-complete and actually fun to write. It supports variables, control flow, functions, closures, arrays, and hashes, while stripping away unnecessary punctuation.

The interpreter is written in Go, meaning it compiles down to a single, dependency-free binary for macOS, Linux, and Windows.

---

## Features

- **Turing Complete** — Variables, `damn` (if/else/else-if), `hell` (loops) with `break`, and recursion
- **First-Class Functions** — Define anonymous or named functions with `bitch`, return with `ass`
- **Closures** — Functions capture their lexical scope
- **Data Structures** — Arrays `[1, 2, 3]` and Hashes `{"key": "value"}` with index read/write, including hash key insertion
- **Built-in Functions** — `len` (String / Array / Hash), `push`, `type`, `readFile`, `writeFile`
- **Error Handling** — Syntax errors are collected with line numbers; runtime errors abort with a clear message **and a line number**
- **Comments** — `//` line comments
- **String Escapes** — `\"`, `\\`, `\n`, `\t`, `\r`
- **REPL** — Interactive shell with multi-line block support (bracket-balanced input, `//` comments skipped)
- **Cross-Platform** — Native binaries for macOS (Intel & Apple Silicon), Linux, and Windows; test suite runs reliably in CI

---

## Usage

Create a file named `hello.fuck`:

```fuck
shit name = "world"
fuck "hello " + name
```

Run it via the CLI:

```bash
./fucklang hello.fuck
```

Or launch the interactive REPL:

```bash
./fucklang
```

The REPL supports multi-line blocks: if you open `{`, `(` or `[`, it keeps reading until brackets are balanced. Line comments (`//`) are ignored during bracket balancing, so a `{` inside a comment won't trap you.

---

## Syntax Cheatsheet

| Feature | Keyword / Syntax | Example |
| :--- | :--- | :--- |
| Variable | `shit` | `shit x = 42` |
| Print | `fuck` | `fuck x` |
| Condition | `damn` / `else` / `else damn` | `damn x > 5 { fuck "big" } else { fuck "small" }` |
| Loop | `hell` | `hell x < 10 { shit x = x + 1 }` |
| Break | `break` | `hell true { damn x > 3 { break } }` |
| Function | `bitch` | `bitch add(a, b) { ass a + b }` |
| Anonymous Function | `bitch(...)` | `shit f = bitch(x) { ass x * 2 }` |
| Return | `ass` | `ass a + b` |
| Array | `[...]` | `shit a = [1, 2, 3]` |
| Hash | `{...}` | `shit h = {"name": "Alice"}` |
| Index Read | `x[i]` | `fuck arr[0]` / `fuck user["name"]` |
| Index Write | `x[i] = v` | `arr[0] = 9` / `user["role"] = "admin"` |
| Comment | `//` | `// this is a comment` |
| Statement Separator | `;` (optional) | `shit x = 5; fuck x` |

### Break

`break` exits the innermost enclosing `hell` loop immediately. It cannot cross a function boundary, and using it outside any loop is a runtime error.

```fuck
shit i = 0
shit sum = 0
hell true {
    damn i >= 5 {
        break
    }
    shit sum = sum + i
    shit i = i + 1
}
fuck sum    // 10
```

### Index Assignment

Both arrays and hashes support index assignment. For hashes, assigning to a new key inserts it:

```fuck
shit arr = [1, 2, 3]
arr[1] = 20
fuck arr[1]          // 20

shit user = {"name": "Alice"}
user["role"] = "admin"
fuck user["role"]    // admin
fuck len(user)       // 2
```

### Else If

```fuck
shit n = 5
damn n > 10 {
    fuck "big"
} else damn n > 3 {
    fuck "medium"
} else {
    fuck "small"
}
```

### Built-ins

| Name | Signature | Description |
| :--- | :--- | :--- |
| `len` | `len(x)` | Length of String, Array, or Hash |
| `push` | `push(arr, v)` | Append `v` to Array `arr` **in place**; returns the same array |
| `type` | `type(x)` | Returns the type name as a String (e.g. `"INTEGER"`, `"HASH"`) |
| `readFile` | `readFile(path)` | Read a file as String |
| `writeFile` | `writeFile(path, content)` | Write String to file (0644) |

---

## Semantics Notes

Some behaviours that are intentional but easy to be surprised by:

- **Integers only.** All numbers are signed 64-bit. `/` is integer division (`5 / 2 == 2`). There are no floats.
- **`push` mutates in place.** `shit b = push(a, 1)` makes `a` and `b` refer to the same array.
- **Loop variables are captured by reference.** Closures created inside a `hell` loop share the same loop variable. If you need per-iteration capture, copy the variable into a fresh binding inside the loop body.
- **Hash printing order is unspecified.** `Hash.Inspect` iterates a Go map, so key order is not stable. Do not rely on it.
- **Array index assignment does not grow the array.** `arr[i] = v` requires `0 <= i < len(arr)`; out-of-bounds writes are runtime errors.
- **`break` cannot cross a function boundary.** A `break` inside a function body that is called from a loop is a runtime error, not a loop exit.
- **No immediate function invocation.** You cannot call an anonymous function directly (e.g., `bitch() { ... }()` is a syntax error). Assign it to a variable first, then call it.

---

## Testing & CI

The project has two layers of tests:

1. **Evaluator logic tests** (`tests/tests_test.go`) — inline source snippets covering arithmetic, strings, comments, conditions (including `else if`), variables, built-ins, hashes, index assignment, functions, closures, argument-count errors, loops, and `break`.
2. **Black-box `.fuck` tests** — every `tests/*.fuck` file is executed, its stdout is captured, and compared byte-for-byte against the matching `tests/*.expected` file. Syntax errors, runtime errors, and output mismatches all fail the test.

To run the tests locally:

```bash
go clean -testcache
go test -v ./...
```

> **Line endings matter.** `.fuck` and `.expected` files are compared byte-for-byte, so they must be checked out with LF. The repository's `.gitattributes` enforces `eol=lf` for both extensions. If you edit them on Windows, make sure your editor doesn't save CRLF.

CI runs the full test suite on **Ubuntu** (GitHub Actions) for fast, deterministic feedback. After tests pass, a separate job cross-compiles binaries for Linux, Windows, and macOS, and uploads them as artifacts.

---

## Build from Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/CaoHaoran-Dev/FuckLang.git
cd FuckLang
go build -o fucklang ./cmd/fucklang
```

To cross-compile the binaries:

```bash
GOOS=linux   GOARCH=amd64 go build -o fucklang-linux        ./cmd/fucklang
GOOS=windows GOARCH=amd64 go build -o fucklang.exe          ./cmd/fucklang
GOOS=darwin  GOARCH=arm64 go build -o fucklang-mac-arm64    ./cmd/fucklang
GOOS=darwin  GOARCH=amd64 go build -o fucklang-mac-amd64    ./cmd/fucklang
```

Or use the Makefile:

```bash
make build     # local build
make test      # clean test cache and run all tests
make release   # cross-compile all platforms into build/
make clean     # remove build artifacts
```

---

## Project Layout

```text
.
├── cmd/fucklang/main.go          # CLI entry + REPL
├── internal/
│   ├── ast/ast.go                # AST nodes
│   ├── token/token.go            # Token types + keywords
│   ├── lexer/lexer.go            # Lexer (comments, line tracking, escapes)
│   ├── parser/parser.go          # Pratt parser + error collection
│   ├── evaluator/
│   │   ├── evaluator.go          # Tree-walking evaluator
│   │   └── builtins.go           # len / push / type / readFile / writeFile
│   └── object/object.go          # Runtime values
├── tests/
│   ├── *_test.go                 # Logic tests + black-box runner
│   ├── *.fuck                    # Black-box programs
│   └── *.expected                # Expected stdout for each .fuck
├── .github/workflows/            # CI (Ubuntu tests + cross-compile)
├── .gitattributes                # LF normalization for .fuck / .expected
├── Makefile
├── go.mod
└── README.md
```

---

## Contributing

Contributions are welcome!

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

If you add a new `.fuck` black-box test, remember to also add the matching `.expected` file — otherwise `TestFuckFiles` will fail with a "missing expected file" error.

This project is licensed under the [MIT License](LICENSE).