package engine

type Engine struct {
	Config Config

	searchTerm searchTerm
}

func New() *Engine {
	return &Engine{}
}
