package engine

import "testing"

func newTestEngine(cfg Config, query string) *Engine {
	e := New()
	e.Config = cfg
	if err := e.SetSearchTerm(query); err != nil {
		panic(err)
	}
	e.Classify()
	return e
}

func TestIndexWithQuestionMark(t *testing.T) {
	tests := []struct {
		name string
		s    []byte
		sub  []byte
		want int
	}{
		{"simple match", []byte("hello"), []byte("hello"), 0},
		{"no match", []byte("hello"), []byte("world"), -1},
		{"sub with question matches any", []byte("hxllo"), []byte("h\x00llo"), 0},
		{"sub with question no match wrong length", []byte("hi"), []byte("h\x00i\x00"), -1},
		{"empty sub", []byte("hello"), []byte{}, 0},
		{"empty text", []byte{}, []byte("a"), -1},
		{"question at start", []byte("abc"), []byte("\x00bc"), 0},
		{"question at end", []byte("abc"), []byte("ab\x00"), 0},
		{"multiple questions", []byte("axcye"), []byte("a\x00c\x00e"), 0},
		{"no question byte plain index", []byte("hello world"), []byte("world"), 6},
		{"sub longer than s", []byte("ab"), []byte("abc"), -1},
		{"question matching space", []byte("a b"), []byte("a\x00b"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexWithQuestionMark(tt.s, tt.sub)
			if got != tt.want {
				t.Errorf("indexWithQuestionMark(%q, %q) = %d, want %d", string(tt.s), string(tt.sub), got, tt.want)
			}
		})
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		b    byte
		want bool
	}{
		{'a', true}, {'z', true}, {'A', true}, {'Z', true},
		{'0', true}, {'9', true}, {'_', true},
		{' ', false}, {'.', false}, {'-', false}, {'[', false},
		{'\n', false}, {'\t', false}, {'@', false}, {'#', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.b), func(t *testing.T) {
			got := isWordChar(tt.b)
			if got != tt.want {
				t.Errorf("isWordChar(%q) = %v, want %v", tt.b, got, tt.want)
			}
		})
	}
}

func TestFindTermIndex(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		query       string
		offset      int
		wantStart   int
		wantEnd     int
		wantNoMatch bool
	}{
		{
			name:      "simple match",
			text:      "hello world",
			query:     "hello",
			offset:    0,
			wantStart: 0,
			wantEnd:   5,
		},
		{
			name:      "match with offset",
			text:      "hello hello",
			query:     "hello",
			offset:    6,
			wantStart: 6,
			wantEnd:   11,
		},
		{
			name:        "no match",
			text:        "hello world",
			query:       "xyz",
			offset:      0,
			wantNoMatch: true,
		},
		{
			name:        "offset past end",
			text:        "hello",
			query:       "hello",
			offset:      10,
			wantNoMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{}, tt.query)
			term := e.searchTerm.whiteList[0]
			start, end := e.findTermIndex([]byte(tt.text), term, tt.offset)
			if tt.wantNoMatch {
				if start != -1 {
					t.Errorf("expected no match, got start=%d", start)
				}
				return
			}
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("findTermIndex = (%d, %d), want (%d, %d)", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestFindTermIndex_ExactWord(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		query       string
		wantStart   int
		wantEnd     int
		wantNoMatch bool
	}{
		{
			name:      "exact word match",
			text:      "hello world",
			query:     "world",
			wantStart: 6,
			wantEnd:   11,
		},
		{
			name:        "partial word no match",
			text:        "helloworld",
			query:       "hello",
			wantNoMatch: true,
		},
		{
			name:      "word at start boundary",
			text:      "hello-world",
			query:     "hello",
			wantStart: 0,
			wantEnd:   5,
		},
		{
			name:      "word at end boundary",
			text:      "hello world",
			query:     "world",
			wantStart: 6,
			wantEnd:   11,
		},
		{
			name:        "embedded word no match",
			text:        "abcworlddef",
			query:       "world",
			wantNoMatch: true,
		},
		{
			name:      "skip partial then find exact",
			text:      "helloworld hello",
			query:     "hello",
			wantStart: 11,
			wantEnd:   16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{ExactWord: true}, tt.query)
			term := e.searchTerm.whiteList[0]
			start, end := e.findTermIndex([]byte(tt.text), term, 0)
			if tt.wantNoMatch {
				if start != -1 {
					t.Errorf("expected no match, got start=%d", start)
				}
				return
			}
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("findTermIndex = (%d, %d), want (%d, %d)", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestFindTermIndex_Wildcard(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		query       string
		wantStart   int
		wantEnd     int
		wantNoMatch bool
	}{
		{
			name:      "star prefix",
			text:      "prefix_hello_suffix",
			query:     "*hello",
			wantStart: 0,
			wantEnd:   12,
		},
		{
			name:      "star suffix",
			text:      "prefix_hello",
			query:     "hello*",
			wantStart: 7,
			wantEnd:   12,
		},
		{
			name:      "star both sides",
			text:      "prefix_hello_suffix",
			query:     "*hello*",
			wantStart: 0,
			wantEnd:   19,
		},
		{
			name:      "question mark",
			text:      "heXlo",
			query:     "he?lo",
			wantStart: 0,
			wantEnd:   5,
		},
		{
			name:        "question mark no match",
			text:        "heXXlo",
			query:       "he?lo",
			wantNoMatch: true,
		},
		{
			name:      "multiple segments",
			text:      "start_mid_end",
			query:     "start*end",
			wantStart: 0,
			wantEnd:   13,
		},
		{
			name:      "star match empty",
			text:      "startend",
			query:     "start*end",
			wantStart: 0,
			wantEnd:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{Wildcard: true}, tt.query)
			st := e.searchTerm
			// The wildcard term could be in whiteList or greyList etc.
			var term Term
			if len(st.whiteList) > 0 {
				term = st.whiteList[0]
			}
			start, end := e.findTermIndex([]byte(tt.text), term, 0)
			if tt.wantNoMatch {
				if start != -1 {
					t.Errorf("expected no match, got start=%d", start)
				}
				return
			}
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("findTermIndex = (%d, %d), want (%d, %d)", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestFindWildcardInterval(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		termStr     string
		wantStart   int
		wantEnd     int
		wantNoMatch bool
	}{
		{
			name:      "simple interval",
			text:      "abc_def_ghi",
			termStr:   "abc*ghi",
			wantStart: 0,
			wantEnd:   11,
		},
		{
			name:      "interval with question",
			text:      "abXdefYghi",
			termStr:   "ab?def?ghi",
			wantStart: 0,
			wantEnd:   10,
		},
		{
			name:      "no wildcard segments",
			text:      "hello",
			termStr:   "hello",
			wantStart: 0,
			wantEnd:   5,
		},
		{
			name:        "no match",
			text:        "abc_xyz",
			termStr:     "abc*ghi",
			wantNoMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{Wildcard: true}, tt.termStr)
			term := e.searchTerm.whiteList[0]
			start, end := e.findWildcardInterval([]byte(tt.text), term)
			if tt.wantNoMatch {
				if start != -1 || end != -1 {
					t.Errorf("expected no match, got (%d, %d)", start, end)
				}
				return
			}
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("findWildcardInterval = (%d, %d), want (%d, %d)", start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestContainAnyCheckOnlyBytes(t *testing.T) {
	tests := []struct {
		name  string
		query string
		text  string
		want  int // -1 for no match, >=0 for index
	}{
		{
			name:  "match first",
			query: "~hello ~world",
			text:  "hello there",
			want:  0,
		},
		{
			name:  "match second",
			query: "~hello ~world",
			text:  "the world",
			want:  1,
		},
		{
			name:  "no match",
			query: "~hello ~world",
			text:  "goodbye",
			want:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{}, tt.query)
			got := e.containAnyCheckOnlyBytes(e.searchTerm.greyList, []byte(tt.text))
			if got != tt.want {
				t.Errorf("containAnyCheckOnlyBytes = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestContainAllCheckOnlyBytes(t *testing.T) {
	tests := []struct {
		name  string
		query string
		text  string
		want  bool
	}{
		{
			name:  "both present",
			query: "hello world",
			text:  "hello world",
			want:  true,
		},
		{
			name:  "only one present",
			query: "hello world",
			text:  "hello there",
			want:  false,
		},
		{
			name:  "none present",
			query: "hello world",
			text:  "goodbye",
			want:  false,
		},
		{
			name:  "single term",
			query: "hello",
			text:  "hello world",
			want:  true,
		},
		{
			name:  "empty query",
			query: "",
			text:  "anything",
			want:  true, // no terms to check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{}, tt.query)
			got := e.containAllCheckOnlyBytes(e.searchTerm.whiteList, []byte(tt.text))
			if got != tt.want {
				t.Errorf("containAllCheckOnlyBytes = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainOrderedCheckOnlyBytes(t *testing.T) {
	tests := []struct {
		name  string
		query string
		text  string
		want  bool
	}{
		{
			name:  "correct order",
			query: "^func ^main",
			text:  "func main() {",
			want:  true,
		},
		{
			name:  "wrong order",
			query: "^func ^main",
			text:  "main calls func",
			want:  false,
		},
		{
			name:  "present but interleaved",
			query: "^start ^end",
			text:  "start middle end",
			want:  true,
		},
		{
			name:  "missing second",
			query: "^func ^main",
			text:  "func only",
			want:  false,
		},
		{
			name:  "empty query",
			query: "",
			text:  "anything",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{}, tt.query)
			got := e.containOrderedCheckOnlyBytes(e.searchTerm.orderedList, []byte(tt.text))
			if got != tt.want {
				t.Errorf("containOrderedCheckOnlyBytes = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckOnlyBytes(t *testing.T) {
	tests := []struct {
		name  string
		query string
		cfg   Config
		text  string
		want  bool
	}{
		{
			name:  "AND both match",
			query: "error database",
			text:  "error in database",
			want:  true,
		},
		{
			name:  "AND one missing",
			query: "error database",
			text:  "error only",
			want:  false,
		},
		{
			name:  "term negated by NOT",
			query: "error -timeout",
			text:  "error with timeout",
			want:  false,
		},
		{
			name:  "OR satisfied",
			query: "struct ~json ~xml",
			text:  "struct with json",
			want:  true,
		},
		{
			name:  "OR not satisfied",
			query: "struct ~json ~xml",
			text:  "struct only",
			want:  false,
		},
		{
			name:  "ORDERED correct",
			query: "^func ^main",
			text:  "func main()",
			want:  true,
		},
		{
			name:  "ORDERED wrong",
			query: "^func ^main",
			text:  "main func()",
			want:  false,
		},
		{
			name:  "full boolean: AND+NOT+OR",
			query: "error -timeout ~critical ~fatal",
			text:  "error critical in database",
			want:  true,
		},
		{
			name:  "full boolean: blacklist blocks",
			query: "error -timeout ~critical ~fatal",
			text:  "error timeout critical",
			want:  false,
		},
		{
			name:  "AND+ORDERED combined",
			query: "start ^middle ^end",
			text:  "start middle end",
			want:  true,
		},
		{
			name:  "case insensitive match",
			query: "Error Database",
			cfg:   Config{IgnoreCase: true},
			text:  "error in database",
			want:  true,
		},
		{
			name:  "exact word match",
			query: "cat",
			cfg:   Config{ExactWord: true},
			text:  "cat dog",
			want:  true,
		},
		{
			name:  "exact word no partial",
			query: "cat",
			cfg:   Config{ExactWord: true},
			text:  "caterpillar",
			want:  false,
		},
		{
			name:  "empty query matches everything",
			query: "",
			text:  "anything",
			want:  true,
		},
		{
			name:  "NOT alone blocks",
			query: "-bad",
			text:  "good content",
			want:  true,
		},
		{
			name:  "NOT alone rejects",
			query: "-bad",
			text:  "bad content",
			want:  false,
		},
		{
			name:  "OR alone passes",
			query: "~maybe",
			text:  "maybe this",
			want:  true,
		},
		{
			name:  "OR alone fails",
			query: "~maybe",
			text:  "not that",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(tt.cfg, tt.query)
			got := e.CheckOnlyBytes([]byte(tt.text))
			if got != tt.want {
				t.Errorf("CheckOnlyBytes(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		text         string
		wantNil      bool
		wantEmpty    bool
		wantMatchCnt int
	}{
		{
			name:         "simple match",
			query:        "hello",
			text:         "say hello world",
			wantNil:      false,
			wantMatchCnt: 1,
		},
		{
			name:    "no match",
			query:   "hello",
			text:    "goodbye world",
			wantNil: true,
		},
		{
			name:      "empty query returns empty matches",
			query:     "",
			text:      "anything",
			wantEmpty: true,
		},
		{
			name:         "blacklist ignored by Search",
			query:        "hello -world",
			text:         "hello world",
			wantNil:      false,
			wantMatchCnt: 1,
		},
		{
			name:    "AND missing returns nil",
			query:   "hello world",
			text:    "hello there",
			wantNil: true,
		},
		{
			name:         "multiple matches same term",
			query:        "hello",
			text:         "hello hello hello",
			wantNil:      false,
			wantMatchCnt: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine(Config{}, tt.query)
			matches := e.Search(tt.text)
			if tt.wantNil {
				if matches != nil {
					t.Errorf("expected nil matches, got %v", matches)
				}
				return
			}
			if tt.wantEmpty {
				if matches == nil || len(matches) != 0 {
					t.Errorf("expected empty matches, got %v", matches)
				}
				return
			}
			if len(matches) != tt.wantMatchCnt {
				t.Errorf("got %d matches, want %d; matches=%v", len(matches), tt.wantMatchCnt, matches)
			}
		})
	}
}

func TestSearch_MatchPositions(t *testing.T) {
	e := newTestEngine(Config{}, "hello")
	matches := e.Search("say hello world")
	if matches == nil || len(matches) != 1 {
		t.Fatalf("expected 1 match, got %v", matches)
	}
	if matches[0][0] != 4 || matches[0][1] != 9 {
		t.Errorf("match position = (%d,%d), want (4,9)", matches[0][0], matches[0][1])
	}
}

func TestSearch_MultipleTerms(t *testing.T) {
	e := newTestEngine(Config{}, "hello world")
	matches := e.Search("hello beautiful world")
	if matches == nil {
		t.Fatal("expected matches, got nil")
	}
	// Should have matches for both hello and world sorted by position
	if len(matches) < 2 {
		t.Errorf("expected at least 2 matches, got %d", len(matches))
	}
	for i := 1; i < len(matches); i++ {
		if matches[i][0] < matches[i-1][0] {
			t.Errorf("matches not sorted by position: %v", matches)
			break
		}
	}
}

func TestSearch_OrderedMatchPositions(t *testing.T) {
	e := newTestEngine(Config{}, "^func ^main")
	matches := e.Search("func main()")
	if matches == nil || len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %v", matches)
	}
	// func starts at 0
	if matches[0][0] != 0 {
		t.Errorf("first match start = %d, want 0", matches[0][0])
	}
	// main should come after func
	if matches[1][0] <= matches[0][0] {
		t.Errorf("ordered matches wrong: %v", matches)
	}
}

func TestSearch_SortedByPosition(t *testing.T) {
	// Using OR to get matches in potentially wrong order
	e := newTestEngine(Config{}, "~world ~hello ~end")
	matches := e.Search("hello world end")
	if matches == nil {
		t.Fatal("expected matches")
	}
	for i := 1; i < len(matches); i++ {
		if matches[i][0] < matches[i-1][0] {
			t.Errorf("matches not sorted: %v", matches)
			break
		}
	}
}

func TestContainAny(t *testing.T) {
	e := newTestEngine(Config{}, "~cat ~dog ~bird")
	matches := e.containAny(e.searchTerm.greyList, "I have a cat and a dog")
	if matches == nil {
		t.Fatal("expected matches")
	}
	// Should have at least one match
	if len(matches) < 1 {
		t.Errorf("expected at least 1 match, got %d", len(matches))
	}
}

func TestContainAll(t *testing.T) {
	e := newTestEngine(Config{}, "cat dog")
	matches := e.containAll(e.searchTerm.whiteList, "cat and dog")
	if matches == nil {
		t.Fatal("expected matches for containAll")
	}
}

func TestContainOrdered(t *testing.T) {
	e := newTestEngine(Config{}, "^start ^end")
	matches := e.containOrdered(e.searchTerm.orderedList, "start middle end")
	if matches == nil {
		t.Fatal("expected matches for containOrdered")
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0][0] > matches[1][0] {
		t.Errorf("ordered matches out of order: %v", matches)
	}
}

func TestContainAll_NoMatch(t *testing.T) {
	e := newTestEngine(Config{}, "cat dog")
	matches := e.containAll(e.searchTerm.whiteList, "cat only")
	if matches != nil {
		t.Errorf("expected nil, got %v", matches)
	}
}

func TestContainOrdered_NoMatch(t *testing.T) {
	e := newTestEngine(Config{}, "^start ^end")
	matches := e.containOrdered(e.searchTerm.orderedList, "end then start")
	if matches != nil {
		t.Errorf("expected nil, got %v", matches)
	}
}
