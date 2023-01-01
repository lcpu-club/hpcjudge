package models

type ResourceControl struct {
	Memory int64 `json:"memory" yaml:"memory" toml:"memory"` // Memory limit in MBytes
	CPU    int64 `json:"cpu" yaml:"cpu" toml:"cpu"`          // 100 for one cpu, 200 for two, 300 for three...
}
