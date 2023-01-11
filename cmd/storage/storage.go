package storage

import (
	"fmt"
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "name",
		Usage:  "show name",
		Action: serve,
	}
}

func serve(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin) error {
		fmt.Print(s.Name())
		return nil
	})
}
