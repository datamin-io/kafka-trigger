package cli

import (
	"kafkatrigger/cli/command"

	"github.com/urfave/cli/v2"
)

func NewApplication() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			command.RunCommand,
		},
	}
}
