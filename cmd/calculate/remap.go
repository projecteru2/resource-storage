package calculate

import (
	"github.com/mitchellh/mapstructure"
	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"
	"github.com/urfave/cli/v2"
)

func CalculateRemap() *cli.Command {
	return &cli.Command{
		Name:   binary.CalculateRemapCommand,
		Usage:  "remap resource",
		Action: calculateRemap,
	}
}

func calculateRemap(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}

		workloadsResource := map[string]*coretypes.RawParams{}
		for ID, data := range *in.RawParams("workloads_resource") {
			workloadsResource[ID] = &coretypes.RawParams{}
			_ = mapstructure.Decode(data, workloadsResource[ID])
		}
		// NO NEED REMAP VOLUME
		return s.CalculateRemap(c.Context, nodename, workloadsResource)
	})
}
