package configure

import (
	"time"

	"github.com/satori/uuid"
)

type Configure struct {
	ID               uuid.UUID           `yaml:"id"`
	SpawnCmd         string              `yaml:"spawn-cmd"`
	Nsq              *NsqConfigure       `yaml:"nsq"`
	MinIO            *MinIOConfigure     `yaml:"minio"`
	Redis            *RedisConfigure     `yaml:"redis"`
	Discovery        *DiscoveryConfigure `yaml:"discovery"`
	Bridge           *BridgeConfigure    `yaml:"bridge"`
	EnableStatistics bool                `yaml:"enable-statistics"`
}

type NsqConfigure struct {
	Nsqd         *NsqdConfigure       `yaml:"nsqd"`
	NsqLookupd   *NsqLookupdConfigure `yaml:"nsqlookupd"`
	MaxAttempts  int                  `yaml:"max-attempts"`
	RequeueDelay time.Duration        `yaml:"requeue-delay"`
	MsgTimeout   time.Duration        `yaml:"msg-timeout"` // Minimum: 1s
	AuthSecret   string               `yaml:"auth-secret"`
	Concurrent   int                  `yaml:"concurrent"`
	Topics       *NsqTopicConfigure   `yaml:"topics"`
	Channel      string               `yaml:"channel"`
}

type NsqdConfigure struct {
	Address string `yaml:"address"`
}

type NsqTopicConfigure struct {
	Judge  string `yaml:"judge"`
	Report string `yaml:"report"`
}

type NsqLookupdConfigure struct {
	Address []string `yaml:"address"`
}

type MinIOConfigure struct {
	Endpoint        string                     `yaml:"endpoint"`
	Credentials     *MinIOCredentialsConfigure `yaml:"credentials"`
	SSL             bool                       `yaml:"ssl"`
	Buckets         *MinIOBucketsConfigure     `yaml:"buckets"`
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

type RedisConfigure struct {
	Address   string                `yaml:"address"`
	Password  string                `yaml:"password"`
	KeepAlive time.Duration         `yaml:"keep-alive"`
	Database  int                   `yaml:"database"`
	Prefix    string                `yaml:"prefix"`
	Expire    *RedisExpireConfigure `yaml:"expire"`
}

type RedisExpireConfigure struct {
	Report time.Duration `yaml:"report"`
	Judge  time.Duration `yaml:"judge"`
}

type DiscoveryConfigure struct {
	Address   []string      `yaml:"address"`
	AccessKey string        `yaml:"access-key"`
	Timeout   time.Duration `yaml:"timeout"`
}

type BridgeConfigure struct {
	SecretKey string        `yaml:"secret-key"`
	Timeout   time.Duration `yaml:"timeout"`
}
