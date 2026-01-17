package engine

import (
	"bufio"
	"io"
	"log"
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

func splitAnyKeepSep(s string, seps string) []string {
	sepMap := make(map[rune]struct{})
	for _, r := range seps {
		sepMap[r] = struct{}{}
	}

	var result []string
	var buf strings.Builder

	for _, r := range s {
		buf.WriteRune(r)
		if _, ok := sepMap[r]; ok {
			result = append(result, buf.String())
			buf.Reset()
		}
	}

	if buf.Len() > 0 {
		result = append(result, buf.String())
	}
	return result
}

func (engine *Engine) ProcessStream(input io.Reader, combinedFlag IntFlag) func(func(int, string) bool) {
	return func(yield func(int, string) bool) {
		scanner := bufio.NewScanner(input)

		buf := make([]byte, 0, engine.Config.BufferSize)
		scanner.Buffer(buf, engine.Config.BufferMaxSize)

		currentIndex := 0

		for scanner.Scan() {
			rawLine := scanner.Text()

			parts := []string{rawLine}

			parts = applyIf(parts, combinedFlag&CHARSEP != 0 && engine.Config.CharSep != "", func(p string) []string { return splitAnyKeepSep(p, engine.Config.CharSep) })

			parts = applyIf(parts, combinedFlag&WORDS != 0, func(p string) []string { return splitAndTrim(p, "") })

			for _, part := range parts {
				if !yield(currentIndex, part) {
					return
				}
				currentIndex++
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading stdin: %v", err)
		}
	}
}
