package metrics

import (
	"fmt"
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"

	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func GetMetricsCommand() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMetricsCommand,
		Usage:  "show metrics",
		Action: metric,
	}
}

func metric(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) error {
		podname := in.String("podname")
		nodename := in.String("nodename")
		r, err := s.GetMetrics(c.Context, podname, nodename)
		if err != nil {
			return err
		}
		fmt.Print(string(r))
		return nil
	})
}
