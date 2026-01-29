package engine

import (
	"bytes"
	"sort"
	"unicode"
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

func isBoundaryChar(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return true
	}
	r := rune(text[pos])
	return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
}

func (engine *Engine) findAllTermIndices(text string, term Term) [][2]int {
	var matches [][2]int
	var searchBytes []byte
	textBytes := []byte(text)

	if engine.Config.IgnoreCase {
		searchBytes = bytes.ToLower(textBytes)
	} else {
		searchBytes = textBytes
	}

	currentOffset := 0

	for currentOffset < len(searchBytes) {
		var start, end int
		if term.HasWildcard {
			s, e := engine.findWildcardInterval(searchBytes[currentOffset:], term)
			if s == -1 {
				break
			}
			start = s
			end = e
		} else {
			idx := bytes.Index(searchBytes[currentOffset:], term.Bytes)
			if idx == -1 {
				break
			}
			start = idx
			end = idx + len(term.Bytes)

			realStart := currentOffset + start
			realEnd := currentOffset + end
			if engine.Config.ExactWord {
				if !isBoundaryChar(text, realStart-1) || !isBoundaryChar(text, realEnd) {
					currentOffset = realStart + 1
					continue
				}
			}
		}

		realStart := currentOffset + start
		realEnd := currentOffset + end

		matches = append(matches, [2]int{realStart, realEnd})

		if realEnd > currentOffset {
			currentOffset = realEnd
		} else {
			currentOffset++
		}
	}

	return matches
}

func (engine *Engine) findTermIndex(text string, term Term, startIdx int) (int, int) {
	if startIdx >= len(text) {
		return -1, -1
	}

	var searchBytes []byte
	textBytes := []byte(text)

	if engine.Config.IgnoreCase {
		searchBytes = bytes.ToLower(textBytes)
	} else {
		searchBytes = textBytes
	}

	if startIdx >= len(searchBytes) {
		return -1, -1
	}

	currentIdx := startIdx

	for {
		var start, end int

		if term.HasWildcard {
			s, e := engine.findWildcardInterval(searchBytes[currentIdx:], term)
			if s == -1 {
				return -1, -1
			}
			start = s
			end = e
		} else {
			idx := bytes.Index(searchBytes[currentIdx:], term.Bytes)
			if idx == -1 {
				return -1, -1
			}
			start = idx
			end = idx + len(term.Bytes)
		}

		realStart := currentIdx + start
		realEnd := currentIdx + end

		if !engine.Config.ExactWord {
			return realStart, realEnd
		}

		if isBoundaryChar(text, realStart-1) && isBoundaryChar(text, realEnd) {
			return realStart, realEnd
		}

		currentIdx = realStart + 1
		if currentIdx >= len(searchBytes) {
			return -1, -1
		}
	}
}

func (engine *Engine) containAnyCheckOnlyBytes(list []Term, text []byte) int {
	for i, term := range list {
		if term.HasWildcard {
			start, _ := engine.findWildcardInterval(text, term)
			if start != -1 {
				return i
			}
		} else {
			if bytes.Contains(text, term.Bytes) {
				return i
			}
		}
	}
	return -1
}

func (engine *Engine) containAny(terms []Term, text string) [][2]int {
	var matches [][2]int

	for _, term := range terms {
		termMatches := engine.findAllTermIndices(text, term)

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
		if term.HasWildcard {
			start, _ := engine.findWildcardInterval(text, term)
			if start == -1 {
				return false
			}
		} else {
			if !bytes.Contains(text, term.Bytes) {
				return false
			}
		}
	}
	return true
}

func (engine *Engine) containAll(terms []Term, text string) [][2]int {
	var matches [][2]int

	for _, term := range terms {
		termMatches := engine.findAllTermIndices(text, term)

		if len(termMatches) == 0 {
			return nil
		}
		matches = append(matches, termMatches...)
	}
	return matches
}

func (engine *Engine) containOrderedCheckOnlyBytes(list []Term, text []byte) bool {
	currentIdx := 0
	for _, term := range list {
		searchArea := text[currentIdx:]

		if term.HasWildcard {
			start, end := engine.findWildcardInterval(searchArea, term)
			if start == -1 {
				return false
			}
			currentIdx += end
		} else {
			idx := bytes.Index(searchArea, term.Bytes)
			if idx == -1 {
				return false
			}
			currentIdx += idx + len(term.Bytes)
		}
	}
	return true
}

func (engine *Engine) containOrdered(terms []Term, text string) [][2]int {
	var matches [][2]int
	currentPos := 0

	for _, term := range terms {
		start, end := engine.findTermIndex(text, term, currentPos)

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
