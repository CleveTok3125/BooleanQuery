package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func benchDataPath() string {
	// Try common locations for testdata/bench.log
	candidates := []string{
		"testdata/bench.log",
		"../testdata/bench.log",
		filepath.Join("..", "testdata", "bench.log"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "testdata/bench.log"
}

var benchText = []byte("2026-06-13 10:30:45 ERROR [app.server] critical: database connection timeout after 30s (host=db-01, db=prod)")
var benchTextStr = string(benchText)

func BenchmarkCheckOnlyBytes(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{}, "error -timeout ~critical ~fatal")
	for b.Loop() {
		e.CheckOnlyBytes(benchText)
	}
}

func BenchmarkCheckOnlyBytes_ExactWord(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{ExactWord: true}, "error")
	text := []byte("error_handler: no error found")
	for b.Loop() {
		e.CheckOnlyBytes(text)
	}
}

func BenchmarkSearch(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{}, "error database")
	for b.Loop() {
		e.Search(benchTextStr)
	}
}

func BenchmarkSearch_MultipleTerms(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{}, "error timeout database connection")
	for b.Loop() {
		e.Search(benchTextStr)
	}
}

func BenchmarkParseWildcard(b *testing.B) {
	b.ReportAllocs()
	patterns := []string{
		"simple",
		"wild*card",
		"he?lo*wo?ld",
		"a*b*c*d*e",
		"@*escape@?test",
		"*prefix*suffix*",
		"a?b?c?d?e?f",
	}
	for b.Loop() {
		for _, p := range patterns {
			parseWildcard(p)
		}
	}
}

func BenchmarkHighlightTo(b *testing.B) {
	b.ReportAllocs()
	e := New()
	matches := [][2]int{{0, 4}, {11, 19}, {29, 37}}
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e.HighlightTo(&buf, benchTextStr, matches)
	}
}

func BenchmarkHighlightTo_NoColor(b *testing.B) {
	b.ReportAllocs()
	e := &Engine{Config: Config{NoColor: true}}
	matches := [][2]int{{0, 4}, {11, 19}, {29, 37}}
	var buf bytes.Buffer
	for b.Loop() {
		buf.Reset()
		e.HighlightTo(&buf, benchTextStr, matches)
	}
}

func benchStream(b *testing.B, lines int, lineLen int) {
	b.Helper()
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	line := strings.Repeat("x", lineLen) + "\n"
	input := strings.Repeat(line, lines)

	b.ResetTimer()
	for b.Loop() {
		iter := e.ProcessStream(strings.NewReader(input), CHARSEP)
		iter(func(_ int, _ []byte) bool { return true })
	}
}

func BenchmarkProcessStream_100Lines_Short(b *testing.B) {
	b.ReportAllocs()
	benchStream(b, 100, 50)
}

func BenchmarkProcessStream_1000Lines_Long(b *testing.B) {
	b.ReportAllocs()
	benchStream(b, 1000, 500)
}

func BenchmarkProcessStream_WordsMode(b *testing.B) {
	b.ReportAllocs()
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024
	input := strings.Repeat("word1 word2 word3\n", 500)

	for b.Loop() {
		iter := e.ProcessStream(strings.NewReader(input), WORDS)
		iter(func(_ int, _ []byte) bool { return true })
	}
}

func BenchmarkClassify(b *testing.B) {
	b.ReportAllocs()
	queries := []string{
		"error -timeout ~critical ~fatal ^func ^main",
		"database connection timeout error warning",
		"+start +middle -end ~maybe",
	}
	for b.Loop() {
		for _, q := range queries {
			e := New()
			e.SetSearchTerm(q)
			e.Classify()
		}
	}
}

func BenchmarkFindTermIndex(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{}, "connection")
	term := e.searchTerm.whiteList[0]
	for b.Loop() {
		e.findTermIndex(benchText, term, 0)
	}
}

func BenchmarkFindTermIndex_Wildcard(b *testing.B) {
	b.ReportAllocs()
	e := newTestEngine(Config{Wildcard: true}, "*connect*timeout")
	term := e.searchTerm.whiteList[0]
	for b.Loop() {
		e.findTermIndex(benchText, term, 0)
	}
}

func BenchmarkFileSearch(b *testing.B) {
	b.ReportAllocs()
	path := benchDataPath()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skip("testdata/bench.log not found, run: go run ./cmd/genbench")
	}
	lines := strings.Split(string(data), "\n")
	b.ResetTimer()

	e := newTestEngine(Config{}, "ERROR timeout")
	for b.Loop() {
		for _, line := range lines {
			e.CheckOnlyBytes([]byte(line))
		}
	}
}

func BenchmarkFileSearch_ExactWord(b *testing.B) {
	b.ReportAllocs()
	path := benchDataPath()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skip("testdata/bench.log not found, run: go run ./cmd/genbench")
	}
	lines := strings.Split(string(data), "\n")
	b.ResetTimer()

	e := newTestEngine(Config{ExactWord: true}, "ERROR")
	for b.Loop() {
		for _, line := range lines {
			e.CheckOnlyBytes([]byte(line))
		}
	}
}

func openBenchFile(b *testing.B) *os.File {
	path := benchDataPath()
	f, err := os.Open(path)
	if err != nil {
		b.Skip("testdata/bench.log not found, run: go run ./cmd/genbench")
	}
	return f
}

func BenchmarkFileStreamProcess(b *testing.B) {
	b.ReportAllocs()
	e := New()
	e.Config.BufferSize = 64 * 1024
	e.Config.BufferMaxSize = 1024 * 1024

	if err := e.SetSearchTerm("ERROR timeout"); err != nil {
		b.Fatal(err)
	}
	e.Classify()

	b.ResetTimer()
	for b.Loop() {
		f := openBenchFile(b)
		iter := e.ProcessStream(f, CHARSEP)
		iter(func(_ int, part []byte) bool {
			e.CheckOnlyBytes(part)
			return true
		})
		f.Close()
	}
}

func BenchmarkFileStreamProcess_Wildcard(b *testing.B) {
	b.ReportAllocs()
	e := New()
	e.Config.Wildcard = true
	e.Config.BufferSize = 64 * 1024
	e.Config.BufferMaxSize = 1024 * 1024

	if err := e.SetSearchTerm("*connect*timeout*"); err != nil {
		b.Fatal(err)
	}
	e.Classify()

	b.ResetTimer()
	for b.Loop() {
		f := openBenchFile(b)
		iter := e.ProcessStream(f, CHARSEP)
		iter(func(_ int, part []byte) bool {
			e.CheckOnlyBytes(part)
			return true
		})
		f.Close()
	}
}
