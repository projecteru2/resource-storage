package metrics

import (
	"fmt"
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"

	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/urfave/cli/v2"
)

func DescriptionCommand() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMetricsDescriptionCommand,
		Usage:  "show metrics descriptions",
		Action: serve,
	}
}

func serve(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin) error {
		r, err := s.GetMetricsDescription(c.Context)
		if err != nil {
			return err
		}
		fmt.Print(string(r))
		return nil
	})
}
