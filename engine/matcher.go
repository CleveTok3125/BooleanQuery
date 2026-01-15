package engine

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

func buildLPS(pattern string) []int {
	m := len(pattern)
	lps := make([]int, m)

	lenPrev := 0
	i := 1

	for i < m {
		if pattern[i] == pattern[lenPrev] {
			lenPrev++
			lps[i] = lenPrev
			i++
		} else {
			if lenPrev != 0 {
				lenPrev = lps[lenPrev-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	return lps
}

func KMPSearch(text, pattern string) []int {
	n, m := len(text), len(pattern)
	if m == 0 {
		positions := make([]int, n+1)
		for i := range positions {
			positions[i] = i
		}
		return positions
	}

	if n < m {
		return nil
	}

	lps := buildLPS(pattern)
	var result []int

	i, j := 0, 0
	for i < n {
		if text[i] == pattern[j] {
			i++
			j++
			if j == m {
				result = append(result, i-j)
				j = lps[j-1]
			}
		} else {
			if j != 0 {
				j = lps[j-1]
			} else {
				i++
			}
		}
	}
	return result
}

func searchWord(term string, text string) (bool, error) {
	escaped := regexp.QuoteMeta(term)
	pattern := fmt.Sprintf(`\b%s\b`, escaped)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(text), nil
}

func (engine *Engine) containAnyHelper(term string, text string) bool {
	if engine.Config.ExactWord {
		ok, err := searchWord(term, text)
		if err != nil {
			log.Printf("searchWord error: %v", err)
		}

		if ok {
			return true
		}
	}

	if !engine.Config.ExactWord && len(KMPSearch(text, term)) > 0 {
		return true
	}

	return false
}

func (engine *Engine) containAny(terms []string, text string) bool {
	if engine.Config.IgnoreCase {
		text = strings.ToLower(text)
	}

	for _, term := range terms {
		if engine.Config.IgnoreCase {
			term = strings.ToLower(term)
		}

		if engine.containAnyHelper(term, text) {
			return true
		}
	}

	return false
}

func (engine *Engine) containAll(terms []string, text string) bool {
	if engine.Config.IgnoreCase {
		text = strings.ToLower(text)
	}

	for _, term := range terms {
		if engine.Config.IgnoreCase {
			term = strings.ToLower(term)
		}

		if !engine.containAnyHelper(term, text) {
			return false
		}
	}

	return true
}

func (engine *Engine) IsMatch(text string) bool {
	blacklist := engine.searchTerm.blackList
	whitelist := engine.searchTerm.whiteList
	greylist := engine.searchTerm.greyList

	if len(blacklist) > 0 && engine.containAny(blacklist, text) {
		return false
	}

	if len(whitelist) > 0 && !engine.containAll(whitelist, text) {
		return false
	}

	if len(greylist) > 0 && !engine.containAny(greylist, text) {
		return false
	}

	return true
}

func (engine *Engine) PrepareHighlight() {
	terms := append([]string{}, engine.searchTerm.whiteList...)
	terms = append(terms, engine.searchTerm.greyList...)

	if len(terms) == 0 {
		return
	}

	var patterns []string
	for _, term := range terms {
		escaped := regexp.QuoteMeta(term)
		if engine.Config.ExactWord {
			patterns = append(patterns, fmt.Sprintf(`\b%s\b`, escaped))
		} else {
			patterns = append(patterns, escaped)
		}
	}

	fullPattern := strings.Join(patterns, "|")
	if engine.Config.IgnoreCase {
		fullPattern = "(?i)(" + fullPattern + ")"
	} else {
		fullPattern = "(" + fullPattern + ")"
	}

	re, err := regexp.Compile(fullPattern)
	if err != nil {
		log.Printf("Error compiling highlight regex: %v", err)
		return
	}
	engine.highlightRe = re
}

func (engine *Engine) Highlight(text string) string {
	if engine.Config.NoColor {
		return text
	}

	if engine.highlightRe == nil {
		return text
	}

	return engine.highlightRe.ReplaceAllString(text, ColorRed+"$1"+ColorReset)
}
