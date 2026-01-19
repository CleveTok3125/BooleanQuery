// Package app handling processes & user interfaces
package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"BooleanQuery/engine"
)

type lineInfo struct {
	index int
	text  string
}

func printLine(w io.Writer, e *engine.Engine, index int, matches [][2]int, part string, filename string) {
	cIndex := engine.ColorBrightBlack
	cReset := engine.ColorReset
	cFile := engine.ColorMagenta

	if e.Config.NoColor {
		cIndex = ""
		cReset = ""
		cFile = ""
	}

	prefix := ""
	filePrefix := ""

	if filename != "" && cli.ShowFilePrefix {
		filePrefix = fmt.Sprintf("%s%s%s:", cFile, filename, cReset)
	}

	if !cli.NoIndex {
		lineStr := fmt.Sprintf("%*d", padLine, index+1)

		colDisplay := -1
		if len(matches) > 0 {
			colDisplay = matches[0][0]
		}

		if colDisplay != -1 {
			prefix = fmt.Sprintf("%s%s%s:%-*d|%s ", filePrefix, cIndex, lineStr, padCol, colDisplay+1, cReset)
		} else {
			spacePadding := ""
			if padCol > 0 {
				spacePadding = strings.Repeat(" ", padCol)
			}

			prefix = fmt.Sprintf("%s%s%s:%s|%s ", filePrefix, cIndex, lineStr, spacePadding, cReset)
		}
	} else if filePrefix != "" {
		prefix = filePrefix + " "
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

	iter := e.ProcessStream(reader, combinedFlag, displayName)

	iter(func(index int, part string) bool {
		matches := e.Search(part)

		if matches != nil {
			shouldPrintHeader := !cli.NoFileHeader && len(cli.Files) > 1

			if shouldPrintHeader && !headerPrinted && displayName != "" {
				fmt.Fprintf(w, "\n--- File: %s ---\n\n", displayName)
				headerPrinted = true
			}

			for _, item := range beforeBuffer {
				if item.index > lastPrintedIndex {
					printLine(w, e, item.index, nil, item.text, displayName)
					lastPrintedIndex = item.index
				}
			}

			printLine(w, e, index, matches, part, displayName)
			lastPrintedIndex = index

			beforeBuffer = nil
			linesToPrintAfter = cli.After

		} else {
			if linesToPrintAfter > 0 {
				if index > lastPrintedIndex {
					printLine(w, e, index, nil, part, displayName)
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

func getDisplayName(path string) string {
	return path
}

func processFile(w io.Writer, e *engine.Engine, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	displayName := getDisplayName(path)
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

	displayName := getDisplayName(path)
	processInput(bufWriter, e, f, displayName)

	bufWriter.Flush()

	stat, _ := tmpFile.Stat()
	hasContent := stat.Size() > 0

	return tmpFile.Name(), hasContent, nil
}
