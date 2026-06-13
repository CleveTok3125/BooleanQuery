package engine

import (
	"bytes"
)

func mergeIntervals(matches [][2]int) [][2]int {
	if len(matches) <= 1 {
		return matches
	}

	result := make([][2]int, 0, len(matches))
	current := matches[0]

	for i := 1; i < len(matches); i++ {
		next := matches[i]

		if next[0] <= current[1] {
			if next[1] > current[1] {
				current[1] = next[1]
			}
		} else {
			result = append(result, current)
			current = next
		}
	}
	result = append(result, current)
	return result
}

func (engine *Engine) HighlightTo(buf *bytes.Buffer, text string, matches [][2]int) {
	if engine.Config.NoColor || len(matches) == 0 {
		buf.WriteString(text)
		return
	}

	mergedMatches := mergeIntervals(matches)

	cursor := 0
	textLen := len(text)

	for _, match := range mergedMatches {
		start, end := match[0], match[1]

		if start > textLen {
			start = textLen
		}
		if end > textLen {
			end = textLen
		}
		if start < cursor {
			start = cursor
		}

		if start > cursor {
			buf.WriteString(text[cursor:start])
		}

		buf.WriteString(ColorRed)
		buf.WriteString(text[start:end])
		buf.WriteString(ColorReset)

		cursor = end
	}

	if cursor < textLen {
		buf.WriteString(text[cursor:])
	}
}
