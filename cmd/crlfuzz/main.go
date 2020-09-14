package main

import "github.com/dwisiswant0/crlfuzz/internal/runner"

func main() {
	options := runner.ParseOptions()
	runner.New(options)
}
