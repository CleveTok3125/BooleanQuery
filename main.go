package main

import (
	"log"
	"os"

	"BooleanQuery/app"
	"BooleanQuery/engine"
)

func main() {
	app.ParseCLI()

	e := engine.New()

	app.ApplyConfigToEngine(e)

	if err := e.SetSearchTerm(app.GetQuery()); err != nil {
		log.Fatal("Error splitting search term:", err)
	}
	e.Classify()

	found := app.Run(e)

	if found {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
