package node

import (
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"

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
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) error {
		return nil
	})
}

func removeNode(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) error {
		nodename := in.String("nodename")
		if nodename == "" {
			return types.ErrEmptyNodeName
		}
		return s.RemoveNode(c.Context, nodename)
	})
}
