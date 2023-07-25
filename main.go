package main

import (
	"flag"
	"log"

	"staticinsighter/analyzer"
)

func main() {
	// path := os.Getenv("TARGET_PATH")
	var path = flag.String("path", "", "specify target path")
	flag.Parse()

	analyzer.SetLogLevel(analyzer.DebugLevel)
	if err := analyzer.Analyze(*path); err != nil {
		log.Fatalf("analyze fail: %s", err)
	}
}
