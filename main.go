package main

import (
	"log"
	"os"

	"staticinsighter/analyzer"
)

func main() {
	path := os.Getenv("TARGET_PATH")

	analyzer.SetLogLevel(analyzer.DebugLevel)
	if err := analyzer.Analyze(path); err != nil {
		log.Fatalf("analyze fail: %s", err)
	}
}
