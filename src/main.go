package main

import (
	"log"
	"os"

	"BooleanQuery/src/app"
	"BooleanQuery/src/engine"
)

func main() {
	app.ParseCLI()

	e := engine.New()

	app.ApplyConfigToEngine(e)

	if err := e.SetSearchTerms(app.GetQueries()); err != nil {
		log.Fatal("Error splitting search term:", err)
	}

	found := app.Run(e)

	if found {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
