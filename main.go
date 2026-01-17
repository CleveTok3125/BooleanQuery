package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

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

type lineInfo struct {
	index int
	text  string
}

var errorLock sync.Mutex

func printLine(w io.Writer, e *engine.Engine, index int, matches [][2]int, part string) {
	cIndex := engine.ColorBrightBlack
	cReset := engine.ColorReset

	if e.Config.NoColor {
		cIndex = ""
		cReset = ""
	}

	prefix := ""
	if !e.Config.NoIndex {
		colDisplay := -1
		if len(matches) > 0 {
			colDisplay = matches[0][0]
		}

		if colDisplay != -1 {
			prefix = fmt.Sprintf("%s%d:%d:%s ", cIndex, index+1, colDisplay+1, cReset)
		} else {
			prefix = fmt.Sprintf("%s%d:%s ", cIndex, index, cReset)
		}
	}

	printText := e.Highlight(part, matches)

	fmt.Fprintf(w, "%s%s\n", prefix, printText)
}

func processInput(w io.Writer, e *engine.Engine, reader io.Reader, displayName string) {
	combinedFlag := engine.CombineFlags(engine.CHARSEP)

	var beforeBuffer []lineInfo
	linesToPrintAfter := 0
	lastPrintedIndex := -1

	headerPrinted := false

	if !headerPrinted && displayName != "" {
		fmt.Fprintf(w, "--- File: %s ---\n", displayName)
		headerPrinted = true
	}

	iter := e.ProcessStream(reader, combinedFlag)

	iter(func(index int, part string) bool {
		matches := e.Search(part)
		matched := matches != nil

		if matched {
			for _, item := range beforeBuffer {
				if item.index > lastPrintedIndex {
					printLine(w, e, item.index, nil, item.text)
					lastPrintedIndex = item.index
				}
			}

			printLine(w, e, index, matches, part)
			lastPrintedIndex = index

			beforeBuffer = nil
			linesToPrintAfter = cli.After

		} else {
			if linesToPrintAfter > 0 {
				if index > lastPrintedIndex {
					printLine(w, e, index, nil, part)
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

func getDisplayName(path string, e *engine.Engine) string {
	if len(cli.Files) > 1 && !e.Config.NoFilename {
		return path
	}
	return ""
}

func processFile(w io.Writer, e *engine.Engine, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	displayName := getDisplayName(path, e)
	processInput(w, e, f, displayName)
	return nil
}

func processFileToTemp(e *engine.Engine, path string) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer f.Close()

	tmpFile, err := os.CreateTemp("", "bq-buffer-*")
	if err != nil {
		return "", false, err
	}
	defer tmpFile.Close()

	bufWriter := bufio.NewWriter(tmpFile)

	displayName := getDisplayName(path, e)
	processInput(bufWriter, e, f, displayName)

	bufWriter.Flush()

	stat, _ := tmpFile.Stat()
	hasContent := stat.Size() > 0

	return tmpFile.Name(), hasContent, nil
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	var tempFiles []string
	var tempFilesLock sync.Mutex

	go func() {
		<-c
		tempFilesLock.Lock()
		for _, f := range tempFiles {
			os.Remove(f)
		}
		tempFilesLock.Unlock()
		os.Exit(1)
	}()

	stdoutWriter := bufio.NewWriter(os.Stdout)
	defer stdoutWriter.Flush()

	if len(cli.Files) <= 1 || cli.Stream {
		if len(cli.Files) == 0 {
			processInput(stdoutWriter, e, os.Stdin, "")
			return
		}

		for _, path := range cli.Files {
			if err := processFile(stdoutWriter, e, path); err != nil {
				stdoutWriter.Flush()
				fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
			}
			stdoutWriter.Flush()
		}
		return
	}

	numCPU := runtime.NumCPU()
	sem := make(chan struct{}, numCPU)
	var wg sync.WaitGroup

	tempResults := make([]string, len(cli.Files))

	for i, path := range cli.Files {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			tmpPath, hasContent, err := processFileToTemp(e, p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", p, err)
				return
			}

			tempFilesLock.Lock()
			tempFiles = append(tempFiles, tmpPath)
			tempFilesLock.Unlock()

			if hasContent {
				tempResults[idx] = tmpPath
			} else {
				os.Remove(tmpPath)
				tempResults[idx] = ""
			}
		}(i, path)
	}

	wg.Wait()

	for _, tmpPath := range tempResults {
		if tmpPath == "" {
			continue
		}

		f, err := os.Open(tmpPath)
		if err == nil {
			io.Copy(stdoutWriter, f)
			f.Close()
		}

		os.Remove(tmpPath)
		stdoutWriter.Flush()
	}
}
