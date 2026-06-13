package engine

import (
	"bytes"
	"testing"
)

func TestParseWildcard(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantHasWild bool
		wantSegs    [][]byte
	}{
		{
			name:        "no wildcard",
			text:        "hello",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("hello")},
		},
		{
			name:        "star wildcard",
			text:        "he*lo",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("he"), []byte("lo")},
		},
		{
			name:        "question wildcard",
			text:        "he?lo",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("he\x00lo")},
		},
		{
			name:        "multiple wildcards",
			text:        "a*b*c",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		},
		{
			name:        "leading star",
			text:        "*abc",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte(""), []byte("abc")},
		},
		{
			name:        "trailing star",
			text:        "abc*",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("abc"), []byte("")},
		},
		{
			name:        "only star",
			text:        "*",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte(""), []byte("")},
		},
		{
			name:        "only question",
			text:        "?",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("\x00")},
		},
		{
			name:        "escape star",
			text:        "he@*lo",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("he*lo")},
		},
		{
			name:        "escape question",
			text:        "he@?lo",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("he?lo")},
		},
		{
			name:        "escape escape char",
			text:        "he@@lo",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("he@lo")},
		},
		{
			name:        "escaped star followed by wild star",
			text:        "@*abc*",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("*abc"), []byte("")},
		},
		{
			name:        "escape at end no next char",
			text:        "abc@",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("abc@")},
		},
		{
			name:        "empty string",
			text:        "",
			wantHasWild: false,
			wantSegs:    [][]byte{[]byte("")},
		},
		{
			name:        "multiple questions",
			text:        "a?b?c",
			wantHasWild: true,
			wantSegs:    [][]byte{[]byte("a\x00b\x00c")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs, hasWild := parseWildcard(tt.text)
			if hasWild != tt.wantHasWild {
				t.Errorf("hasWild = %v, want %v", hasWild, tt.wantHasWild)
			}
			if len(segs) != len(tt.wantSegs) {
				t.Fatalf("len(segs) = %d, want %d; segs = %v", len(segs), len(tt.wantSegs), segs)
			}
			for i := range segs {
				if !bytes.Equal(segs[i], tt.wantSegs[i]) {
					t.Errorf("segs[%d] = %v, want %v", i, segs[i], tt.wantSegs[i])
				}
			}
		})
	}
}

func TestCreateTerm(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		wildcardMode bool
		wantBytes    string
		wantHasWild  bool
		wantSegsLen  int
	}{
		{
			name:         "no wildcard mode",
			text:         "hello",
			wildcardMode: false,
			wantBytes:    "hello",
			wantHasWild:  false,
			wantSegsLen:  0,
		},
		{
			name:         "wildcard mode no wildcards",
			text:         "hello",
			wildcardMode: true,
			wantBytes:    "hello",
			wantHasWild:  false,
			wantSegsLen:  1,
		},
		{
			name:         "wildcard mode with star",
			text:         "he*lo",
			wildcardMode: true,
			wantBytes:    "he*lo",
			wantHasWild:  true,
			wantSegsLen:  2,
		},
		{
			name:         "wildcard mode with question",
			text:         "he?lo",
			wildcardMode: true,
			wantBytes:    "he?lo",
			wantHasWild:  true,
			wantSegsLen:  1,
		},
		{
			name:         "empty string no wildcard mode",
			text:         "",
			wildcardMode: false,
			wantBytes:    "",
			wantHasWild:  false,
			wantSegsLen:  0,
		},
		{
			name:         "empty string wildcard mode",
			text:         "",
			wildcardMode: true,
			wantBytes:    "",
			wantHasWild:  false,
			wantSegsLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := createTerm(tt.text, tt.wildcardMode)
			if string(term.Bytes) != tt.wantBytes {
				t.Errorf("term.Bytes = %q, want %q", string(term.Bytes), tt.wantBytes)
			}
			if term.HasWildcard != tt.wantHasWild {
				t.Errorf("term.HasWildcard = %v, want %v", term.HasWildcard, tt.wantHasWild)
			}
			if len(term.Segments) != tt.wantSegsLen {
				t.Errorf("len(term.Segments) = %d, want %d", len(term.Segments), tt.wantSegsLen)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		cfg     Config
		wantAnd int
		wantNot int
		wantOr  int
		wantOrd int
	}{
		{
			name:    "default AND terms",
			query:   "error database",
			cfg:     Config{},
			wantAnd: 2,
		},
		{
			name:    "explicit AND plus NOT",
			query:   "+error -timeout",
			cfg:     Config{},
			wantAnd: 1,
			wantNot: 1,
		},
		{
			name:    "OR terms",
			query:   "struct ~json ~xml",
			cfg:     Config{},
			wantAnd: 1,
			wantOr:  2,
		},
		{
			name:    "ORDERED terms",
			query:   "^func ^main",
			cfg:     Config{},
			wantOrd: 2,
		},
		{
			name:    "mixed all types",
			query:   "error -timeout ~warning ~info ^func ^main",
			cfg:     Config{},
			wantAnd: 1,
			wantNot: 1,
			wantOr:  2,
			wantOrd: 2,
		},
		{
			name:    "case insensitive",
			query:   "Error Timeout",
			cfg:     Config{IgnoreCase: true},
			wantAnd: 2,
		},
		{
			name:    "no terms",
			query:   "",
			cfg:     Config{},
			wantAnd: 0,
		},
		{
			name:    "double prefix for literal",
			query:   "++example +-dash",
			cfg:     Config{},
			wantAnd: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Config = tt.cfg
			if err := e.SetSearchTerm(tt.query); err != nil {
				t.Fatalf("SetSearchTerm failed: %v", err)
			}
			e.Classify()

			if len(e.searchTerms[0].whiteList) != tt.wantAnd {
				t.Errorf("whitelist len = %d, want %d", len(e.searchTerms[0].whiteList), tt.wantAnd)
			}
			if len(e.searchTerms[0].blackList) != tt.wantNot {
				t.Errorf("blacklist len = %d, want %d", len(e.searchTerms[0].blackList), tt.wantNot)
			}
			if len(e.searchTerms[0].greyList) != tt.wantOr {
				t.Errorf("greylist len = %d, want %d", len(e.searchTerms[0].greyList), tt.wantOr)
			}
			if len(e.searchTerms[0].orderedList) != tt.wantOrd {
				t.Errorf("orderedlist len = %d, want %d", len(e.searchTerms[0].orderedList), tt.wantOrd)
			}
		})
	}
}

func TestClassify_CaseInsensitive(t *testing.T) {
	e := New()
	e.Config = Config{IgnoreCase: true}
	if err := e.SetSearchTerm("ERROR -TIMEOUT ~WARNING"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	if len(e.searchTerms[0].whiteList) != 1 {
		t.Fatalf("expected 1 whitelist term, got %d", len(e.searchTerms[0].whiteList))
	}
	if string(e.searchTerms[0].whiteList[0].Bytes) != "error" {
		t.Errorf("whitelist term = %q, want %q", string(e.searchTerms[0].whiteList[0].Bytes), "error")
	}

	if len(e.searchTerms[0].blackList) != 1 {
		t.Fatalf("expected 1 blacklist term, got %d", len(e.searchTerms[0].blackList))
	}
	if string(e.searchTerms[0].blackList[0].Bytes) != "timeout" {
		t.Errorf("blacklist term = %q, want %q", string(e.searchTerms[0].blackList[0].Bytes), "timeout")
	}

	if len(e.searchTerms[0].greyList) != 1 {
		t.Fatalf("expected 1 greylist term, got %d", len(e.searchTerms[0].greyList))
	}
	if string(e.searchTerms[0].greyList[0].Bytes) != "warning" {
		t.Errorf("greylist term = %q, want %q", string(e.searchTerms[0].greyList[0].Bytes), "warning")
	}
}

func TestClassify_WildcardMode(t *testing.T) {
	e := New()
	e.Config = Config{Wildcard: true}
	if err := e.SetSearchTerm("+he*lo -good?bye"); err != nil {
		t.Fatal(err)
	}
	e.Classify()

	if len(e.searchTerms[0].whiteList) != 1 {
		t.Fatalf("expected 1 whitelist term, got %d", len(e.searchTerms[0].whiteList))
	}
	if !e.searchTerms[0].whiteList[0].HasWildcard {
		t.Error("expected whitelist term to have wildcard")
	}

	if len(e.searchTerms[0].blackList) != 1 {
		t.Fatalf("expected 1 blacklist term, got %d", len(e.searchTerms[0].blackList))
	}
	if !e.searchTerms[0].blackList[0].HasWildcard {
		t.Error("expected blacklist term to have wildcard")
	}
}

func TestSetSearchTerm_NoErrorForUnclosedQuote(t *testing.T) {
	e := New()
	err := e.SetSearchTerm("error 'unclosed")
	if err != nil {
		t.Fatalf("SetSearchTerm returned error: %v", err)
	}

	e.Classify()
	if len(e.GetSearchTerm().whiteList) != 2 {
		t.Fatalf("expected 2 whitelist terms (error, 'unclosed), got %d", len(e.GetSearchTerm().whiteList))
	}
	if string(e.GetSearchTerm().whiteList[0].Bytes) != "error" {
		t.Errorf("first term = %q, want %q", string(e.GetSearchTerm().whiteList[0].Bytes), "error")
	}
	if string(e.GetSearchTerm().whiteList[1].Bytes) != "'unclosed" {
		t.Errorf("second term = %q, want %q", string(e.GetSearchTerm().whiteList[1].Bytes), "'unclosed")
	}
}

func TestSetSearchTerm_LiteralQuote(t *testing.T) {
	e := New()
	err := e.SetSearchTerm("'")
	if err != nil {
		t.Fatalf("SetSearchTerm returned error: %v", err)
	}

	e.Classify()
	if len(e.GetSearchTerm().whiteList) != 1 {
		t.Fatalf("expected 1 term, got %d", len(e.GetSearchTerm().whiteList))
	}
	if string(e.GetSearchTerm().whiteList[0].Bytes) != "'" {
		t.Errorf("first term = %q, want %q", string(e.GetSearchTerm().whiteList[0].Bytes), "'")
	}
}

func TestGetSearchTerm(t *testing.T) {
	e := New()
	e.SetSearchTerm("hello world")
	e.Classify()

	st := e.GetSearchTerm()
	if len(st.whiteList) != 2 {
		t.Errorf("expected 2 whitelist terms, got %d", len(st.whiteList))
	}
	if string(st.whiteList[0].Bytes) != "hello" {
		t.Errorf("first term = %q, want %q", string(st.whiteList[0].Bytes), "hello")
	}
	if string(st.whiteList[1].Bytes) != "world" {
		t.Errorf("second term = %q, want %q", string(st.whiteList[1].Bytes), "world")
	}
}
