package node

import (
	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"
	"github.com/urfave/cli/v2"
)

func GetNodeResourceInfo() *cli.Command {
	return &cli.Command{
		Name:   binary.GetNodeResourceInfoCommand,
		Usage:  "get node resource info",
		Action: getNodeResourceInfo,
	}
}

func getNodeResourceInfo(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadsResource := in.SliceRawParams("workloads_resource")
		return s.GetNodeResourceInfo(c.Context, nodename, workloadsResource)
	})
}
