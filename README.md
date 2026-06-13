# BooleanQuery

**BooleanQuery** is a high-performance,
concurrent CLI text search tool written in Go.
It allows searching using complex boolean logic (AND, OR, NOT, ORDERED AND)
directly within a simple query string.

## Key Features

* **Advanced Boolean Logic**:
  * **AND**: All terms must be present.
  * **OR** (GreyList): At least one term must be present.
  * **NOT**: Term must not be present.
  * **ORDERED AND**: Terms must appear in a specific sequence.
* **Parallel Processing**: Process multiple files concurrently.
* **Context Printing**:
Viewing lines before (`-B`), after (`-A`), or surrounding (`-C`) the match.
* **Smart Highlighting**:
Automatically colors matched terms for better readability.
* **Customizable separator character**
* **Wildcard support**: Basic wildcard support: `*` and `?`

## Installation

**Prerequisites:** Go 1.25.4 or higher.

```fish
# Clone the repository
git clone https://github.com/CleveTok3125/BooleanQuery.git
cd BooleanQuery

# Build the binary
make build
```

### Install to system

```bash
sudo make install          # → /usr/local/bin/bq
```

### Install to user-local

```bash
make install BINDIR=$HOME/.local/bin
```

### Uninstall

```bash
sudo make uninstall
```

### Build manually (without make)

```fish
go build -o bq ./src/
```

## Usage

### Basic Syntax

```fish
./bq "QUERY_STRING" [FILE_PATHS...]
```

If no files are provided, `BooleanQuery` reads from stdin.

### Search Logic (Query Syntax)

The engine classifies search terms based on their prefix:

1. **AND (Default or `+`)**:
The line **must** contain this term.

    * Example: `"error database"` or `"+error +database"`
    * *Matches lines containing both "error" AND "database".*

2. **NOT (`-`)**:
The line **must not** contain this term.

    * Example: `"error -timeout"`
    * *Matches lines containing "error" but NOT "timeout".*

3. **OR (`~`)**:
If **OR** terms are present, the line must contain **at least one** of them.

    * Example: `"struct ~json ~xml"`
    * *Matches lines containing "struct" AND (either "json" OR "xml").*

4. **ORDERED AND (`^`)**:
The line must contain these terms in the **exact order** they appear in the query.

    * Example: `"^func ^main"`
    * *Matches `func main() { ... }*`
    * *Does NOT match `// main is called by func wrapper` (wrong order).*

### Notes

* Grouping can be done by using `'` (single quotes)
for strings containing special characters:
  * Example: `./bq "+'+2 cards' ~reverse"`
  * However, in some special cases,
    such as `*` matches `*` in wildcard,
    the escape sequence must be used.

* Problem with `\`:
  * Due to multiple layers of shell parsing,
  `\` is interpreted as an escape sequence and is escaped before reaching the query.
  * There is a workaround for this problem,
  you can use multiple `\` or `'` (single quotes).
  * For example, you can find the string `\query` by passing in `"\\\query"` or `"'\query'"`
  (double quotes surrounding single quotes).

* In wildcard mode, `*` and `?` are occupied and implicitly understood as query syntax:
  * To find `*` and `?` literally, use the escape character `@`.
  * *Why use `@` instead of `\`?* Due to the issue mentioned above.

* Only the first character of a query group is checked to see
if it's a valid query prefix.
  * You can search for strings containing a character matching the query prefix
  at the beginning of the query by specifying the query prefix in query:
  * `"++example"` will find strings containing `+example`
  * `"+-example"` will find strings containing `-example`
  * `"-+example"` will find strings NOT containing `+example`

* Using a NOT query (`-`) at the beginning of a query string
can cause the parser to misinterpret it as an argument.\
    This can be fixed by placing it after or grouping it with `'` (single quotes).

* To ensure the print order in parallel processing,
a disk buffer will be used and temporary file generated will be stored
in the temporary directory.

## Performance

Benchmarked at commit `9baa90c` on 2026-06-13.
Hardware: AMD Ryzen 5 7535HS, Linux amd64, Go 1.25.4.

### Unit Benchmarks (single-line matching)

| Benchmark | Time/op | Bytes/op | Allocs/op |
|---|---|---|---|
| CheckOnlyBytes | 41.2 ns | 0 | 0 |
| CheckOnlyBytes (ExactWord) | 24.1 ns | 0 | 0 |
| Search | 95.8 ns | 112 | 1 |
| Search (Multiple Terms) | 97.8 ns | 112 | 1 |
| HighlightTo (color) | 73.1 ns | 48 | 1 |
| HighlightTo (no-color) | 6.52 ns | 0 | 0 |
| findTermIndex | 27.3 ns | 0 | 0 |
| findTermIndex (wildcard) | 44.0 ns | 0 | 0 |
| ParseWildcard | 1.04 µs | 1,224 | 31 |

### Stream Processing (in-memory)

| Benchmark | Lines×Size | Time/op | Bytes/op | Allocs/op |
|---|---|---|---|---|
| ProcessStream | 100 × 50 B | 2.14 µs | 4,256 | 3 |
| ProcessStream | 1000 × 500 B | 32.5 µs | 4,256 | 3 |
| ProcessStream | WORDS mode | 50.3 µs | 4,256 | 3 |

### Query Classification

| Benchmark | Time/op | Bytes/op | Allocs/op |
|---|---|---|---|
| Classify (3 queries) | 683 µs | 25,000 | 171 |

### File Benchmarks (1 GB, 12 million lines)

Data generated with `cmd/genbench`:

```fish
go run ./cmd/genbench .test/bigbench.log 12000000
go test -bench=BenchmarkFile -benchmem ./src/engine/
```

| Benchmark | Time/op | Throughput | Bytes/op | Allocs/op |
|---|---|---|---|---|
| FileSearch | 801 ms | ~1.25 GB/s | 1.16 GB | 12 M |
| FileSearch (ExactWord) | 789 ms | ~1.27 GB/s | 1.16 GB | 12 M |
| FileStreamProcess | 696 ms | ~1.44 GB/s | 66 KB | 8 |
| FileStreamProcess (Wildcard) | 931 ms | ~1.07 GB/s | 66 KB | 8 |

> **Note:** File benchmarks allocate 12 M allocations because every line is converted to `string` for the `Split` call. For zero-alloc line-wise matching, use `ProcessStream` directly instead of pre-loading the entire file.

## Development

### Prerequisites

- Go 1.25.4 or higher
- GNU Make

### Commands

```makefile
make build     # Build the binary
make test      # Run all tests
make fmt       # Format source code
make vet       # Static analysis
make tidy      # Verify go.mod/go.sum are clean
make check     # Run all checks: format, vet, build, tidy
make clean     # Remove built binary
```

`make check` is the gate before CI — run it locally to ensure everything passes:

```bash
make check
```

## ToDo

* [x] **Recursive Search**: Support for scanning directories recursively.
* [x] **Binary File Detection**:
Automatically detect and skip binary files to prevent terminal corruption.
* [x] Add `-l` (`--count`) flag to print only the count of matching lines.
* [x] ~~Add (`--count-files`) flag to print only matched files.~~\
    Same results can be achieved by using `--files-with-matches` and `wc -l`.
* [x] Add column indices for matches.
* [x] Wildcard support
