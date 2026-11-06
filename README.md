# FuckLang

> A lightweight, Turing-complete scripting language where the keywords are English profanity.

---

## Overview

FuckLang is a small, cross-platform interpreted language. It was built with a simple goal: prove that a language doesn't need a massive ecosystem or complex syntax to be Turing-complete and actually fun to write. It supports variables, control flow, functions, closures, arrays, and hashes, while stripping away unnecessary punctuation. 

The interpreter is written in Go, meaning it compiles down to a single, dependency-free binary for macOS, Linux, and Windows.

---

## Features

- **Turing Complete** — Variables, `damn` (if/else), `hell` (loops), and recursion
- **First-Class Functions** — Define anonymous or named functions with `bitch`, return with `ass`
- **Closures** — Functions capture their lexical scope
- **Data Structures** — Arrays `[1, 2, 3]` and Hashes `{"key": "value"}` with index assignment
- **Built-in Functions** — `len`, `push`, `type`, `readFile`, `writeFile`
- **Static Error Handling** — Prevents silent failures with line-number tracking for syntax and runtime errors
- **REPL** — Interactive shell included out of the box
- **Cross-Platform** — Native binaries for macOS (Intel & Apple Silicon), Linux, and Windows

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

---

## Syntax Cheatsheet

| Feature | Keyword | Example |
| :--- | :--- | :--- |
| Variable | `shit` | `shit x = 42` |
| Print | `fuck` | `fuck x` |
| Condition | `damn` / `else` | `damn x > 5 { fuck "big" }` |
| Loop | `hell` | `hell x < 10 { shit x = x + 1 }` |
| Function | `bitch` | `bitch add(a, b) { ass a + b }` |
| Return | `ass` | `ass a + b` |

---

## Testing & CI

The project includes a comprehensive test suite covering arithmetic, logic, loops, closures, and error states. 

To run the tests locally:
```bash
go test -v ./...
```

A GitHub Actions workflow is included to run tests and cross-compile binaries for all three major platforms on every push.

---

## Build from Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/your-username/FuckLang.git
cd FuckLang
go build -o fucklang ./cmd/fucklang
```

To cross-compile the binaries:
```bash
GOOS=linux GOARCH=amd64 go build -o fucklang-linux ./cmd/fucklang
GOOS=windows GOARCH=amd64 go build -o fucklang.exe ./cmd/fucklang
GOOS=darwin GOARCH=arm64 go build -o fucklang-mac ./cmd/fucklang
```

---

## Contributing

Contributions are welcome! 

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

This project is licensed under the [MIT License](LICENSE).