package calculate

import (
	"github.com/projecteru2/core/resource3/plugins/binary"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/cmd"
	"github.com/projecteru2/resource-storage/storage"
	"github.com/urfave/cli/v2"
)

func CalculateDeploy() *cli.Command {
	return &cli.Command{
		Name:   binary.CalculateDeployCommand,
		Usage:  "calculate deploy pan",
		Action: calculateDeploy,
	}
}

func calculateDeploy(c *cli.Context) error {
	return cmd.Serve(c, func(s *storage.Plugin, in *types.RawParams) (interface{}, error) {
		nodename := in.String("nodename")
		if nodename == "" {
			return nil, types.ErrEmptyNodeName
		}
		deployCount := in.Int("deploy_count")

		workloadResourceRequest := in.RawParams("workload_resource_request")
		return s.CalculateDeploy(c.Context, nodename, deployCount, workloadResourceRequest)
	})
}
