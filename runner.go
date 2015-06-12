package godog

import (
	"fmt"
	"log"
)

func Run() error {
	log.Println("running godoc, num registered steps:", len(stepHandlers), "color test:", red("red"))
	return nil
}

func red(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}
