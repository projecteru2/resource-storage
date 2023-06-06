package calculate

import (
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"
	"github.com/urfave/cli/v2"
)

func CalculateRealloc() *cli.Command { // nolint
	return &cli.Command{
		Name:   binary.CalculateReallocCommand,
		Usage:  "calculate realloc plan",
		Action: calculateRealloc,
	}
}

func calculateRealloc(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadResource := in.RawParams("workload_resource")
		workloadResourceRequest := in.RawParams("workload_resource_request")

		return s.CalculateRealloc(c.Context, nodename, workloadResource, workloadResourceRequest)
	})
}
