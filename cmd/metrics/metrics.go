package metrics

import (
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"

	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v2"
)

func GetMetrics() *cli.Command {
	return &cli.Command{
		Name:   binary.GetMetricsCommand,
		Usage:  "show metrics",
		Action: metric,
	}
}

func metric(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		podname := in.String("podname")
		nodename := in.String("nodename")
		return s.GetMetrics(c.Context, podname, nodename)
	})
}
