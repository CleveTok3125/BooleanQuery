package engine

import "strings"

func mergeIntervals(matches [][2]int) [][2]int {
	if len(matches) <= 1 {
		return matches
	}

	var result [][2]int
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

func (engine *Engine) Highlight(text string, matches [][2]int) string {
	if engine.Config.NoColor || len(matches) == 0 {
		return text
	}

	mergedMatches := mergeIntervals(matches)

	var sb strings.Builder
	cursor := 0

	for _, match := range mergedMatches {
		start, end := match[0], match[1]

		if start > len(text) {
			start = len(text)
		}
		if end > len(text) {
			end = len(text)
		}
		if start < cursor {
			start = cursor
		}

		sb.WriteString(text[cursor:start])

		sb.WriteString(ColorRed)
		sb.WriteString(text[start:end])
		sb.WriteString(ColorReset)

		cursor = end
	}

	if cursor < len(text) {
		sb.WriteString(text[cursor:])
	}

	return sb.String()
}
