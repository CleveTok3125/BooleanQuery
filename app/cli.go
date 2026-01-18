package app

import (
	"BooleanQuery/engine"

	"github.com/alecthomas/kong"
)

var cli struct {
	Query string   `arg:"" help:"Search query string"`
	Files []string `arg:"" optional:"" type:"path" help:"File paths to search. Reads from stdin if empty."`

	// Config flags
	CharSep    string `help:"Separator for splitting lines."`
	IgnoreCase bool   `short:"i" help:"Ignore case sensitivity." default:"false"`
	ExactWord  bool   `help:"Match exact words only." default:"false"`
	NoColor    bool   `help:"Disable colored output." default:"false"`
	NoIndex    bool   `help:"Disable index numbers in output." default:"false"`
	NoFilename bool   `help:"Disable file name prefix." default:"false"`

	Stream bool `help:"Force sequential processing (single-threaded) and print immediately." default:"false"`

	After   int `short:"A" help:"Print N lines after match."`
	Before  int `short:"B" help:"Print N lines before match."`
	Context int `short:"C" help:"Print N lines before and after match."`
}

func ParseCLI() {
	kong.Parse(&cli)

	if cli.Context > 0 {
		cli.After = cli.Context
		cli.Before = cli.Context
	}
}

func GetQuery() string {
	return cli.Query
}

func ApplyConfigToEngine(e *engine.Engine) {
	e.Config.CharSep = cli.CharSep
	e.Config.IgnoreCase = cli.IgnoreCase
	e.Config.ExactWord = cli.ExactWord
	e.Config.NoColor = cli.NoColor
	e.Config.NoIndex = cli.NoIndex
	e.Config.NoFilename = cli.NoFilename
}
