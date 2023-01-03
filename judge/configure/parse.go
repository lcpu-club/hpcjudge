package configure

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfigure(path string) (*Configure, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := new(Configure)
	err = yaml.Unmarshal(f, c)
	if err != nil {
		return nil, err
	}
	if c.Bridge == nil ||
		c.Discovery == nil ||
		c.MinIO == nil ||
		c.MinIO.Credentials == nil ||
		c.Nsq == nil ||
		c.Nsq.NsqLookupd == nil ||
		c.Nsq.Nsqd == nil ||
		c.Nsq.Topics == nil ||
		c.Redis == nil ||
		c.Redis.Expire == nil {
		return nil, fmt.Errorf("invalid configure, some required parameters not set")
	}
	return c, nil
}
