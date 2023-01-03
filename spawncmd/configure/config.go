package configure

type Configure struct {
	CgroupsBasePath string            `yaml:"cgroups-base-path"`
	StoragePath     map[string]string `yaml:"storage-path"`
}
