package main

import (
	_ "embed"
	"strings"

	"github.com/karanshah229/gistsync/cmd"
)

//go:embed VERSION
var version string

func main() {
	cmd.SetVersion(strings.TrimSpace(version))
	cmd.Execute()
}
