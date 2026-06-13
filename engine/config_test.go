package engine

import "testing"

func TestCombineFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags []IntFlag
		want  IntFlag
	}{
		{"no flags", nil, 0},
		{"single flag", []IntFlag{WORDS}, WORDS},
		{"single flag charsep", []IntFlag{CHARSEP}, CHARSEP},
		{"combined", []IntFlag{WORDS, CHARSEP}, WORDS | CHARSEP},
		{"duplicate", []IntFlag{WORDS, WORDS}, WORDS},
		{"all flags", []IntFlag{WORDS, CHARSEP}, WORDS | CHARSEP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CombineFlags(tt.flags...)
			if got != tt.want {
				t.Errorf("CombineFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
