# BooleanQuery

**BooleanQuery** is a high-performance,
concurrent CLI text search tool written in Go.
It allows searching using complex boolean logic (AND, OR, NOT)
directly within a simple query string.

## Key Features

* **Advanced Boolean Logic**:
Combine **AND**, **NOT**, and **OR** (GreyList) logic in a single query.
* **High Concurrency**: Process multiple files in parallel.
* **Context Printing**:
Viewing lines before (`-B`), after (`-A`), or surrounding (`-C`) the match.
* **Smart Highlighting**:
Automatically colors matched terms for better readability.

## Installation

**Prerequisites:** Go 1.25.5 or higher.

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

3. **OR / GreyList (`~`)**:
If "GreyList" terms are present, the line must contain **at least one** of them.

    * Example: `"struct ~json ~xml"`
    * *Matches lines containing "struct" AND (either "json" OR "xml").*

## ToDo

* [ ] **Recursive Search**: Support for scanning directories recursively.
* [ ] **Binary File Detection**:
Automatically detect and skip binary files to prevent terminal corruption.
* [ ] Add `-c` (`--count`) flag to print only the count of matching lines.
* [ ] Add (`--count-files`) flag to print only matched files.
* [ ] Add column indices for matches.
