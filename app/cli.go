package app

import (
	"os"

	"BooleanQuery/engine"

	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

var cli struct {
	Query string   `arg:"" help:"Search query string."`
	Files []string `arg:"" optional:"" help:"File paths to search. Reads from stdin if empty."`

	CharSep              string `short:"c" help:"Separator for splitting lines." default:"\n"`
	IgnoreCase           bool   `short:"i" help:"Ignore case sensitivity." default:"false"`
	ExactWord            bool   `short:"w" help:"Match exact words only." default:"false"`
	Wildcard             bool   `short:"W" help:"Enable wildcard matching." default:"false"`
	Recursive            bool   `short:"r" help:"Read all files under each directory recursively." default:"false"`
	DereferenceRecursive bool   `short:"R" help:"Similar to --recursive but follow all symbolic links." default:"false"`

	Grep bool `short:"g" name:"grep-like" help:"Enable grep-compatible plain output (hides header, padding, index, etc)."`

	Color            string `help:"When to use colors: auto, always, never." enum:"auto,always,never" default:"auto"`
	NoIndex          bool   `help:"Disable index numbers in output." default:"false"`
	NoFileHeader     bool   `help:"Disable file header." default:"false"`
	ShowFilePrefix   bool   `short:"f" help:"Show file path in prefix." default:"false"`
	FilesWithMatches bool   `short:"F" help:"Print only names of matched files."`
	Count            bool   `short:"l" help:"Print a count of matched lines per file."`
	NoWarn           bool   `help:"Suppress warning messages."`

	Stream bool `help:"Force sequential processing (single-threaded) and print immediately." default:"false"`

	After   int `short:"A" help:"Print N lines after match."`
	Before  int `short:"B" help:"Print N lines before match."`
	Context int `short:"C" help:"Print N lines before and after match."`

	AllowBinary bool `help:"Process binary files (skip binary lines instead of stopping)." default:"false"`
	MaxBuffer   int  `help:"Max buffer size in KB (max line length)." default:"1024"`
}

func ParseCLI() {
	kong.Parse(&cli)

	if cli.Grep {
		cli.NoIndex = true
		cli.NoFileHeader = true
		cli.ShowFilePrefix = true
	}

	if cli.Context > 0 {
		cli.After = cli.Context
		cli.Before = cli.Context
	}
}

func GetQuery() string {
	return cli.Query
}

func isOutputPiped() bool {
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

func ApplyConfigToEngine(e *engine.Engine) {
	e.Config.CharSep = cli.CharSep
	e.Config.IgnoreCase = cli.IgnoreCase
	e.Config.ExactWord = cli.ExactWord
	e.Config.Wildcard = cli.Wildcard

	switch cli.Color {
	case "always":
		e.Config.NoColor = false
	case "never":
		e.Config.NoColor = true
	default: // "auto"
		e.Config.NoColor = isOutputPiped()
	}

	e.Config.AllowBinary = cli.AllowBinary

	maxBytes := cli.MaxBuffer * 1024
	e.Config.BufferMaxSize = max(4096, maxBytes)
	e.Config.BufferSize = min(64*1024, maxBytes)
}
