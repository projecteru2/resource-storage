package main

import (
	"fmt"
	"os"

	"github/projecteru2/resource-storage/cmd"
	"github/projecteru2/resource-storage/cmd/metrics"
	"github/projecteru2/resource-storage/cmd/storage"
	"github/projecteru2/resource-storage/version"

	"github.com/urfave/cli/v2"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.String())
	}

	app := cli.NewApp()
	app.Name = version.NAME
	app.Usage = "Run eru resource storage plugin"
	app.Version = version.VERSION
	app.Commands = []*cli.Command{
		storage.Command(),
		metrics.DescriptionCommand(),
		metrics.GetMetricsCommand(),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Value:       "storage.yaml",
			Usage:       "config file path for plugin, in yaml",
			Destination: &cmd.ConfigPath,
			EnvVars:     []string{"ERU_RESOURCE_CONFIG_PATH"},
		},
		&cli.BoolFlag{
			Name:        "embedded-storage",
			Usage:       "active embedded storage",
			Destination: &cmd.EmbeddedStorage,
		},
	}
	_ = app.Run(os.Args)
}
