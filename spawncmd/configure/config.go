package configure

type Configure struct {
	CgroupsBasePath string            `json:"cgroups-base-path"`
	StoragePath     map[string]string `json:"storage-path"`
}
