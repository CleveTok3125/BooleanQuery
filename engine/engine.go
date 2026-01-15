package engine

import "regexp"

type Engine struct {
	Config Config

	searchTerm  searchTerm
	highlightRe *regexp.Regexp
}

func New() *Engine {
	return &Engine{}
}
