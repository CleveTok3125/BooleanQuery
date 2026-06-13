package engine

import "errors"

var ErrBinaryFile = errors.New("binary file detected")

type Engine struct {
	Config      Config
	searchTerms []searchTerm
}

func New() *Engine {
	return &Engine{}
}
