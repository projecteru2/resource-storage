package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/projecteru2/core/utils"
)

const auto = "AUTO"

// VolumeBinding src:dst[:flags][:size][:read_IOPS:write_IOPS:read_bytes:write_bytes]
type VolumeBinding struct {
	Source      string `json:"source" mapstructure:"source"`
	Destination string `json:"destination" mapstructure:"destination"`
	Flags       string `json:"flags" mapstructure:"flags"`
	SizeInBytes int64  `json:"size_in_bytes" mapstructure:"size_in_bytes"`
	ReadIOPS    int64  `json:"read_iops" mapstructure:"read_iops"`
	WriteIOPS   int64  `json:"write_iops" mapstructure:"write_iops"`
	ReadBPS     int64  `json:"read_bps" mapstructure:"read_bps"`
	WriteBPS    int64  `json:"write_bps" mapstructure:"write_bps"`
}

// NewVolumeBinding returns pointer of VolumeBinding
func NewVolumeBinding(volume string) (_ *VolumeBinding, err error) {
	var src, dst, flags string
	var size, readIOPS, writeIOPS, readBPS, writeBPS int64

	parts := strings.Split(volume, ":")
	if len(parts) > 8 || len(parts) < 2 {
		return nil, errors.Wrap(ErrInvalidVolume, volume)
	}
	if len(parts) == 2 {
		parts = append(parts, "rw")
	}
	for len(parts) < 8 {
		parts = append(parts, "0")
	}
	src = parts[0]
	dst = parts[1]
	flags = parts[2]

	ptrs := []*int64{&size, &readIOPS, &writeIOPS, &readBPS, &writeBPS}
	for i, ptr := range ptrs {
		value, err := utils.ParseRAMInHuman(parts[i+3])
		if err != nil {
			return nil, err
		}
		*ptr = value
	}

	flagParts := strings.Split(flags, "")
	sort.Strings(flagParts)

	vb := &VolumeBinding{
		Source:      src,
		Destination: dst,
		Flags:       strings.Join(flagParts, ""),
		SizeInBytes: size,
		ReadIOPS:    readIOPS,
		WriteIOPS:   writeIOPS,
		ReadBPS:     readBPS,
		WriteBPS:    writeBPS,
	}

	if vb.Flags == "" {
		vb.Flags = "rw"
	}

	return vb, vb.Validate()
}

// Validate return error if invalid
func (vb VolumeBinding) Validate() error {
	if vb.Destination == "" {
		return errors.Wrapf(ErrInvalidVolume, "dest must be provided: %+v", vb)
	}
	return nil
}

// RequireSchedule returns true if volume binding requires schedule
func (vb VolumeBinding) RequireSchedule() bool {
	return strings.HasSuffix(vb.Source, auto) || vb.Source == ""
}

// RequireScheduleUnlimitedQuota .
func (vb VolumeBinding) RequireScheduleUnlimitedQuota() bool {
	return vb.RequireSchedule() && vb.SizeInBytes == 0 && vb.ReadIOPS == 0 && vb.WriteIOPS == 0 && vb.ReadBPS == 0 && vb.WriteBPS == 0
}

// RequireScheduleMonopoly returns true if volume binding requires monopoly schedule
func (vb VolumeBinding) RequireScheduleMonopoly() bool {
	return vb.RequireSchedule() && strings.Contains(vb.Flags, "m")
}

// RequireIOPS returns true if volume binding requires IOPS / BPS
func (vb VolumeBinding) RequireIOPS() bool {
	return vb.ReadIOPS > 0 || vb.WriteIOPS > 0 || vb.ReadBPS > 0 || vb.WriteBPS > 0
}

// ValidIOParameters returns true if all io related parameters are valid
func (vb VolumeBinding) ValidIOParameters() bool {
	return vb.ReadIOPS >= 0 && vb.WriteIOPS >= 0 && vb.ReadBPS >= 0 && vb.WriteBPS >= 0
}

// ToString returns volume string
func (vb VolumeBinding) ToString(normalize bool) (volume string) {
	flags := vb.Flags
	if normalize {
		flags = strings.ReplaceAll(flags, "m", "")
	}

	if strings.Contains(flags, "o") {
		flags = strings.ReplaceAll(flags, "o", "")
		flags = strings.ReplaceAll(flags, "r", "ro")
		flags = strings.ReplaceAll(flags, "w", "wo")
	}

	if !normalize {
		volume = fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d:%d", vb.Source, vb.Destination, flags, vb.SizeInBytes, vb.ReadIOPS, vb.WriteIOPS, vb.ReadBPS, vb.WriteBPS)
	} else {
		switch {
		case vb.Flags == "" && vb.SizeInBytes == 0:
			volume = fmt.Sprintf("%s:%s", vb.Source, vb.Destination)
		case vb.ReadIOPS != 0 || vb.WriteIOPS != 0 || vb.ReadBPS != 0 || vb.WriteBPS != 0:
			volume = fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d:%d", vb.Source, vb.Destination, flags, vb.SizeInBytes, vb.ReadIOPS, vb.WriteIOPS, vb.ReadBPS, vb.WriteBPS)
		default:
			volume = fmt.Sprintf("%s:%s:%s:%d", vb.Source, vb.Destination, flags, vb.SizeInBytes)
		}
	}
	return volume
}

// GetMapKey .
func (vb VolumeBinding) GetMapKey() [3]string {
	return [3]string{vb.Source, vb.Destination, vb.Flags}
}

// DeepCopy .
func (vb VolumeBinding) DeepCopy() *VolumeBinding {
	return &VolumeBinding{
		Source:      vb.Source,
		Destination: vb.Destination,
		Flags:       vb.Flags,
		SizeInBytes: vb.SizeInBytes,
		ReadIOPS:    vb.ReadIOPS,
		WriteIOPS:   vb.WriteIOPS,
		ReadBPS:     vb.ReadBPS,
		WriteBPS:    vb.WriteBPS,
	}
}

// VolumeBindings is a collection of VolumeBinding
type VolumeBindings []*VolumeBinding

// NewVolumeBindings return VolumeBindings of reference type
func NewVolumeBindings(volumes []string) (volumeBindings VolumeBindings, err error) {
	for _, vb := range volumes {
		volumeBinding, err := NewVolumeBinding(vb)
		if err != nil {
			return nil, err
		}
		volumeBindings = append(volumeBindings, volumeBinding)
	}
	return
}

// Validate .
func (vbs VolumeBindings) Validate() error {
	for _, vb := range vbs {
		if vb.RequireScheduleMonopoly() && vb.RequireScheduleUnlimitedQuota() {
			return errors.Wrapf(ErrInvalidVolume, "monopoly volume can't be unlimited: %+v", vb)
		}
		if !vb.ValidIOParameters() {
			return errors.Wrapf(ErrInvalidParams, "%+v", vb)
		}
	}
	return nil
}

// UnmarshalJSON is used for encoding/json.Unmarshal
func (vbs *VolumeBindings) UnmarshalJSON(b []byte) (err error) {
	volumes := []string{}
	if err = json.Unmarshal(b, &volumes); err != nil {
		return err
	}
	*vbs, err = NewVolumeBindings(volumes)
	return
}

// MarshalJSON is used for encoding/json.Marshal
func (vbs VolumeBindings) MarshalJSON() ([]byte, error) {
	volumes := []string{}
	for _, vb := range vbs {
		volumes = append(volumes, vb.ToString(false))
	}
	bs, err := json.Marshal(volumes)
	return bs, err
}

func (vbs VolumeBindings) String() string {
	volumes := []string{}
	for _, vb := range vbs {
		volumes = append(volumes, vb.ToString(false))
	}
	return strings.Join(volumes, ",")
}

// TotalSize .
func (vbs VolumeBindings) TotalSize() (total int64) {
	for _, vb := range vbs {
		total += vb.SizeInBytes
	}
	return
}

// ApplyPlan creates new VolumeBindings according to volume plan
func (vbs VolumeBindings) ApplyPlan(plan VolumePlan) (res VolumeBindings) {
	for _, vb := range vbs {
		newVb := vb.DeepCopy()
		if vmap, _ := plan.GetVolumes(vb); vmap != nil {
			newVb.Source = vmap.GetDevice()
			if vmap.GetSize() > newVb.SizeInBytes {
				newVb.SizeInBytes = vmap.GetSize()
			}
		}
		res = append(res, newVb)
	}
	return
}

// MergeVolumeBindings combines two VolumeBindings
func MergeVolumeBindings(vbs1 VolumeBindings, vbs2 ...VolumeBindings) (vbs VolumeBindings) {
	vbMap := map[[3]string]*VolumeBinding{}
	for _, vbs := range append(vbs2, vbs1) {
		for _, vb := range vbs {
			if binding, ok := vbMap[vb.GetMapKey()]; ok {
				binding.SizeInBytes += vb.SizeInBytes
				binding.ReadIOPS += vb.ReadIOPS
				binding.WriteIOPS += vb.WriteIOPS
				binding.ReadBPS += vb.ReadBPS
				binding.WriteBPS += vb.WriteBPS
			} else {
				vbMap[vb.GetMapKey()] = &VolumeBinding{
					Source:      vb.Source,
					Destination: vb.Destination,
					Flags:       vb.Flags,
					SizeInBytes: vb.SizeInBytes,
					ReadIOPS:    vb.ReadIOPS,
					WriteIOPS:   vb.WriteIOPS,
					ReadBPS:     vb.ReadBPS,
					WriteBPS:    vb.WriteBPS,
				}
			}
		}
	}

	for _, vb := range vbMap {
		if vb.SizeInBytes >= 0 {
			vbs = append(vbs, vb)
		}
	}
	return vbs
}

// Volumes .
type Volumes map[string]int64

// DeepCopy .
func (v Volumes) DeepCopy() Volumes {
	res := Volumes{}
	for key, value := range v {
		res[key] = value
	}
	return res
}

// Add .
func (v Volumes) Add(v1 Volumes) {
	for key, value := range v1 {
		v[key] += value
	}
}

// Sub .
func (v Volumes) Sub(v1 Volumes) {
	for key, value := range v1 {
		v[key] -= value
	}
}

// GetDevice returns the first device
func (v Volumes) GetDevice() string {
	for key := range v {
		return key
	}
	return ""
}

// GetSize returns the first size
func (v Volumes) GetSize() int64 {
	for _, size := range v {
		return size
	}
	return 0
}

// Total .
func (v Volumes) Total() int64 {
	res := int64(0)
	for _, size := range v {
		res += size
	}
	return res
}

// VolumePlan is map from volume string to volumeMap: {"AUTO:/data:rw:100": Volumes{"/sda1": 100}}
type VolumePlan map[*VolumeBinding]Volumes

// UnmarshalJSON .
func (p *VolumePlan) UnmarshalJSON(b []byte) (err error) {
	if *p == nil {
		*p = VolumePlan{}
	}
	plan := map[string]Volumes{}
	if err = json.Unmarshal(b, &plan); err != nil {
		return err
	}
	for volume, vmap := range plan {
		vb, err := NewVolumeBinding(volume)
		if err != nil {
			return err
		}
		(*p)[vb] = vmap
	}
	return
}

// MarshalJSON .
func (p VolumePlan) MarshalJSON() ([]byte, error) {
	plan := map[string]Volumes{}
	for vb, vmap := range p {
		plan[vb.ToString(false)] = vmap
	}
	bs, err := json.Marshal(plan)
	return bs, err
}

// String .
func (p VolumePlan) String() string {
	bs, err := p.MarshalJSON()
	if err != nil {
		return "can not marshal volume plan"
	}
	return string(bs)
}

// Merge .
func (p VolumePlan) Merge(p2 VolumePlan) {
	for vb, vm := range p2 {
		if oldVM, oldVB := p.GetVolumes(vb); oldVB != nil {
			delete(p, oldVB)
			vm[vm.GetDevice()] += oldVM.GetSize()
			vm = Volumes{vm.GetDevice(): vm.GetSize() + oldVM.GetSize()}
			vb = &VolumeBinding{
				Source:      vb.Source,
				Destination: vb.Destination,
				Flags:       vb.Flags,
				SizeInBytes: vb.SizeInBytes + oldVB.SizeInBytes,
				ReadIOPS:    vb.ReadIOPS + oldVB.ReadIOPS,
				WriteIOPS:   vb.WriteIOPS + oldVB.WriteIOPS,
				ReadBPS:     vb.ReadBPS + oldVB.ReadBPS,
				WriteBPS:    vb.WriteBPS + oldVB.WriteBPS,
			}
		}
		p[vb] = vm
	}
}

// GetVolumes looks up Volumes according to volume destination directory
func (p VolumePlan) GetVolumes(vb *VolumeBinding) (volMap Volumes, volume *VolumeBinding) {
	for volume, volMap := range p {
		if vb.Destination == volume.Destination {
			return volMap, volume
		}
	}
	return
}
