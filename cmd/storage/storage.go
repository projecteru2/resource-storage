package storage

import (
	"github/projecteru2/resource-storage/cmd"

	"github.com/projecteru2/core/resource3/plugins"
	cli "github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "name",
		Usage:  "show name",
		Action: name,
	}
}

func name(c *cli.Context) error {
	return cmd.Serve(c, func(p plugins.Plugin) error {
		print(`{"name": "` + p.Name() + `"}`)
		return nil
	})
}
