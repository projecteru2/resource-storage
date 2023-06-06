package storage

import (
	"context"
	"fmt"
	"strings"

	resourcetypes "github.com/projecteru2/core/resource/types"
)

// GetMetricsDescription .
func (p Plugin) GetMetricsDescription(_ context.Context) ([]resourcetypes.RawParams, error) {
	return []resourcetypes.RawParams{
		{
			"name":   "storage_used",
			"help":   "node used storage.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "storage_capacity",
			"help":   "node available storage.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
	}, nil
}

// GetMetrics .
func (p Plugin) GetMetrics(ctx context.Context, podname, nodename string) ([]resourcetypes.RawParams, error) {
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}
	safeNodename := strings.ReplaceAll(nodename, ".", "_")
	return []resourcetypes.RawParams{
		{
			"name":   "storage_used",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.Usage.Storage),
			"key":    fmt.Sprintf("core.node.%s.storage.used", safeNodename),
		},
		{
			"name":   "storage_capacity",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.Capacity.Storage),
			"key":    fmt.Sprintf("core.node.%s.storage.used", safeNodename),
		},
	}, nil
}
