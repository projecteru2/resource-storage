package storage

import (
	"fmt"

	"github.com/projecteru2/resource-storage/storage/types"
)

func (p Plugin) toIOPSOptions(disks types.Disks) map[string]string {
	IOPSOptions := map[string]string{}
	for _, disk := range disks {
		IOPSOptions[disk.Device] = fmt.Sprintf("%d:%d:%d:%d", disk.ReadIOPS, disk.WriteIOPS, disk.ReadBPS, disk.WriteBPS)
	}
	return IOPSOptions
}

func getVolumePlanLimit(volumeRequest types.VolumeBindings, volumeLimit types.VolumeBindings, volumePlan types.VolumePlan) types.VolumePlan {
	volumePlanLimit := types.VolumePlan{}

	volumeBindingToVolumes := map[[3]string]types.Volumes{}
	for binding, volumeMap := range volumePlan {
		volumeBindingToVolumes[binding.GetMapKey()] = volumeMap
	}

	for index, binding := range volumeLimit {
		if !binding.RequireSchedule() {
			continue
		}
		if volumeMap, ok := volumeBindingToVolumes[binding.GetMapKey()]; ok {
			volumePlanLimit[binding] = types.Volumes{volumeMap.GetDevice(): volumeMap.GetSize() + binding.SizeInBytes - volumeRequest[index].SizeInBytes}
		}
	}
	return volumePlanLimit
}

func getDisksLimit(volumeLimit types.VolumeBindings, volumePlanLimit types.VolumePlan, disks types.Disks) types.Disks {
	disksLimit := types.Disks{}
	for _, binding := range volumeLimit {
		if binding.RequireIOPS() && !binding.RequireSchedule() {
			disk := disks.GetDiskByPath(binding.Source)
			disksLimit.Add(types.Disks{&types.Disk{
				Device:    disk.Device,
				Mounts:    disk.Mounts,
				ReadIOPS:  binding.ReadIOPS,
				WriteIOPS: binding.WriteIOPS,
				ReadBPS:   binding.ReadBPS,
				WriteBPS:  binding.WriteBPS,
			}})
		}
	}
	for binding, volumeMap := range volumePlanLimit {
		if !binding.RequireIOPS() {
			continue
		}
		disk := disks.GetDiskByPath(volumeMap.GetDevice())
		disksLimit.Add(types.Disks{&types.Disk{
			Device:    disk.Device,
			Mounts:    disk.Mounts,
			ReadIOPS:  binding.ReadIOPS,
			WriteIOPS: binding.WriteIOPS,
			ReadBPS:   binding.ReadBPS,
			WriteBPS:  binding.WriteBPS,
		}})
	}
	return disksLimit
}

func getDeltaWorkloadResourceArgs(originResource, targetWorkloadResource *types.WorkloadResource) *types.WorkloadResource {
	deltaVolumes := types.Volumes{}
	for _, volumeMap := range targetWorkloadResource.VolumePlanRequest {
		deltaVolumes.Add(volumeMap)
	}
	for _, volumeMap := range originResource.VolumePlanRequest {
		deltaVolumes.Sub(volumeMap)
	}

	deltaDisks := targetWorkloadResource.DisksRequest.DeepCopy()
	deltaDisks.Sub(originResource.DisksRequest)

	return &types.WorkloadResource{
		VolumePlanRequest: types.VolumePlan{&types.VolumeBinding{
			Source:      "fake-source",
			Destination: "fake-destination",
			Flags:       "fake-flags",
			SizeInBytes: 0,
		}: deltaVolumes},
		StorageRequest: targetWorkloadResource.StorageRequest - originResource.StorageRequest,
		DisksRequest:   deltaDisks,
	}
}
