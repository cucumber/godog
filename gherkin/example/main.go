package main

import (
	"log"
	"os"

	"github.com/DATA-DOG/godog/gherkin"
)

func main() {
	feature, err := gherkin.Parse("ls.feature")
	switch {
	case err == gherkin.ErrEmpty:
		log.Println("the feature file is empty and does not describe any feature")
		return
	case err != nil:
		log.Println("the feature file is incorrect or could not be read:", err)
		os.Exit(1)
	}
	log.Println("have parsed a feature:", feature.Title, "with", len(feature.Scenarios), "scenarios")
}
