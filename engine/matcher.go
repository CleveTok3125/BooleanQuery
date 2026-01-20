package engine

import (
	"bytes"
	"sort"
	"strings"
	"unicode"
)

func isBoundaryChar(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return true
	}
	r := rune(text[pos])
	return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
}

func (engine *Engine) findAllTermIndices(text string, term string) [][2]int {
	var matches [][2]int
	currentIdx := 0
	termLen := len(term)

	for {
		idx := strings.Index(text[currentIdx:], term)
		if idx == -1 {
			break
		}

		realStart := currentIdx + idx
		realEnd := realStart + termLen

		isValid := true
		if engine.Config.ExactWord {
			if !isBoundaryChar(text, realStart-1) || !isBoundaryChar(text, realEnd) {
				isValid = false
			}
		}

		if isValid {
			matches = append(matches, [2]int{realStart, realEnd})
		}

		currentIdx = realStart + 1
	}
	return matches
}

func (engine *Engine) findTermIndex(text string, term string, startIdx int) (int, int) {
	currentIdx := startIdx
	termLen := len(term)

	for {
		idx := strings.Index(text[currentIdx:], term)
		if idx == -1 {
			return -1, -1
		}

		realStart := currentIdx + idx
		realEnd := realStart + termLen

		if !engine.Config.ExactWord {
			return realStart, realEnd
		}

		if isBoundaryChar(text, realStart-1) && isBoundaryChar(text, realEnd) {
			return realStart, realEnd
		}

		currentIdx = realStart + 1
	}
}

func (engine *Engine) containAnyCheckOnly(terms []string, text string) int {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	for _, term := range terms {
		matchTerm := term

		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		start, _ := engine.findTermIndex(matchText, matchTerm, 0)
		if start != -1 {
			return start
		}
	}

	return -1
}

func (engine *Engine) containAnyCheckOnlyBytes(list []string, text []byte) int {
	for i, term := range list {
		if bytes.Contains(text, []byte(term)) {
			return i
		}
	}
	return -1
}

func (engine *Engine) containAny(terms []string, text string) [][2]int {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	var matches [][2]int

	for _, term := range terms {
		matchTerm := term
		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		termMatches := engine.findAllTermIndices(matchText, matchTerm)

		if len(termMatches) > 0 {
			matches = append(matches, termMatches...)
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return matches
}

func (engine *Engine) containAllCheckOnly(terms []string, text string) bool {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	for _, term := range terms {
		matchTerm := term
		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		start, _ := engine.findTermIndex(matchText, matchTerm, 0)
		if start == -1 {
			return false
		}
	}
	return true
}

func (engine *Engine) containAllCheckOnlyBytes(list []string, text []byte) bool {
	for _, term := range list {
		if !bytes.Contains(text, []byte(term)) {
			return false
		}
	}
	return true
}

func (engine *Engine) containAll(terms []string, text string) [][2]int {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	var matches [][2]int

	for _, term := range terms {
		matchTerm := term
		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		termMatches := engine.findAllTermIndices(matchText, matchTerm)

		if len(termMatches) == 0 {
			return nil
		}

		matches = append(matches, termMatches...)
	}

	return matches
}

func (engine *Engine) containOrderedCheckOnly(terms []string, text string) bool {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	currentPos := 0
	for _, term := range terms {
		matchTerm := term
		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		_, end := engine.findTermIndex(matchText, matchTerm, currentPos)
		if end == -1 {
			return false
		}
		currentPos = end
	}
	return true
}

func (engine *Engine) containOrderedCheckOnlyBytes(list []string, text []byte) bool {
	currentIdx := 0
	for _, term := range list {
		idx := bytes.Index(text[currentIdx:], []byte(term))
		if idx == -1 {
			return false
		}
		currentIdx += idx + len(term)
	}
	return true
}

func (engine *Engine) containOrdered(terms []string, text string) [][2]int {
	matchText := text
	if engine.Config.IgnoreCase {
		matchText = strings.ToLower(text)
	}

	var matches [][2]int
	currentPos := 0

	for _, term := range terms {
		matchTerm := term
		if engine.Config.IgnoreCase {
			matchTerm = strings.ToLower(term)
		}

		start, end := engine.findTermIndex(matchText, matchTerm, currentPos)

		if start == -1 {
			return nil
		}

		matches = append(matches, [2]int{start, end})

		currentPos = end
	}

	return matches
}

func (engine *Engine) CheckOnlyBytes(text []byte) bool {
	if engine.Config.IgnoreCase {
		text = bytes.ToLower(text)
	}

	if engine.containAnyCheckOnlyBytes(engine.searchTerm.blackList, text) != -1 {
		return false
	}

	if len(engine.searchTerm.orderedList) > 0 {
		if !engine.containOrderedCheckOnlyBytes(engine.searchTerm.orderedList, text) {
			return false
		}
	}

	if len(engine.searchTerm.whiteList) > 0 {
		if !engine.containAllCheckOnlyBytes(engine.searchTerm.whiteList, text) {
			return false
		}
	}

	if len(engine.searchTerm.greyList) > 0 {
		if engine.containAnyCheckOnlyBytes(engine.searchTerm.greyList, text) == -1 {
			return false
		}
	}

	return true
}

func (engine *Engine) Search(text string) [][2]int {
	if engine.containAnyCheckOnly(engine.searchTerm.blackList, text) != -1 {
		return nil
	}

	if engine.Config.NoColor {
		if len(engine.searchTerm.orderedList) > 0 {
			if !engine.containOrderedCheckOnly(engine.searchTerm.orderedList, text) {
				return nil
			}
		}

		if len(engine.searchTerm.whiteList) > 0 {
			if !engine.containAllCheckOnly(engine.searchTerm.whiteList, text) {
				return nil
			}
		}

		return [][2]int{}
	}

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
