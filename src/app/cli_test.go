package app

import (
	"os"
	"testing"

	"BooleanQuery/src/engine"
)

func setCLIDefaults() {
	cli.CharSep = "\n"
	cli.IgnoreCase = false
	cli.ExactWord = false
	cli.Wildcard = false
	cli.Color = "auto"
	cli.AllowBinary = false
	cli.MaxBuffer = 1024
	cli.NoIndex = false
	cli.NoFileHeader = false
	cli.ShowFilePrefix = false
	cli.Grep = false
	cli.Count = false
	cli.FilesWithMatches = false
	cli.Context = 0
	cli.Before = 0
	cli.After = 0
}

func TestApplyConfigToEngine(t *testing.T) {
	tests := []struct {
		name  string
		fn    func()
		check func(t *testing.T, e *engine.Engine)
	}{
		{
			name: "basic config mapping",
			fn: func() {
				cli.CharSep = ","
				cli.IgnoreCase = true
				cli.ExactWord = true
				cli.Wildcard = true
				cli.Color = "never"
				cli.AllowBinary = true
				cli.MaxBuffer = 256
			},
			check: func(t *testing.T, e *engine.Engine) {
				if e.Config.CharSep != "," {
					t.Errorf("CharSep = %q, want %q", e.Config.CharSep, ",")
				}
				if !e.Config.IgnoreCase {
					t.Error("IgnoreCase should be true")
				}
				if !e.Config.ExactWord {
					t.Error("ExactWord should be true")
				}
				if !e.Config.Wildcard {
					t.Error("Wildcard should be true")
				}
				if !e.Config.NoColor {
					t.Error("NoColor should be true for color=never")
				}
				if !e.Config.AllowBinary {
					t.Error("AllowBinary should be true")
				}
				if e.Config.BufferMaxSize < 4096 {
					t.Errorf("BufferMaxSize too small: %d", e.Config.BufferMaxSize)
				}
			},
		},
		{
			name: "color always",
			fn: func() {
				cli.Color = "always"
			},
			check: func(t *testing.T, e *engine.Engine) {
				if e.Config.NoColor {
					t.Error("NoColor should be false for color=always")
				}
			},
		},
		{
			name: "buffer size minimum enforced",
			fn: func() {
				cli.MaxBuffer = 1
			},
			check: func(t *testing.T, e *engine.Engine) {
				if e.Config.BufferMaxSize < 4096 {
					t.Errorf("BufferMaxSize = %d, want at least 4096", e.Config.BufferMaxSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := cli
			defer func() { cli = old }()
			setCLIDefaults()

			tt.fn()
			e := engine.New()
			ApplyConfigToEngine(e)
			tt.check(t, e)
		})
	}
}

func TestParseCLI_Grep(t *testing.T) {
	old := cli
	oldArgs := os.Args
	defer func() {
		cli = old
		os.Args = oldArgs
	}()
	os.Args = []string{"bq", "-g", "test"}

	ParseCLI()

	if !cli.NoIndex {
		t.Error("Grep should set NoIndex")
	}
	if !cli.NoFileHeader {
		t.Error("Grep should set NoFileHeader")
	}
	if !cli.ShowFilePrefix {
		t.Error("Grep should set ShowFilePrefix")
	}
}

func TestParseCLI_Context(t *testing.T) {
	old := cli
	oldArgs := os.Args
	defer func() {
		cli = old
		os.Args = oldArgs
	}()
	os.Args = []string{"bq", "-C", "5", "test"}

	ParseCLI()

	if cli.Before != 5 {
		t.Errorf("Before = %d, want 5", cli.Before)
	}
	if cli.After != 5 {
		t.Errorf("After = %d, want 5", cli.After)
	}
}

func TestGetQueries(t *testing.T) {
	old := cli
	defer func() { cli = old }()
	setCLIDefaults()

	cli.Query = "test query"
	got := GetQueries()
	if len(got) != 1 || got[0] != "test query" {
		t.Errorf("GetQueries() = %v, want [test query]", got)
	}
}

func TestGetQueries_Regexp(t *testing.T) {
	old := cli
	defer func() { cli = old }()
	setCLIDefaults()

	cli.Patterns = []string{"error", "warning"}
	got := GetQueries()
	if len(got) != 2 || got[0] != "error" || got[1] != "warning" {
		t.Errorf("GetQueries() = %v, want [error warning]", got)
	}
}
