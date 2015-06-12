package main

import (
	"log"
	"regexp"

	"github.com/DATA-DOG/godog"
)

func SomeContext(g godog.Suite) {
	f := godog.StepHandlerFunc(func(args ...interface{}) error {
		log.Println("step triggered")
		return nil
	})
	g.Step(regexp.MustCompile("hello"), f)
}
