package main

import (
	"flag"

	"github.com/riverchu/pkg/log"

	"staticinsighter/analyzer"
)

func main() {
	// path := os.Getenv("TARGET_PATH")
	var path = flag.String("path", "", "specify target path")
	var entry = flag.String("entry", "", "specify entry of project")
	flag.Parse()

	defer log.Flush()
	analyzer.SetMode(analyzer.ModeDebug)
	if err := analyzer.Analyze(*path, *entry); err != nil {
		log.Fatal("analyze fail: %s", err)
	}
}
