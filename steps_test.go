package godog

import (
	"log"
	"regexp"
)

func SomeContext(g Suite) {
	f := StepHandlerFunc(func(args ...interface{}) error {
		log.Println("step triggered")
		return nil
	})
	g.Step(regexp.MustCompile("hello"), f)
}
