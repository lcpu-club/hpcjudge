package configure

import "time"

type Configure struct {
	Bridge      *BridgeConfigure  `json:"bridge"`
	StoragePath map[string]string `json:"storage-path"`
}

type BridgeConfigure struct {
	SecretKey []byte        `yaml:"secret-key"`
	Timeout   time.Duration `yaml:"timeout"`
	Address   []string      `yaml:"address"`
}
