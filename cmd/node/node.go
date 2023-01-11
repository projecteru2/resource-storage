package node

import (
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"

	"github.com/mitchellh/mapstructure"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	"github.com/urfave/cli/v2"
)

func AddNodeCommand() *cli.Command {
	return &cli.Command{
		Name:   binary.AddNodeCommand,
		Usage:  "add node",
		Action: addNode,
	}
}

func RemoveNodeCommand() *cli.Command {
	return &cli.Command{
		Name:   binary.RemoveNodeCommand,
		Usage:  "remove node",
		Action: removeNode,
	}
}

func addNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		engineInfo := in.RawParams("info")
		resource := in.RawParams("resource")
		info := &enginetypes.Info{}
		if err := mapstructure.Decode(engineInfo, info); err != nil {
			return nil, err
		}
		return s.AddNode(c.Context, nodename, resource, info)

	})
}

func removeNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		return nil, s.RemoveNode(c.Context, nodename)
	})
}
