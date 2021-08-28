package main

import (
	"os"

	"github.com/arcology/3rd-party/tm/cli"
	"github.com/arcology/pool-svc/service"
)

func main() {
	st := service.StartCmd

	cmd := cli.PrepareMainCmd(st, "BC", os.ExpandEnv("$HOME/monacos/pool"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
