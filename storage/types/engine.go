package types

// EngineParams .
type EngineParams struct {
	Volumes       []string          `json:"volumes" mapstructure:"volumes"`
	VolumeChanged bool              `json:"volume_changed" mapstructure:"volume_changed"` // indicates whether the realloc request includes new volumes
	Storage       int64             `json:"storage" mapstructure:"storage"`
	IOPSOptions   map[string]string `json:"iops_options" mapstructure:"iops_options"`
}
