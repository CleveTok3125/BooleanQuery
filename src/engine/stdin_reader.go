package engine

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

func (engine *Engine) ProcessStream(input io.Reader, combinedFlag IntFlag) func(func(int, []byte) bool) error {
	return func(yield func(int, []byte) bool) error {
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

			if isBinaryContent(part) {
				if engine.Config.AllowBinary {
					continue
				} else {
					return ErrBinaryFile
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
							return nil
						}
						currentIndex++
						break
					}

					if !yield(currentIndex, remaining[:end]) {
						return nil
					}
					currentIndex++
					remaining = remaining[end:]
				}
			} else {
				if !yield(currentIndex, part) {
					return nil
				}
				currentIndex++
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
		return nil
	}
}

func isBinaryContent(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
		if b < 32 && b != 9 && b != 10 && b != 13 {
			return true
		}
	}
	return false
}
