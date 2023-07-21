package main

import (
	"os"

	"staticinsighter/analyzer"
)

func main() {
	path := os.Getenv("TARGET_PATH")

	analyzer.Analyze(path)
}
