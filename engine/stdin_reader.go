package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
)

func (engine *Engine) ProcessStream(input io.Reader, combinedFlag IntFlag, streamName string) func(func(int, []byte) bool) {
	return func(yield func(int, []byte) bool) {
		scanner := bufio.NewScanner(input)

		buf := make([]byte, 0, engine.Config.BufferSize)
		scanner.Buffer(buf, engine.Config.BufferMaxSize)

		separators := engine.Config.CharSep
		if separators == "" {
			separators = "\n"
		}

		if separators == "\n" {
			scanner.Split(bufio.ScanLines)
		} else if len(separators) == 1 {
			sep := separators[0]
			scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 {
					return 0, nil, nil
				}
				if i := bytes.IndexByte(data, sep); i >= 0 {
					return i + 1, data[0:i], nil
				}
				if atEOF {
					return len(data), data, nil
				}
				return 0, nil, nil
			})
		} else {
			scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 {
					return 0, nil, nil
				}
				if i := bytes.IndexAny(data, separators); i >= 0 {
					return i + 1, data[0:i], nil
				}
				if atEOF {
					return len(data), data, nil
				}
				return 0, nil, nil
			})
		}

		currentIndex := 0

		for scanner.Scan() {
			part := scanner.Bytes()

			if bytes.IndexByte(part, 0) != -1 {
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

			if combinedFlag&WORDS != 0 {
				remaining := part
				for {
					start := bytes.IndexFunc(remaining, func(r rune) bool {
						return !unicode.IsSpace(r)
					})
					if start == -1 {
						break
					}
					remaining = remaining[start:]

					end := bytes.IndexFunc(remaining, unicode.IsSpace)
					if end == -1 {
						if !yield(currentIndex, remaining) {
							return
						}
						currentIndex++
						break
					}

					if !yield(currentIndex, remaining[:end]) {
						return
					}
					currentIndex++
					remaining = remaining[end:]
				}
			} else {
				if !yield(currentIndex, part) {
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
