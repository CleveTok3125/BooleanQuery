package engine

import (
	"strings"
	"testing"
)

func collectStream(e *Engine, input string, flag IntFlag) ([]string, error) {
	var result []string
	iter := e.ProcessStream(strings.NewReader(input), flag)
	err := iter(func(_ int, part []byte) bool {
		result = append(result, string(part))
		return true
	})
	return result, err
}

func TestProcessStream_NewlineSepDefault(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "line one\nline two\nline three\n"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"line one", "line two", "line three"}
	if len(result) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(result), len(want), result)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, result[i], want[i])
		}
	}
}

func TestProcessStream_EmptyInput(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	result, err := collectStream(e, "", CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 lines, got %d: %v", len(result), result)
	}
}

func TestProcessStream_SingleCharSep(t *testing.T) {
	e := New()
	e.Config.CharSep = ":"
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "a:b:c:d"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a", "b", "c", "d"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Errorf("part[%d] = %q, want %q", i, result[i], want[i])
		}
	}
}

func TestProcessStream_MultiCharSep(t *testing.T) {
	e := New()
	e.Config.CharSep = ",\n"
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "a,b\nc,d"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a", "b", "c", "d"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Errorf("part[%d] = %q, want %q", i, result[i], want[i])
		}
	}
}

func TestProcessStream_NoTrailingSep(t *testing.T) {
	e := New()
	e.Config.CharSep = ":"
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "a:b:c"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
}

func TestProcessStream_TrailingNewline(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	result, err := collectStream(e, "hello\n", CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"hello"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
}

func TestProcessStream_YieldStops(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "a\nb\nc\n"
	var result []string
	iter := e.ProcessStream(strings.NewReader(input), CHARSEP)
	err := iter(func(_ int, part []byte) bool {
		result = append(result, string(part))
		return false // stop after first
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected only ['a'], got %v", result)
	}
}

func TestProcessStream_WordsMode(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "hello world\nfoo bar baz\n"
	result, err := collectStream(e, input, WORDS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"hello", "world", "foo", "bar", "baz"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Errorf("word[%d] = %q, want %q", i, result[i], want[i])
		}
	}
}

func TestProcessStream_WordsModeWithExtraSpaces(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "  hello   world  \n  foo  "
	result, err := collectStream(e, input, WORDS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"hello", "world", "foo"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
}

func TestProcessStream_WordsModeYieldStop(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "a b c"
	var result []string
	iter := e.ProcessStream(strings.NewReader(input), WORDS)
	err := iter(func(idx int, part []byte) bool {
		result = append(result, string(part))
		return idx < 1 // stop after 2 items (idx 0 and 1)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 words, got %v", result)
	}
}

func TestProcessStream_BinaryDetection(t *testing.T) {
	e := New()
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	// Null byte in content should trigger binary detection
	input := "hello\x00world\n"
	_, err := collectStream(e, input, CHARSEP)
	if err != ErrBinaryFile {
		t.Errorf("expected ErrBinaryFile, got %v", err)
	}
}

func TestProcessStream_AllowBinary(t *testing.T) {
	e := New()
	e.Config.AllowBinary = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "hello\x00world\nnext line\n"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Binary line skipped, only "next line" should remain
	want := []string{"next line"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	if result[0] != want[0] {
		t.Errorf("got %q, want %q", result[0], want[0])
	}
}

func TestProcessStream_AllowBinaryMultipleLines(t *testing.T) {
	e := New()
	e.Config.AllowBinary = true
	e.Config.BufferSize = 4096
	e.Config.BufferMaxSize = 64 * 1024

	input := "good\nbad\x00\nalso good\n"
	result, err := collectStream(e, input, CHARSEP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"good", "also good"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Errorf("line[%d] = %q, want %q", i, result[i], want[i])
		}
	}
}
