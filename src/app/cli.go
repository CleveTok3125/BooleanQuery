package app

import (
	"fmt"
	"os"
	"strings"

	"BooleanQuery/src/engine"

	"github.com/alecthomas/kong"
	"golang.org/x/term"
)

var cli struct {
	Query    string   `arg:"" optional:"" help:"Search query string."`
	Patterns []string `short:"e" name:"pattern" help:"Pattern (can be specified multiple times)."`
	Files    []string `arg:"" optional:"" help:"File paths to search. Reads from stdin if empty."`

	CharSep              string `short:"c" help:"Separator for splitting lines." default:"\n"`
	IgnoreCase           bool   `short:"i" help:"Ignore case sensitivity." default:"false"`
	ExactWord            bool   `short:"w" help:"Match exact words only." default:"false"`
	Wildcard             bool   `short:"W" help:"Enable wildcard matching." default:"false"`
	Recursive            bool   `short:"r" help:"Read all files under each directory recursively." default:"false"`
	DereferenceRecursive bool   `short:"R" help:"Similar to --recursive but follow all symbolic links." default:"false"`

	Grep bool `short:"g" name:"grep-like" help:"Enable grep-compatible plain output (hides header, padding, index, etc)."`

	Color            string   `help:"When to use colors: auto, always, never." enum:"auto,always,never" default:"auto"`
	NoIndex          bool     `help:"Disable index numbers in output." default:"false"`
	NoFileHeader     bool     `help:"Disable file header." default:"false"`
	ShowFilePrefix   bool     `short:"f" help:"Show file path in prefix." default:"false"`
	FilesWithMatches bool     `short:"F" help:"Print only names of matched files."`
	Count            bool     `short:"l" help:"Print a count of matched lines per file."`
	Info             []string `short:"I" help:"Show info messages by category (comma-separated). Available: binary, buffer, dir, all."`

	Stream bool `help:"Force sequential processing (single-threaded) and print immediately." default:"false"`

	After   int `short:"A" help:"Print N lines after match."`
	Before  int `short:"B" help:"Print N lines before match."`
	Context int `short:"C" help:"Print N lines before and after match."`

	AllowBinary bool `help:"Process binary files (skip binary lines instead of stopping)." default:"false"`
	MaxBuffer   int  `help:"Max buffer size in KB (max line length)." default:"1024"`
}

var validInfoCategories = map[string]bool{
	"binary": true,
	"buffer": true,
	"dir":    true,
}

func isInfo(category string) bool {
	for _, item := range cli.Info {
		for _, f := range strings.Split(item, ",") {
			f = strings.TrimSpace(f)
			if f == "all" {
				return validInfoCategories[category]
			}
			if f == category {
				return true
			}
		}
	}
	return false
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

	if len(cli.Patterns) > 0 && cli.Query != "" {
		cli.Files = append([]string{cli.Query}, cli.Files...)
		cli.Query = ""
	}

	for i, p := range cli.Patterns {
		if len(p) > 0 && p[0] == '=' {
			cli.Patterns[i] = p[1:]
		}
	}

	for _, item := range cli.Info {
		for _, f := range strings.Split(item, ",") {
			f = strings.TrimSpace(f)
			if f == "all" {
				continue
			}
			if !validInfoCategories[f] {
				fmt.Fprintf(os.Stderr, "bq: invalid info category %q. Available: binary, buffer, dir, all\n", f)
				os.Exit(2)
			}
		}
	}
}

func GetQueries() []string {
	if len(cli.Patterns) > 0 {
		return cli.Patterns
	}
	if cli.Query == "" {
		fmt.Fprintf(os.Stderr, "bq: no search pattern specified\n")
		os.Exit(2)
	}
	return []string{cli.Query}
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
