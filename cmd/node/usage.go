package node

import (
	"github.com/projecteru2/core/resource/plugins/binary"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"
	"github.com/urfave/cli/v2"
)

func SetNodeResourceUsage() *cli.Command {
	return &cli.Command{
		Name:   binary.SetNodeResourceUsageCommand,
		Usage:  "set node usage",
		Action: setNodeResourceUsage,
	}
}

func setNodeResourceUsage(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in resourcetypes.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		incr := in.Bool("incr")
		delta := in.Bool("delta")
		resource := in.RawParams("resource")
		resourceRequest := in.RawParams("resource_request")
		workloadsResource := in.SliceRawParams("workloads_resource")
		return s.SetNodeResourceUsage(c.Context, nodename, resource, resourceRequest, workloadsResource, delta, incr)
	})
}
