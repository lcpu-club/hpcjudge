package configure

import "time"

type Configure struct {
	Bridge      *BridgeConfigure  `yaml:"bridge"`
	StoragePath map[string]string `yaml:"storage-path"`
}

type BridgeConfigure struct {
	SecretKey string        `yaml:"secret-key"`
	Timeout   time.Duration `yaml:"timeout"`
	Address   []string      `yaml:"address"`
}
