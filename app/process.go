// Package app handling processes & user interfaces
package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"BooleanQuery/engine"
)

type lineInfo struct {
	index int
	text  string
}

func writePadding(buf *bytes.Buffer, count int) {
	for range count {
		buf.WriteByte(' ')
	}
}

func printLine(w io.Writer, e *engine.Engine, index int, matches [][2]int, part string, cachedFilePrefix []byte, buf *bytes.Buffer) {
	buf.Reset()

	if len(cachedFilePrefix) > 0 {
		buf.Write(cachedFilePrefix)
	}

	if !cli.NoIndex {
		if !e.Config.NoColor {
			buf.WriteString(engine.ColorBrightBlack)
		}

		lineStr := strconv.Itoa(index + 1)

		if len(lineStr) < padLine {
			writePadding(buf, padLine-len(lineStr))
		}
		buf.WriteString(lineStr)

		colDisplay := -1
		if len(matches) > 0 {
			colDisplay = matches[0][0]
		}

		if colDisplay != -1 {
			buf.WriteByte(':')
			colStr := strconv.Itoa(colDisplay + 1)
			buf.WriteString(colStr)

			if len(colStr) < padCol {
				writePadding(buf, padCol-len(colStr))
			}
			buf.WriteByte('|')
		} else {
			buf.WriteByte(':')
			writePadding(buf, padCol)
			buf.WriteByte('|')
		}

		buf.WriteByte(' ')

		if !e.Config.NoColor {
			buf.WriteString(engine.ColorReset)
		}
	}

	e.HighlightTo(buf, part, matches)
	buf.WriteByte('\n')

	w.Write(buf.Bytes())
}

func processInput(w io.Writer, e *engine.Engine, reader io.Reader, displayName string) bool {
	combinedFlag := engine.CombineFlags(engine.CHARSEP)

	shouldPrintHeader := !cli.NoFileHeader && len(cli.Files) > 1 && !cli.Count

	var cachedFilePrefix []byte
	if displayName != "" && cli.ShowFilePrefix {
		tmp := bytes.NewBuffer(make([]byte, 0, 100))
		if !e.Config.NoColor {
			tmp.WriteString(engine.ColorMagenta)
		}
		tmp.WriteString(displayName)
		if !e.Config.NoColor {
			tmp.WriteString(engine.ColorReset)
		}
		tmp.WriteByte(':')
		cachedFilePrefix = tmp.Bytes()
	}

	var beforeBuffer []lineInfo
	linesToPrintAfter := 0
	lastPrintedIndex := -1
	headerPrinted := false

	foundAnyMatch := false
	matchCount := 0

	printBuf := bytes.NewBuffer(make([]byte, 0, e.Config.BufferSize))

	iter := e.ProcessStream(reader, combinedFlag, displayName)

	iter(func(index int, partBytes []byte) bool {
		if !e.CheckOnlyBytes(partBytes) {
			if !cli.Count && (linesToPrintAfter > 0 || cli.Before > 0) {
				partStr := string(partBytes)

				if linesToPrintAfter > 0 {
					if index > lastPrintedIndex {
						printLine(w, e, index, nil, partStr, cachedFilePrefix, printBuf)
						lastPrintedIndex = index
					}
					linesToPrintAfter--
				}

				if cli.Before > 0 {
					beforeBuffer = append(beforeBuffer, lineInfo{index, partStr})
					if len(beforeBuffer) > cli.Before {
						beforeBuffer = beforeBuffer[1:]
					}
				}
			}

			return true
		}

		foundAnyMatch = true

		if cli.FilesWithMatches {
			fmt.Fprintln(w, displayName)
			return false
		}

		matchCount++

		if cli.Count {
			return true
		}

		part := string(partBytes)

		var matches [][2]int
		if !e.Config.NoColor {
			matches = e.Search(part)
		}

		if shouldPrintHeader && !headerPrinted && displayName != "" {
			fmt.Fprintf(w, "\n--- File: %s ---\n\n", displayName)
			headerPrinted = true
		}

		for _, item := range beforeBuffer {
			if item.index > lastPrintedIndex {
				printLine(w, e, item.index, nil, item.text, cachedFilePrefix, printBuf)
				lastPrintedIndex = item.index
			}
		}

		printLine(w, e, index, matches, part, cachedFilePrefix, printBuf)
		lastPrintedIndex = index

		beforeBuffer = nil
		linesToPrintAfter = cli.After

		return true
	})

	if cli.Count {
		if displayName != "" && (len(cli.Files) > 1 || cli.ShowFilePrefix) {
			if !e.Config.NoColor {
				fmt.Fprintf(w, "%s%s%s:%d\n", engine.ColorMagenta, displayName, engine.ColorReset, matchCount)
			} else {
				fmt.Fprintf(w, "%s:%d\n", displayName, matchCount)
			}
		} else {
			fmt.Fprintf(w, "%d\n", matchCount)
		}
	}

	return foundAnyMatch
}

func processFile(w io.Writer, e *engine.Engine, path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	found := processInput(w, e, f, path)
	return found, nil
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

	processInput(bufWriter, e, f, path)

	bufWriter.Flush()

	stat, _ := tmpFile.Stat()
	hasContent := stat.Size() > 0

	return tmpFile.Name(), hasContent, nil
}
