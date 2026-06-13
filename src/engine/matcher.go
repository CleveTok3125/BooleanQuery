package engine

import (
	"bytes"
	"sort"
)

func indexWithQuestionMark(s, sub []byte) int {
	if bytes.IndexByte(sub, WildcardZero) == -1 {
		return bytes.Index(s, sub)
	}

	n := len(sub)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-n; i++ {
		match := true
		for j := range n {
			if sub[j] != WildcardZero && sub[j] != s[i+j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (engine *Engine) findWildcardInterval(text []byte, term Term) (int, int) {
	if len(term.Segments) == 0 {
		return 0, len(text)
	}

	startIndex := -1
	currentPos := 0

	for _, segment := range term.Segments {
		if len(segment) == 0 {
			continue
		}

		idx := indexWithQuestionMark(text[currentPos:], segment)
		if idx == -1 {
			return -1, -1
		}

		realIdx := currentPos + idx

		if startIndex == -1 {
			startIndex = realIdx
		}

		currentPos = realIdx + len(segment)
	}

	finalEnd := currentPos
	finalStart := startIndex

	if len(term.Bytes) > 0 && term.Bytes[0] == '*' {
		finalStart = 0
	}
	if len(term.Bytes) > 0 && term.Bytes[len(term.Bytes)-1] == '*' {
		finalEnd = len(text)
	}

	if startIndex == -1 {
		return 0, len(text)
	}

	return finalStart, finalEnd
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func (engine *Engine) findTermIndex(text []byte, term Term, startOffset int) (int, int) {
	if startOffset >= len(text) {
		return -1, -1
	}

	offset := startOffset
	for offset < len(text) {
		var start, end int

		if term.HasWildcard {
			s, e := engine.findWildcardInterval(text[offset:], term)
			if s == -1 {
				return -1, -1
			}
			start = s
			end = e
		} else {
			idx := bytes.Index(text[offset:], term.Bytes)
			if idx == -1 {
				return -1, -1
			}
			start = idx
			end = idx + len(term.Bytes)
		}

		realStart := offset + start
		realEnd := offset + end

		if !engine.Config.ExactWord {
			return realStart, realEnd
		}

		isStartBoundary := realStart == 0 || !isWordChar(text[realStart-1])
		isEndBoundary := realEnd == len(text) || !isWordChar(text[realEnd])

		if isStartBoundary && isEndBoundary {
			return realStart, realEnd
		}

		offset = realStart + 1
	}
	return -1, -1
}

func (engine *Engine) containAnyCheckOnlyBytes(list []Term, text []byte) int {
	for i, term := range list {
		start, _ := engine.findTermIndex(text, term, 0)
		if start != -1 {
			return i
		}
	}
	return -1
}

func (engine *Engine) containAny(terms []Term, text string) [][2]int {
	var matches [][2]int
	var textBytes []byte
	if engine.Config.IgnoreCase {
		textBytes = bytes.ToLower([]byte(text))
	} else {
		textBytes = []byte(text)
	}

	for _, term := range terms {
		var termMatches [][2]int
		currentPos := 0
		for currentPos < len(textBytes) {
			start, end := engine.findTermIndex(textBytes, term, currentPos)
			if start == -1 {
				break
			}
			termMatches = append(termMatches, [2]int{start, end})
			if end > start {
				currentPos = end
			} else {
				currentPos = start + 1
			}
		}
		if len(termMatches) > 0 {
			matches = append(matches, termMatches...)
		}
	}
	if len(matches) == 0 {
		return nil
	}
	return matches
}

func (engine *Engine) containAllCheckOnlyBytes(list []Term, text []byte) bool {
	for _, term := range list {
		start, _ := engine.findTermIndex(text, term, 0)
		if start == -1 {
			return false
		}
	}
	return true
}

func (engine *Engine) containAll(terms []Term, text string) [][2]int {
	var matches [][2]int
	var textBytes []byte
	if engine.Config.IgnoreCase {
		textBytes = bytes.ToLower([]byte(text))
	} else {
		textBytes = []byte(text)
	}

	for _, term := range terms {
		var termMatches [][2]int
		currentPos := 0
		for currentPos < len(textBytes) {
			start, end := engine.findTermIndex(textBytes, term, currentPos)
			if start == -1 {
				break
			}
			termMatches = append(termMatches, [2]int{start, end})
			if end > start {
				currentPos = end
			} else {
				currentPos = start + 1
			}
		}
		if len(termMatches) == 0 {
			return nil
		}
		matches = append(matches, termMatches...)
	}
	return matches
}

func (engine *Engine) containOrderedCheckOnlyBytes(list []Term, text []byte) bool {
	currentPos := 0
	for _, term := range list {
		start, end := engine.findTermIndex(text, term, currentPos)
		if start == -1 {
			return false
		}
		currentPos = end
	}
	return true
}

func (engine *Engine) containOrdered(terms []Term, text string) [][2]int {
	var matches [][2]int
	var textBytes []byte
	if engine.Config.IgnoreCase {
		textBytes = bytes.ToLower([]byte(text))
	} else {
		textBytes = []byte(text)
	}

	currentPos := 0
	for _, term := range terms {
		start, end := engine.findTermIndex(textBytes, term, currentPos)
		if start == -1 {
			return nil
		}
		matches = append(matches, [2]int{start, end})
		currentPos = end
	}
	return matches
}

func (engine *Engine) CheckOnlyBytes(text []byte) bool {
	var textToCheck []byte
	if engine.Config.IgnoreCase {
		textToCheck = bytes.ToLower(text)
	} else {
		textToCheck = text
	}

	if engine.containAnyCheckOnlyBytes(engine.searchTerm.blackList, textToCheck) != -1 {
		return false
	}

	if len(engine.searchTerm.orderedList) > 0 {
		if !engine.containOrderedCheckOnlyBytes(engine.searchTerm.orderedList, textToCheck) {
			return false
		}
	}

	if len(engine.searchTerm.whiteList) > 0 {
		if !engine.containAllCheckOnlyBytes(engine.searchTerm.whiteList, textToCheck) {
			return false
		}
	}

	if len(engine.searchTerm.greyList) > 0 {
		if engine.containAnyCheckOnlyBytes(engine.searchTerm.greyList, textToCheck) == -1 {
			return false
		}
	}

	return true
}

func (engine *Engine) Search(text string) [][2]int {
	var allMatches [][2]int

	if len(engine.searchTerm.orderedList) > 0 {
		orderedMatches := engine.containOrdered(engine.searchTerm.orderedList, text)
		if orderedMatches == nil {
			return nil
		}
		allMatches = append(allMatches, orderedMatches...)
	}

	if len(engine.searchTerm.whiteList) > 0 {
		whiteMatches := engine.containAll(engine.searchTerm.whiteList, text)
		if whiteMatches == nil {
			return nil
		}
		allMatches = append(allMatches, whiteMatches...)
	}

	if len(engine.searchTerm.greyList) > 0 {
		greyMatches := engine.containAny(engine.searchTerm.greyList, text)
		if greyMatches != nil {
			allMatches = append(allMatches, greyMatches...)
		}
	}

	if len(allMatches) == 0 {
		if len(engine.searchTerm.orderedList) == 0 &&
			len(engine.searchTerm.whiteList) == 0 &&
			len(engine.searchTerm.greyList) == 0 {
			return [][2]int{}
		}
		return nil
	}

	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i][0] < allMatches[j][0]
	})

	return allMatches
}
