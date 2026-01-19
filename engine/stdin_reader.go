package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func splitAndTrim(input string, sep string) []string {
	parts := strings.Split(input, sep)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func applyIf(parts []string, cond bool, f func(string) []string) []string {
	if !cond {
		return parts
	}
	tmp := make([]string, 0, len(parts))
	for _, p := range parts {
		tmp = append(tmp, f(p)...)
	}
	return tmp
}

// This code is abandoned and can be removed at any time
// Because the separator has been directly integrated into the scanner
// func splitAnyKeepSep(s string, seps string) []string {
// 	sepMap := make(map[rune]struct{})
// 	for _, r := range seps {
// 		sepMap[r] = struct{}{}
// 	}
//
// 	var result []string
// 	var buf strings.Builder
//
// 	for _, r := range s {
// 		buf.WriteRune(r)
// 		if _, ok := sepMap[r]; ok {
// 			result = append(result, buf.String())
// 			buf.Reset()
// 		}
// 	}
//
// 	if buf.Len() > 0 {
// 		result = append(result, buf.String())
// 	}
// 	return result
// }

func (engine *Engine) ProcessStream(input io.Reader, combinedFlag IntFlag, streamName string) func(func(int, string) bool) {
	return func(yield func(int, string) bool) {
		scanner := bufio.NewScanner(input)

		buf := make([]byte, 0, engine.Config.BufferSize)
		scanner.Buffer(buf, engine.Config.BufferMaxSize)

		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			separators := engine.Config.CharSep
			if separators == "" {
				separators = "\n"
			}

			if i := bytes.IndexAny(data, separators); i >= 0 {
				return i + 1, data[0:i], nil
			}

			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		})

		currentIndex := 0

		for scanner.Scan() {
			tokenBytes := scanner.Bytes()

			if bytes.IndexByte(tokenBytes, 0) != -1 {
				if engine.Config.AllowBinary {
					continue
				} else {
					displayName := streamName
					if displayName == "" {
						displayName = "(stdin)"
					}
					fmt.Fprintf(os.Stderr, "Binary file detected: %s\n", displayName)
					return
				}
			}

			part := string(tokenBytes)
			parts := []string{part}

			parts = applyIf(parts, combinedFlag&WORDS != 0, func(p string) []string { return splitAndTrim(p, "") })

			for _, p := range parts {
				if !yield(currentIndex, p) {
					return
				}
				currentIndex++
			}
		}

		if err := scanner.Err(); err != nil {
			if err == bufio.ErrTooLong {
				fmt.Fprintf(os.Stderr, "\nSkipped: The data is too long, exceeding the threshold of %d bytes (Binary file or stream is too large).\n", engine.Config.BufferMaxSize)
			} else {
				fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
			}
		}
	}
}
