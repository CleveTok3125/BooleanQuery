package main

import (
	"log"

	"BooleanQuery/app"
	"BooleanQuery/engine"
)

func main() {
	app.ParseCLI()

	e := engine.New()

	e.Config.BufferMaxSize = 1024 * 1024
	e.Config.BufferSize = 64 * 1024

	app.ApplyConfigToEngine(e)

	if err := e.SetSearchTerm(app.GetQuery()); err != nil {
		log.Fatal("Error splitting search term:", err)
	}
	e.Classify()

	app.Run(e)
}
