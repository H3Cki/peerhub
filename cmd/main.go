package main

import (
	"os"

	"github.com/H3Cki/peerhub/cmd/commands/websocketcmd"
	"github.com/urfave/cli/v2"
)

const (
	appName = "peerhub"
	version = "v0.1.0"
)

func main() {
	app := cli.App{
		Name:        appName,
		Version:     version,
		Description: "peerhub is a server for exchanging signals between webrtc peers",
		Commands: []*cli.Command{
			websocketcmd.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
