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
	SecretKey       string              `yaml:"secret-key"`
	Discovery       *DiscoveryConfigure `yaml:"discovery"`
	StoragePath     map[string]string   `yaml:"storage-path"`
	MinIO           *MinIOConfigure     `yaml:"minio"`
}

type DiscoveryConfigure struct {
	Address   []string      `yaml:"address"`
	AccessKey string        `yaml:"access-key"`
	Timeout   time.Duration `yaml:"timeout"`
}

type MinIOConfigure struct {
	Endpoint    string                     `yaml:"endpoint"`
	Credentials *MinIOCredentialsConfigure `yaml:"credentials"`
	SSL         bool                       `yaml:"ssl"`
	Buckets     *MinIOBucketsConfigure     `yaml:"Buckets"`
}

type MinIOCredentialsConfigure struct {
	AccessKey string `yaml:"access-key"`
	SecretKey string `yaml:"secret-key"`
}

type MinIOBucketsConfigure struct {
	Problem  string `yaml:"problem"`
	Solution string `yaml:"solution"`
}
