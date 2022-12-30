package configure

import (
	"time"

	"github.com/satori/uuid"
)

type Configure struct {
	ID              uuid.UUID           `yaml:"uuid"`
	Tags            []string            `yaml:"tags"`
	Listen          string              `yaml:"listen"`
	ExternalAddress string              `yaml:"external-address"`
	SecretKey       []byte              `yaml:"secret-key"`
	Discovery       *DiscoveryConfigure `yaml:"discovery"`
}

type DiscoveryConfigure struct {
	Address   []string      `yaml:"address"`
	AccessKey string        `yaml:"access-key"`
	Timeout   time.Duration `yaml:"timeout"`
}
