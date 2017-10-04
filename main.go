package main

import (
	"os"

	"github.com/itkq/kinesis-streams-agent/cli"
)

func main() {
	os.Exit(cli.StartCLI())
}
