// Package engine contains the core for the entire program
package engine

import (
	"log"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

type searchTerm struct {
	splited     []string
	whiteList   []string // AND (+)
	blackList   []string // NOT (-)
	greyList    []string // OR (~)
	orderedList []string // ORDERED (^)
}

func (engine *Engine) SetSearchTerm(searchTerm string) {
	words, err := shell.Fields(searchTerm, nil)
	if err != nil {
		log.Fatal("Error splitting search term:", err)
	}

	engine.searchTerm.splited = words
}

func (engine *Engine) Classify() {
	lists := map[string]*[]string{
		"+": &engine.searchTerm.whiteList,
		"-": &engine.searchTerm.blackList,
		"~": &engine.searchTerm.greyList,
		"^": &engine.searchTerm.orderedList,
	}

	for _, term := range engine.searchTerm.splited {
		if strings.TrimSpace(term) == "" {
			continue
		}

		runes := []rune(term)
		firstChar := string(runes[0])
		rest := string(runes[1:])

		if list, ok := lists[firstChar]; ok {
			*list = append(*list, rest)
		} else {
			engine.searchTerm.whiteList = append(engine.searchTerm.whiteList, term)
		}

	}
}

func (engine *Engine) GetSearchTerm() searchTerm {
	return engine.searchTerm
}
