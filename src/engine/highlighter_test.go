package engine

import (
	"bytes"
	"testing"
)

func TestMergeIntervals(t *testing.T) {
	tests := []struct {
		name  string
		input [][2]int
		want  [][2]int
	}{
		{
			name:  "empty",
			input: nil,
			want:  nil,
		},
		{
			name:  "single interval",
			input: [][2]int{{0, 5}},
			want:  [][2]int{{0, 5}},
		},
		{
			name:  "non overlapping sorted",
			input: [][2]int{{0, 2}, {5, 8}},
			want:  [][2]int{{0, 2}, {5, 8}},
		},
		{
			name:  "overlapping merge",
			input: [][2]int{{0, 5}, {3, 8}},
			want:  [][2]int{{0, 8}},
		},
		{
			name:  "adjacent merge",
			input: [][2]int{{0, 5}, {5, 10}},
			want:  [][2]int{{0, 10}},
		},
		{
			name:  "inner interval fully contained",
			input: [][2]int{{0, 10}, {3, 5}},
			want:  [][2]int{{0, 10}},
		},
		{
			name:  "multiple merges",
			input: [][2]int{{0, 3}, {2, 5}, {10, 15}, {14, 20}},
			want:  [][2]int{{0, 5}, {10, 20}},
		},
		{
			name:  "identical intervals",
			input: [][2]int{{0, 5}, {0, 5}},
			want:  [][2]int{{0, 5}},
		},
		{
			name:  "three intervals merging to one",
			input: [][2]int{{0, 3}, {3, 7}, {6, 10}},
			want:  [][2]int{{0, 10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeIntervals(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("mergeIntervals() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("result[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestHighlightTo(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		text    string
		matches [][2]int
		want    string
	}{
		{
			name:    "no color mode",
			noColor: true,
			text:    "hello world",
			matches: [][2]int{{0, 5}},
			want:    "hello world",
		},
		{
			name:    "no matches",
			noColor: false,
			text:    "hello world",
			matches: nil,
			want:    "hello world",
		},
		{
			name:    "single match",
			noColor: false,
			text:    "hello world",
			matches: [][2]int{{0, 5}},
			want:    ColorRed + "hello" + ColorReset + " world",
		},
		{
			name:    "single match in middle",
			noColor: false,
			text:    "say hello world",
			matches: [][2]int{{4, 9}},
			want:    "say " + ColorRed + "hello" + ColorReset + " world",
		},
		{
			name:    "multiple matches",
			noColor: false,
			text:    "hello world hello",
			matches: [][2]int{{0, 5}, {12, 17}},
			want:    ColorRed + "hello" + ColorReset + " world " + ColorRed + "hello" + ColorReset,
		},
		{
			name:    "adjacent matches merged",
			noColor: false,
			text:    "hello world",
			matches: [][2]int{{0, 5}, {6, 11}},
			want:    ColorRed + "hello" + ColorReset + " " + ColorRed + "world" + ColorReset,
		},
		{
			name:    "overlapping matches merged as single highlight",
			noColor: false,
			text:    "hello world",
			matches: [][2]int{{0, 5}, {5, 11}},
			want:    ColorRed + "hello world" + ColorReset,
		},
		{
			name:    "match at end",
			noColor: false,
			text:    "hello world",
			matches: [][2]int{{6, 11}},
			want:    "hello " + ColorRed + "world" + ColorReset,
		},
		{
			name:    "overlapping matches merged",
			noColor: false,
			text:    "abcde",
			matches: [][2]int{{0, 3}, {2, 5}},
			want:    ColorRed + "abcde" + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.Config = Config{NoColor: tt.noColor}
			var buf bytes.Buffer
			e.HighlightTo(&buf, tt.text, tt.matches)
			got := buf.String()
			if got != tt.want {
				t.Errorf("HighlightTo = %q, want %q", got, tt.want)
			}
		})
	}
}
