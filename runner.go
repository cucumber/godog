package godog

import (
	"flag"
	"log"
)

func Run() {
	if !flag.Parsed() {
		flag.Parse()
	}
	log.Println("running godoc, num registered steps:", len(stepHandlers), "color test:", cl("red", red))
	log.Println("will read features in path:", cl(cfg.featuresPath, yellow))
}
