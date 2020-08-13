package main

import "dw1.io/crlfuzz/internal/runner"

func main() {
	options := runner.ParseOptions()
	runner.New(options)
}
