package main

import "github.com/flownative/localbeach/cmd/beach/cmd"

//go:generate go run -tags=dev assets_generate.go

func main() {
	cmd.Execute()
}