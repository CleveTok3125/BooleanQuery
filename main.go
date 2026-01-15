package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

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

	After   int `short:"A" help:"Print N lines after match."`
	Before  int `short:"B" help:"Print N lines before match."`
	Context int `short:"C" help:"Print N lines before and after match."`
}

type lineInfo struct {
	index int
	text  string
}

var errorLock sync.Mutex

func printLine(w io.Writer, e *engine.Engine, index int, part string) {
	cIndex := engine.ColorBrightBlack
	cReset := engine.ColorReset

	if e.Config.NoColor {
		cIndex = ""
		cReset = ""
	}

	if e.Config.NoIndex {
		fmt.Fprintf(w, "%s\n", e.Highlight(part))
	} else {
		fmt.Fprintf(w, "%s%d:%s %s\n", cIndex, index, cReset, e.Highlight(part))
	}
}

func processInput(w io.Writer, e *engine.Engine, reader io.Reader) {
	combinedFlag := engine.CombineFlags(engine.CHARSEP)

	var beforeBuffer []lineInfo
	linesToPrintAfter := 0
	lastPrintedIndex := -1

	iter := e.ProcessStream(reader, combinedFlag)

	iter(func(index int, part string) bool {
		isMatch := e.IsMatch(part)

		if isMatch {
			for _, item := range beforeBuffer {
				if item.index > lastPrintedIndex {
					printLine(w, e, item.index, item.text)
					lastPrintedIndex = item.index
				}
			}

			printLine(w, e, index, part)
			lastPrintedIndex = index

			beforeBuffer = nil
			linesToPrintAfter = cli.After

		} else {
			if linesToPrintAfter > 0 {
				if index > lastPrintedIndex {
					printLine(w, e, index, part)
					lastPrintedIndex = index
				}
				linesToPrintAfter--
			}

			if cli.Before > 0 {
				beforeBuffer = append(beforeBuffer, lineInfo{index, part})

				if len(beforeBuffer) > cli.Before {
					beforeBuffer = beforeBuffer[1:]
				}
			}
		}
		return true
	})
}

func processFile(e *engine.Engine, path string) (*bytes.Buffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var contentBuf bytes.Buffer
	processInput(&contentBuf, e, f)

	var finalBuf bytes.Buffer

	if contentBuf.Len() > 0 {
		if len(cli.Files) > 1 && !e.Config.NoFilename {
			fmt.Fprintf(&finalBuf, "--- File: %s ---\n", path)
		}
		contentBuf.WriteTo(&finalBuf)
	}

	return &finalBuf, nil
}

func main() {
	kong.Parse(&cli)

	if cli.Context > 0 {
		cli.After = cli.Context
		cli.Before = cli.Context
	}

	e := engine.New()
	e.Config.CharSep = cli.CharSep
	e.Config.IgnoreCase = cli.IgnoreCase
	e.Config.ExactWord = cli.ExactWord
	e.Config.NoColor = cli.NoColor
	e.Config.NoIndex = cli.NoIndex
	e.Config.NoFilename = cli.NoFilename

	e.Config.BufferMaxSize = 1024 * 1024
	e.Config.BufferSize = 64 * 1024

	e.SetSearchTerm(cli.Query)
	e.Classify()
	e.PrepareHighlight()

	if len(cli.Files) == 0 {
		processInput(os.Stdout, e, os.Stdin)
	} else {
		numCPU := runtime.NumCPU()
		sem := make(chan struct{}, numCPU)
		var wg sync.WaitGroup

		results := make([]*bytes.Buffer, len(cli.Files))

		for i, path := range cli.Files {
			wg.Add(1)

			go func(idx int, p string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				buf, err := processFile(e, p)
				if err != nil {
					errorLock.Lock()
					fmt.Fprintf(os.Stderr, "Error: %s\n", err)
					errorLock.Unlock()
					return
				}

				results[idx] = buf
			}(i, path)
		}

		wg.Wait()

		for _, buf := range results {
			if buf != nil && buf.Len() > 0 {
				buf.WriteTo(os.Stdout)
			}
		}
	}
}
