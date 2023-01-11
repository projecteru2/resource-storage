package metrics

import (
	"encoding/json"
	"fmt"
	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/storage"
	"os"

	"github.com/projecteru2/core/resource3/plugins/binary"
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
	return cmd.Serve(c, func(s *storage.Plugin) error {
		req := map[string]string{}
		if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
			return err
		}
		r, err := s.GetMetrics(c.Context, req["podname"], req["nodename"])
		if err != nil {
			return err
		}
		fmt.Print(string(r))
		return nil
	})
}
