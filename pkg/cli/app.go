package cli

import (
	"integration/pkg/cli/command"

	"github.com/urfave/cli/v2"
)

func NewApplication() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			command.RunKafkaTriggerCommand,
			command.RunS3ListenerLambdaCommand,
		},
	}
}
