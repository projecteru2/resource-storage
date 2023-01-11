package metrics

import (
	"fmt"
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"

	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func DescriptionCommand() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMetricsDescriptionCommand,
		Usage:  "show metrics descriptions",
		Action: description,
	}
}

func description(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, _ *types.RawParams) error {
		r, err := s.GetMetricsDescription(c.Context)
		if err != nil {
			return err
		}
		fmt.Print(string(r))
		return nil
	})
}
