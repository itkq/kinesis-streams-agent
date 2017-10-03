package main

import (
	"os"

	"github.com/itkq/kinesis-agent-go/cli"
)

func main() {
	os.Exit(cli.StartCLI())
}
