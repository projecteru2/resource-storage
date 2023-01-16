package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-units"
	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource3/plugins/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/resource-storage/storage/types"
	storagetypes "github.com/projecteru2/resource-storage/storage/types"
	"github.com/stretchr/testify/assert"
)

func TestAddNode(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]
	nodeForAdd := "test2"

	req := &plugintypes.NodeResourceRequest{
		"volumes": vols,
	}
	info := &enginetypes.Info{StorageTotal: units.TB}

	// existent node
	_, err := st.AddNode(ctx, node, req, info)
	assert.Equal(t, err, coretypes.ErrNodeExists)

	// normal case
	r, err := st.AddNode(ctx, nodeForAdd, req, info)
	assert.Nil(t, err)
	ni, ok := (*r)["capacity"].(*storagetypes.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, ni.Storage, int64(units.TiB))
}

func TestRemoveNode(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]
	nodeForDel := "test2"

	err := st.RemoveNode(ctx, node)
	assert.Nil(t, err)
	err = st.RemoveNode(ctx, nodeForDel)
	assert.Nil(t, err)
}

func TestGetNodesDeployCapacity(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 10, vols, 0)

	// invalid request
	_, err := st.GetNodesDeployCapacity(ctx, nodes, &plugintypes.WorkloadResourceRequest{"storage": "-1"})
	assert.ErrorIs(t, err, types.ErrInvalidStorage)

	// invalid node
	req := &plugintypes.WorkloadResourceRequest{"storage": "1"}
	_, err = st.GetNodesDeployCapacity(ctx, []string{"??"}, req)
	assert.ErrorIs(t, err, coretypes.ErrInvaildCount)

	// no volume request
	req = &plugintypes.WorkloadResourceRequest{"storage": fmt.Sprintf("%v", units.TiB)}
	r, err := st.GetNodesDeployCapacity(ctx, nodes, req)
	assert.NoError(t, err)
	assert.Equal(t, (*r)["total"], 40)

	// no stroage request
	req = &plugintypes.WorkloadResourceRequest{
		"volumes": []string{"AUTO:/dir0:rwm:1G"},
	}
	r, err = st.GetNodesDeployCapacity(ctx, nodes, req)
	assert.NoError(t, err)
	assert.Equal(t, (*r)["total"], 40)

	// mixed
	req = &plugintypes.WorkloadResourceRequest{
		"volumes": []string{"AUTO:/dir0:rwm:1G"},
		"storage": fmt.Sprintf("%v", units.TiB),
	}
	r, err = st.GetNodesDeployCapacity(ctx, nodes, req)
	assert.NoError(t, err)
	assert.Equal(t, (*r)["total"], 30)
}
