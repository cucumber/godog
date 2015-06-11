package godog

import "log"

func Run() {
	log.Println("running godoc, num registered steps:", len(stepHandlers))
}
