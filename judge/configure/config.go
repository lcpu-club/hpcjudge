package configure

import (
	"time"

	"github.com/satori/uuid"
)

type Configure struct {
	ID              uuid.UUID           `yaml:"uuid"`
	ExternalAddress string              `yaml:"external-address"`
	Nsq             *NsqConfigure       `yaml:"nsq"`
	MinIO           *MinIOConfigure     `yaml:"minio"`
	Discovery       *DiscoveryConfigure `yaml:"discovery"`
}

type NsqConfigure struct {
	Nsqd       *NsqdConfigure       `yaml:"nsqd"`
	NsqLookupd *NsqLookupdConfigure `yaml:"nsqlookupd"`
	AuthSecret string               `yaml:"auth-secret"`
	Concurrent int                  `yaml:"concurrent"`
}

type NsqdConfigure struct {
	Address string              `yaml:"address"`
	Topics  *NsqdTopicConfigure `yaml:"topics"`
}

type NsqdTopicConfigure struct {
	Judge   string `yaml:"judge"`
	Sandbox string `yaml:"sandbox"`
}

type NsqLookupdConfigure struct {
	Address []string                  `yaml:"address"`
	Topics  *NsqLookupdTopicConfigure `yaml:"topics"`
	Channel string                    `yaml:"channel"`
}

type NsqLookupdTopicConfigure struct {
	Judge   string `yaml:"judge"`
	Sandbox string `yaml:"sandbox"`
}

type MinIOConfigure struct {
	Endpoint        string                     `yaml:"endpoint"`
	Credentials     *MinIOCredentialsConfigure `yaml:"credentials"`
	SSL             bool                       `yaml:"ssl"`
	Buckets         *MinIOBucketsConfigure     `yaml:"Buckets"`
	PresignedExpiry time.Duration              `yaml:"presigned-expiry"`
}

type MinIOCredentialsConfigure struct {
	AccessKey string `yaml:"access-key"`
	SecretKey string `yaml:"secret-key"`
}

type MinIOBucketsConfigure struct {
	Problem  string `yaml:"problem"`
	Solution string `yaml:"solution"`
}

type DiscoveryConfigure struct {
	Address    []string      `yaml:"address"`
	AccessKey  string        `yaml:"access-key"`
	Timeout    time.Duration `yaml:"timeout"`
	InformType string        `yaml:"inform-type"`
	InformTags []string      `yaml:"inform-tags"`
}
