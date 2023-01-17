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
	ni, ok := (*r)["capacity"].(*types.NodeResource)
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

func TestSetNodeResourceCapacity(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]

	r, err := st.GetNodeResourceInfo(ctx, nodes[0], nil)
	assert.Nil(t, err)
	v, ok := (*r)["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(4*units.TiB))

	resourceRequest := &plugintypes.NodeResourceRequest{
		"volumes": []string{"/data4:1T"},
		"storage": "1T",
	}

	nodeResource := &plugintypes.NodeResource{
		"volumes": types.VolumeMap{"/data4": units.TiB},
		"storage": units.TiB,
	}

	d, err := st.SetNodeResourceCapacity(ctx, node, nodeResource, nil, true, true)
	assert.NoError(t, err)
	v, ok = (*d)["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(5*units.TiB))

	d, err = st.SetNodeResourceCapacity(ctx, node, nodeResource, nil, true, false)
	assert.NoError(t, err)
	v, ok = (*d)["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(4*units.TiB))

	d, err = st.SetNodeResourceCapacity(ctx, node, nil, resourceRequest, true, true)
	assert.NoError(t, err)
	v, ok = (*d)["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(6*units.TiB))

	d, err = st.SetNodeResourceCapacity(ctx, node, nil, resourceRequest, true, false)
	assert.NoError(t, err)
	v, ok = (*d)["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(4*units.TiB))

	d, err = st.SetNodeResourceCapacity(ctx, node, nil, resourceRequest, false, false)
	assert.NoError(t, err)
	v, ok = (*d)["after"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(2*units.TiB))
}

func TestGetNodeResourceInfo(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]

	// invalid node
	_, err := st.GetNodeResourceInfo(ctx, "abc", nil)
	assert.Error(t, err)

	// normal case
	d, err := st.GetNodeResourceInfo(ctx, node, nil)
	assert.NoError(t, err)
	v, ok := (*d)["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(4*units.TiB))

	// diffs
	workloadsResource := []*plugintypes.WorkloadResource{
		{"storage_request": 1},
		{"storage_limit": 1},
	}

	d, err = st.GetNodeResourceInfo(ctx, node, nil)
	assert.NoError(t, err)
	v, ok = (*d)["capacity"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(4*units.TiB))

	d, err = st.GetNodeResourceInfo(ctx, node, workloadsResource)
	v2, ok := (*d)["diffs"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, v2)
}

func TestFixNodeResource(t *testing.T) {
	ctx := context.Background()
	st := initStorage(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]

	// invalid node
	_, err := st.FixNodeResource(ctx, "abc", nil)
	assert.Error(t, err)

	// normal case
	workloadsResource := []*plugintypes.WorkloadResource{
		{"storage_request": 1},
		{"storage_limit": 1},
	}

	d, err := st.FixNodeResource(ctx, node, workloadsResource)
	assert.NoError(t, err)
	v2, ok := (*d)["diffs"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, v2)

	d, err = st.GetNodeResourceInfo(ctx, node, nil)
	assert.NoError(t, err)
	v, ok := (*d)["usage"].(*types.NodeResource)
	assert.True(t, ok)
	assert.Equal(t, v.Storage, int64(1))
}
