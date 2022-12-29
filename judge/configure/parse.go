package configure

import (
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
	return c, nil
}
