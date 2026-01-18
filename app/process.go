// Package app handling user interfaces
package app

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"BooleanQuery/engine"
)

type lineInfo struct {
	index int
	text  string
}

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
