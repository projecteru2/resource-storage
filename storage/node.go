package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/projecteru2/resource-storage/storage/schedule"
	storagetypes "github.com/projecteru2/resource-storage/storage/types"

	"github.com/cockroachdb/errors"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	"github.com/sanity-io/litter"
)

func (p Plugin) AddNode(ctx context.Context, nodename string, resource plugintypes.NodeResourceRequest, info *enginetypes.Info) (resourcetypes.RawParams, error) {
	// try to get the node resource
	var err error
	if _, err = p.doGetNodeResourceInfo(ctx, nodename); err == nil {
		return nil, coretypes.ErrNodeExists
	}

	if !errors.Is(err, coretypes.ErrInvaildCount) {
		log.WithFunc("resource.storage.AddNode").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	req := &storagetypes.NodeResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}

	// set default value
	if info != nil && req.Storage == 0 {
		req.Storage = info.StorageTotal * rate / 10
	}

	nodeResourceInfo := &storagetypes.NodeResourceInfo{
		Capacity: &storagetypes.NodeResource{
			Volumes: req.Volumes,
			Storage: req.Storage,
			Disks:   req.Disks,
		},
	}

	if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
	}, nil
}

func (p Plugin) RemoveNode(ctx context.Context, nodename string) error {
	if _, err := p.store.Delete(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename)); err != nil {
		log.WithFunc("resource.storage.RemoveNode").WithField("node", nodename).Error(ctx, err, "faield to delete node")
		return err
	}
	return nil
}

func (p Plugin) GetNodesDeployCapacity(ctx context.Context, nodenames []string, resource plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.storage.GetNodesDeployCapacity")
	req := &storagetypes.WorkloadResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodesDeployCapacityMap := map[string]*plugintypes.NodeDeployCapacity{}
	total := 0
	for _, nodename := range nodenames {
		nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
		if err != nil {
			logger.WithField("node", nodename).Error(ctx, err)
			return nil, err
		}
		capacityInfo := p.doGetNodeDeployCapacity(nodeResourceInfo, req)
		if capacityInfo.Capacity > 0 {
			nodesDeployCapacityMap[nodename] = capacityInfo
			if total == math.MaxInt64 || capacityInfo.Capacity == math.MaxInt64 {
				total = math.MaxInt64
			} else {
				total += capacityInfo.Capacity
			}
		}
	}

	return resourcetypes.RawParams{
		"nodes_deploy_capacity_map": nodesDeployCapacityMap,
		"total":                     total,
	}, nil
}

func (p Plugin) SetNodeResourceCapacity(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, delta bool, incr bool) (resourcetypes.RawParams, error) {
	// TODO both nil should do nothing
	logger := log.WithFunc("resource.storage.SetNodeResourceCapacity").WithField("node", "nodename")
	req, nodeResource, _, nodeResourceInfo, err := p.parseNodeResourceInfos(ctx, nodename, resource, resourceRequest, nil)
	if err != nil {
		return nil, err
	}
	origin := nodeResourceInfo.Capacity
	before := origin.DeepCopy()

	if req != nil {
		if len(req.RMDisks) > 0 {
			if delta {
				return nil, coretypes.ErrInvalidEngineArgs
			}
			rmDisksMap := map[string]struct{}{}
			for _, rmDisk := range req.RMDisks {
				rmDisksMap[rmDisk] = struct{}{}
			}
			nodeResourceInfo.Capacity.Disks = utils.Filter(nodeResourceInfo.Capacity.Disks, func(d *storagetypes.Disk) bool {
				_, ok := rmDisksMap[d.Device]
				return !ok
			})
		}
		if !delta {
			req.SkipEmpty(nodeResourceInfo.Capacity)
		}
	}

	nodeResourceInfo.Capacity = p.calculateNodeResource(req, nodeResource, nodeResourceInfo.Capacity, nil, delta, incr)
	if delta {
		nodeResourceInfo.Capacity.RemoveEmpty()
	}

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Capacity,
	}, nil
}

func (p Plugin) GetNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, _, _, _, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource) // nolint
	if err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

func (p Plugin) SetNodeResourceInfo(ctx context.Context, nodename string, capacity plugintypes.NodeResource, usage plugintypes.NodeResource) error {
	capacityResource := &storagetypes.NodeResource{}
	usageResource := &storagetypes.NodeResource{}
	// TODO this api need capacity with data!!!!
	if err := capacityResource.Parse(capacity); err != nil {
		return err
	}
	if capacityResource.Volumes != nil {
		capacityResource.Storage += capacityResource.Volumes.Total()
	}
	if err := usageResource.Parse(usage); err != nil {
		return err
	}

	resourceInfo := &storagetypes.NodeResourceInfo{
		Capacity: capacityResource,
		Usage:    usageResource,
	}

	return p.doSetNodeResourceInfo(ctx, nodename, resourceInfo)
}

func (p Plugin) SetNodeResourceUsage(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, workloadsResource []plugintypes.WorkloadResource, delta bool, incr bool) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.storage.SetNodeResourceUsage").WithField("node", "nodename")
	req, nodeResource, wrksResource, nodeResourceInfo, err := p.parseNodeResourceInfos(ctx, nodename, resource, resourceRequest, workloadsResource)
	if err != nil {
		return nil, err
	}
	origin := nodeResourceInfo.Usage
	before := origin.DeepCopy()

	nodeResourceInfo.Usage = p.calculateNodeResource(req, nodeResource, nodeResourceInfo.Usage, wrksResource, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Usage,
	}, nil
}

func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (*resourcetypes.RawParams, error) {
	nodename := nodenames[0]
	return &resourcetypes.RawParams{
		"nodename": nodename,
		"priority": 1, // TODO why 1?
	}, nil
}

func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, totalVolumes, totalDiskUsage, totalStorageUsage, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	if len(diffs) != 0 {
		nodeResourceInfo.Usage = &storagetypes.NodeResource{
			Volumes: totalVolumes,
			Disks:   totalDiskUsage,
			Storage: totalStorageUsage,
		}
		if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
			log.WithFunc("resource.storage.FixNodeResource").Error(ctx, err)
			diffs = append(diffs, "fix failed")
		}
	}

	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

func (p Plugin) getNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (
	*storagetypes.NodeResourceInfo,
	storagetypes.Volumes, storagetypes.Disks, int64,
	[]string, error,
) {
	logger := log.WithFunc("resource.storage.getNodeResourceInfo").WithField("node", nodename)
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.Error(ctx, err)
		return nil, nil, nil, 0, nil, err
	}

	totalVolumes := storagetypes.Volumes{}
	totalDiskUsage := storagetypes.Disks{}
	totalStorageUsage := int64(0)
	for _, workloadResource := range workloadsResource {
		workloadUsage := &storagetypes.WorkloadResource{}
		if err := workloadUsage.Parse(workloadResource); err != nil {
			logger.Error(ctx, err)
			return nil, nil, nil, 0, nil, err
		}
		for _, volumeMap := range workloadUsage.VolumePlanRequest {
			totalVolumes.Add(volumeMap)
		}
		totalStorageUsage += workloadUsage.StorageRequest
		totalDiskUsage.Add(workloadUsage.DisksRequest.RemoveMounts())
	}

	diffs := []string{}

	if nodeResourceInfo.Usage.Storage != totalStorageUsage {
		diffs = append(diffs, fmt.Sprintf("node.Storage != sum(workload.Storage): %+v != %+v", nodeResourceInfo.Usage.Storage, totalStorageUsage))
	}
	for volume, size := range nodeResourceInfo.Usage.Volumes {
		if totalVolumes[volume] != size {
			diffs = append(diffs, fmt.Sprintf("node.Volumes[%s] != sum(workload.Volumes[%s]): %+v != %+v", volume, volume, size, totalVolumes[volume]))
		}
	}
	for volume, size := range totalVolumes {
		if vol, ok := nodeResourceInfo.Usage.Volumes[volume]; !ok && vol != size {
			diffs = append(diffs, fmt.Sprintf("node.Volumes[%s] != sum(workload.Volumes[%s]): %+v != %+v", volume, volume, nodeResourceInfo.Usage.Volumes[volume], size))
		}
	}
	for _, disk := range nodeResourceInfo.Usage.Disks {
		d := totalDiskUsage.GetDiskByDevice(disk.Device)
		if d == nil {
			d = &storagetypes.Disk{
				Device:    disk.Device,
				Mounts:    disk.Mounts,
				ReadIOPS:  0,
				WriteIOPS: 0,
				ReadBPS:   0,
				WriteBPS:  0,
			}
		}
		d.Mounts = disk.Mounts
		computedDisk := d.String()
		storedDisk := disk.String()
		if computedDisk != storedDisk {
			diffs = append(diffs, fmt.Sprintf("node.Disks[%s] != sum(workload.Disks[%s]): %+v != %+v", disk.Device, disk.Device, storedDisk, computedDisk))
		}
	}

	return nodeResourceInfo, totalVolumes, totalDiskUsage, totalStorageUsage, diffs, nil
}

func (p Plugin) doGetNodeResourceInfo(ctx context.Context, nodename string) (*storagetypes.NodeResourceInfo, error) {
	resourceInfo := &storagetypes.NodeResourceInfo{}
	resp, err := p.store.GetOne(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename))
	if err != nil {
		return nil, err
	}
	return resourceInfo, json.Unmarshal(resp.Value, resourceInfo)
}

func (p Plugin) doSetNodeResourceInfo(ctx context.Context, nodename string, resourceInfo *storagetypes.NodeResourceInfo) error {
	if err := resourceInfo.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(resourceInfo)
	if err != nil {
		return err
	}

	_, err = p.store.Put(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename), string(data))
	return err
}

func (p Plugin) doGetNodeDeployCapacity(nodeResourceInfo *storagetypes.NodeResourceInfo, req *storagetypes.WorkloadResourceRequest) *plugintypes.NodeDeployCapacity {
	capacityInfo := &plugintypes.NodeDeployCapacity{
		Weight: 1,
	}

	// get volume capacity
	volumePlans, _ := schedule.GetVolumePlans(nodeResourceInfo, req.VolumesRequest, p.config.Scheduler.MaxDeployCount)
	capacityInfo.Capacity = len(volumePlans)

	// get storage capacity
	if req.StorageRequest > 0 {
		storageCapacity := int((nodeResourceInfo.Capacity.Storage - nodeResourceInfo.Usage.Storage) / req.StorageRequest)
		if req.VolumesLimit == nil || (storageCapacity < capacityInfo.Capacity) {
			capacityInfo.Capacity = storageCapacity
		}
	}

	// get usage and rate
	if nodeResourceInfo.Capacity.Volumes.Total() == 0 && nodeResourceInfo.Capacity.Storage == 0 {
		return capacityInfo
	}

	if len(req.VolumesRequest) > 0 || req.StorageRequest == 0 {
		capacityInfo.Usage = utils.AdvancedDivide(float64(nodeResourceInfo.Usage.Volumes.Total()), float64(nodeResourceInfo.Capacity.Volumes.Total()))
		capacityInfo.Rate = utils.AdvancedDivide(float64(req.VolumesRequest.TotalSize()), float64(nodeResourceInfo.Capacity.Volumes.Total())) //
	} else if req.StorageRequest > 0 {
		capacityInfo.Usage = utils.AdvancedDivide(float64(nodeResourceInfo.Usage.Storage), float64(nodeResourceInfo.Capacity.Storage))
		capacityInfo.Rate = utils.AdvancedDivide(float64(req.StorageRequest), float64(nodeResourceInfo.Capacity.Storage))
	}

	return capacityInfo
}

// calculateNodeResource priority: node resource request > node resource > workload resource args list
func (p Plugin) calculateNodeResource(req *storagetypes.NodeResourceRequest, nodeResource *storagetypes.NodeResource, origin *storagetypes.NodeResource, workloadsResource []*storagetypes.WorkloadResource, delta bool, incr bool) *storagetypes.NodeResource {
	var resp *storagetypes.NodeResource
	if origin == nil || !delta { // no delta means node resource rewrite with whole new data
		resp = (&storagetypes.NodeResource{}).DeepCopy()
		// 这个接口最诡异的在于，如果 delta 为 false，意味着是全量写入
		// 但这时候 incr 为 false 的话
		// 实际上是 set 进了负值，所以这里 incr 应该强制为 true
		incr = true
	} else {
		resp = origin.DeepCopy()
	}

	if req != nil {
		nodeResource = &storagetypes.NodeResource{
			Volumes: req.Volumes,
			Storage: req.Storage,
			Disks:   req.Disks,
		}
	}

	if nodeResource != nil {
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &storagetypes.NodeResource{
			Volumes: map[string]int64{},
			Storage: workloadResource.StorageRequest,
			Disks:   workloadResource.DisksRequest,
		}
		for _, volumeMap := range workloadResource.VolumePlanRequest {
			nodeResource.Volumes.Add(volumeMap)
		}
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
	}
	return resp
}

func (p Plugin) parseNodeResourceInfos(
	ctx context.Context, nodename string,
	resource plugintypes.NodeResource,
	resourceRequest plugintypes.NodeResourceRequest,
	workloadsResource []plugintypes.WorkloadResource,
) (
	*storagetypes.NodeResourceRequest,
	*storagetypes.NodeResource,
	[]*storagetypes.WorkloadResource,
	*storagetypes.NodeResourceInfo,
	error,
) {
	var req *storagetypes.NodeResourceRequest
	var nodeResource *storagetypes.NodeResource
	wrksResource := []*storagetypes.WorkloadResource{}

	if resourceRequest != nil {
		req = &storagetypes.NodeResourceRequest{}
		if err := req.Parse(resourceRequest); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if resource != nil {
		nodeResource = &storagetypes.NodeResource{}
		if err := nodeResource.Parse(resource); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	for _, workloadResource := range workloadsResource {
		wrkResource := &storagetypes.WorkloadResource{}
		if err := wrkResource.Parse(workloadResource); err != nil {
			return nil, nil, nil, nil, err
		}
		wrksResource = append(wrksResource, wrkResource)
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return req, nodeResource, wrksResource, nodeResourceInfo, nil
}
