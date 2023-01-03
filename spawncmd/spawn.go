package spawncmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
	"github.com/lcpu-club/hpcjudge/common/runner"
	"github.com/lcpu-club/hpcjudge/spawncmd/models"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/satori/uuid"
)

type Spawner struct {
	cgroupBasePath string
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

var cgroupCPUPeriod uint64 = 50000 // in us

func (s *Spawner) ResourceControlToCgroup(path string, res *models.ResourceControl) (cgroup1.Cgroup, error) {
	quota := int64(cgroupCPUPeriod) * res.CPU / 100
	mem := res.Memory * 1024 * 1024
	cg, err := cgroup1.New(cgroup1.StaticPath(path), &specs.LinuxResources{
		Memory: &specs.LinuxMemory{
			Limit: &mem,
		},
		CPU: &specs.LinuxCPU{
			Quota:  &quota,
			Period: &cgroupCPUPeriod,
		},
	})
	if err != nil {
		return nil, err
	}
	return cg, nil
}

func (s *Spawner) SpawnCommand(cmd *exec.Cmd, user string, res *models.ResourceControl, id string) (cgroup1.Cgroup, error) {
	if id == "" {
		id = uuid.NewV4().String()
	}
	cmd, err := runner.CommandUseUser(cmd, user)
	if err != nil {
		return nil, err
	}
	cg, err := s.ResourceControlToCgroup(s.calcCgroupPath(id), res)
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		cg.Delete()
		return nil, err
	}
	err = cg.AddProc(uint64(cmd.Process.Pid))
	if err != nil {
		cmd.Process.Kill()
		cg.Delete()
		return nil, err
	}
	return cg, nil
}
