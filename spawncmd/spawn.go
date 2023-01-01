package spawncmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/lcpu-club/hpcjudge/common"
	"github.com/lcpu-club/hpcjudge/spawncmd/models"
	"github.com/satori/uuid"
)

type Spawner struct {
	cgroupBasePath string
	currentID      int64
	currentIDLock  *sync.Mutex
}

var ErrCgroupsV1NotAvailable = fmt.Errorf("cgroups v1 not available")

func NewSpawner(cgroupBasePath string) *Spawner {
	return &Spawner{
		cgroupBasePath: cgroupBasePath,
	}
}

func (s *Spawner) calcCgroupPath(sub string) string {
	sb := []byte(sub)
	haystack := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_.")
	for i := range sb {
		if !bytes.Contains(haystack, []byte{sb[i]}) {
			sb[i] = '-'
		}
	}
	return filepath.Join(s.cgroupBasePath, string(sb))
}

func (s *Spawner) Init() error {
	if cgroups.Mode() != cgroups.Legacy && cgroups.Mode() != cgroups.Hybrid {
		return ErrCgroupsV1NotAvailable
	}
	return nil
}

func (s *Spawner) ResourceControlToCgroup(path string, res *models.ResourceControl) error {
	cg, err := cgroup1.New(cgroup1.StaticPath(path))
	if err != nil {
		return err
	}
	return nil
}

func (s *Spawner) SpawnCommand(cmd *exec.Cmd, user string, res *models.ResourceControl, id string) error {
	if id == "" {
		id = uuid.NewV4().String()
	}
	cmd, err := common.CommandUseUser(cmd, user)
	if err != nil {
		return err
	}
	s.ResourceControlToCgroup(s.calcCgroupPath(id), res)
	return nil
}
