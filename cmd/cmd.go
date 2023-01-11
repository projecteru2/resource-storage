package cmd

import (
	"github/projecteru2/resource-storage/storage"
	"testing"

	"github.com/projecteru2/core/utils"
	"github.com/urfave/cli/v2"
)

var (
	ConfigPath      string
	EmbeddedStorage bool
)

func Serve(c *cli.Context, f func(s *storage.Plugin) error) error {
	config, err := utils.LoadConfig(ConfigPath)
	if err != nil {
		return err
	}

	var t *testing.T
	if EmbeddedStorage {
		t = &testing.T{}
	}

	s, err := storage.NewPlugin(c.Context, config, t)
	if err != nil {
		return err
	}
	return f(s)
}
