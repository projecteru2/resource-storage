package storage

import (
	"fmt"

	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"

	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func Name() *cli.Command {
	return &cli.Command{
		Name:   "name",
		Usage:  "show name",
		Action: name,
	}
}

func name(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, _ *types.RawParams) (interface{}, error) {
		fmt.Print(s.Name())
		return nil, nil
	})
}
