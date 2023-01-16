package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-units"
	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource3/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	st := initStorage(context.Background(), t)
	assert.Contains(t, st.Name(), st.name)
}

func initStorage(ctx context.Context, t *testing.T) *Plugin {
	config := coretypes.Config{
		Etcd: coretypes.EtcdConfig{
			Prefix: "/storage",
		},
		Scheduler: coretypes.SchedulerConfig{
			MaxShare:       -1,
			ShareBase:      100,
			MaxDeployCount: 100,
		},
	}

	st, err := NewPlugin(ctx, config, t)
	assert.NoError(t, err)
	return st
}

func generateNodes(
	ctx context.Context, t *testing.T, st *Plugin, nums int, vols []string, index int,
) []string {
	reqs := generateNodeResourceRequests(t, nums, vols, index)
	info := &enginetypes.Info{StorageTotal: units.TB}
	names := []string{}
	for name, req := range reqs {
		_, err := st.AddNode(ctx, name, req, info)
		assert.NoError(t, err)
		names = append(names, name)
	}
	t.Cleanup(func() {
		for name := range reqs {
			st.RemoveNode(ctx, name)
		}
	})
	return names
}

func generateNodeResourceRequests(t *testing.T, nums int, vols []string, index int) map[string]*plugintypes.NodeResourceRequest {
	infos := map[string]*plugintypes.NodeResourceRequest{}
	for i := index; i < index+nums; i++ {
		info := &plugintypes.NodeResourceRequest{
			"volumes": vols,
		}
		infos[fmt.Sprintf("test%v", i)] = info
	}
	return infos
}
