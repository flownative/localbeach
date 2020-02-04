// +build ignore

package main

import (
	"log"

	"github.com/flownative/localbeach/assets"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(assets.Assets, vfsgen.Options{
		PackageName:  "assets",
		BuildTags:    "!dev",
		VariableName: "Assets",
		Filename:     "assets/compiled.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
