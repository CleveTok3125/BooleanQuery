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
go build -o bq .
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
  * Due to multiple layers of shell parsing,\
  `\` is interpreted as an escape sequence and is escaped before reaching the query.
  * There is a workaround for this problem,\
  you can use multiple `\` or `'` (single quotes).
  * For example, you can find the string `\query` by passing in `"\\\query"` or `"'\query'"`\
  (two double quotes surrounding a single quote).

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

## ToDo

* [x] ~~**Recursive Search**: Support for scanning directories recursively.~~\
    Alternative solutions such as `find` and `shell grobbing` exist.
* [x] **Binary File Detection**:
Automatically detect and skip binary files to prevent terminal corruption.
* [x] Add `-l` (`--count`) flag to print only the count of matching lines.
* [x] ~~Add (`--count-files`) flag to print only matched files.~~\
    Same results can be achieved by using `--files-with-matches` and `wc -l`.
* [x] Add column indices for matches.
* [x] Wildcard support
