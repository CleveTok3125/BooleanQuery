// Package engine contains the core for the entire program
package engine

import (
	"bytes"
	"errors"
	"strings"
)

const (
	SafeEscapeChar = '@'
	WildcardZero   = 0 // Use byte 0 as a replacement character for the wildcard `?`
)

type Term struct {
	Bytes       []byte
	Segments    [][]byte
	HasWildcard bool
}

type searchTerm struct {
	splited     []string
	whiteList   []Term // AND (+)
	blackList   []Term // NOT (-)
	greyList    []Term // OR (~)
	orderedList []Term // ORDERED (^)
}

func parseWildcard(text string) ([][]byte, bool) {
	var segments [][]byte
	var currentSegment bytes.Buffer
	hasWildcard := false

	n := len(text)
	for i := 0; i < n; i++ {
		char := text[i]

		if char == SafeEscapeChar {
			if i+1 < n {
				nextChar := text[i+1]

				if nextChar == '*' || nextChar == '?' || nextChar == SafeEscapeChar {
					currentSegment.WriteByte(nextChar)
					i++
					continue
				}
			}

			currentSegment.WriteByte(char)
			continue
		}

		if char == '*' {
			hasWildcard = true

			segments = append(segments, append([]byte(nil), currentSegment.Bytes()...))
			currentSegment.Reset()
			continue
		}

		if char == '?' {
			hasWildcard = true
			currentSegment.WriteByte(WildcardZero)
			continue
		}

		currentSegment.WriteByte(char)
	}

	segments = append(segments, append([]byte(nil), currentSegment.Bytes()...))

	return segments, hasWildcard
}

func createTerm(text string, wildcardMode bool) Term {
	if !wildcardMode {
		return Term{Bytes: []byte(text)}
	}

	segments, hasWildcard := parseWildcard(text)

	t := Term{
		Segments:    segments,
		HasWildcard: hasWildcard,
	}

	if !hasWildcard && len(segments) > 0 {
		t.Bytes = segments[0]
	} else {
		t.Bytes = []byte(text)
	}

	return t
}

func splitWords(input string) ([]string, error) {
	var words []string
	var buf bytes.Buffer
	i, n := 0, len(input)

	for i < n {
		switch ch := input[i]; {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			if buf.Len() > 0 {
				words = append(words, buf.String())
				buf.Reset()
			}
			i++
		case ch == '\'':
			if strings.IndexByte(input[i+1:], '\'') == -1 {
				buf.WriteByte(ch)
				i++
				break
			}
			i++
			for i < n {
				if input[i] == '\'' {
					break
				}
				if input[i] == '\\' && i+1 < n && (input[i+1] == '\'' || input[i+1] == '\\') {
					i++
				}
				buf.WriteByte(input[i])
				i++
			}
			if i >= n {
				return nil, errors.New("unclosed single quote")
			}
			i++ // skip closing '
		case ch == '"':
			if strings.IndexByte(input[i+1:], '"') == -1 {
				buf.WriteByte(ch)
				i++
				break
			}
			i++
			for i < n {
				if input[i] == '"' {
					break
				}
				if input[i] == '\\' && i+1 < n && (input[i+1] == '"' || input[i+1] == '\\') {
					i++
				}
				buf.WriteByte(input[i])
				i++
			}
			if i >= n {
				return nil, errors.New("unclosed double quote")
			}
			i++ // skip closing "
		default:
			buf.WriteByte(ch)
			i++
		}
	}

	if buf.Len() > 0 {
		words = append(words, buf.String())
	}

	return words, nil
}

func (engine *Engine) SetSearchTerm(searchTerm string) error {
	return engine.SetSearchTerms([]string{searchTerm})
}

func (engine *Engine) SetSearchTerms(terms []string) error {
	engine.searchTerms = make([]searchTerm, len(terms))
	for i, term := range terms {
		words, err := splitWords(term)
		if err != nil {
			return err
		}
		engine.searchTerms[i] = searchTerm{splited: words}
		engine.classifyQuery(&engine.searchTerms[i])
	}
	return nil
}

func (engine *Engine) classifyQuery(st *searchTerm) {
	st.whiteList = nil
	st.blackList = nil
	st.greyList = nil
	st.orderedList = nil

	lists := map[string]*[]Term{
		"+": &st.whiteList,
		"-": &st.blackList,
		"~": &st.greyList,
		"^": &st.orderedList,
	}

	for _, termStr := range st.splited {
		if termStr == "" {
			continue
		}

		runes := []rune(termStr)
		firstChar := string(runes[0])
		rest := string(runes[1:])

		if engine.Config.IgnoreCase {
			rest = strings.ToLower(rest)
		}

		var t Term
		if list, ok := lists[firstChar]; ok {
			t = createTerm(rest, engine.Config.Wildcard)
			*list = append(*list, t)
		} else {
			termToAdd := termStr
			if engine.Config.IgnoreCase {
				termToAdd = strings.ToLower(termStr)
			}
			t = createTerm(termToAdd, engine.Config.Wildcard)
			st.whiteList = append(st.whiteList, t)
		}
	}
}

func (engine *Engine) Classify() {
	for i := range engine.searchTerms {
		engine.classifyQuery(&engine.searchTerms[i])
	}
}

func (engine *Engine) GetSearchTerm() searchTerm {
	if len(engine.searchTerms) > 0 {
		return engine.searchTerms[0]
	}
	return searchTerm{}
}
