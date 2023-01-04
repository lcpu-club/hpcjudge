package spawncmd

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/lcpu-club/hpcjudge/common/runner"
	"github.com/lcpu-club/hpcjudge/spawncmd/configure"
	"github.com/lcpu-club/hpcjudge/spawncmd/models"
	"github.com/lcpu-club/hpcjudge/utilitycmd/replacer"
	"gopkg.in/yaml.v2"
)

type Command struct {
	configure *configure.Configure
	spawner   *Spawner
}

func NewCommand() *Command {
	cmd := &Command{}
	return cmd
}

func (c *Command) Init(conf string) error {
	cFile, err := os.ReadFile(conf)
	if err != nil {
		return err
	}
	c.configure = new(configure.Configure)
	err = yaml.Unmarshal(cFile, c.configure)
	if err != nil {
		return err
	}
	c.spawner = NewSpawner(c.configure.CgroupsBasePath)
	err = c.spawner.Init()
	return err
}

func (c *Command) deleteFile(path string) error {
	if path == "" {
		return nil
	}
	return os.Remove(path)
}

func (c *Command) RunJudgeScript(d *models.RunJudgeScriptData) error {
	if d.ResourceControl == nil {
		d.ResourceControl = &models.ResourceControl{
			CPU:    100,
			Memory: 1024,
		}
	}
	solutionPath := filepath.Join(c.configure.StoragePath["solution"], d.SolutionID)
	problemPath := filepath.Join(c.configure.StoragePath["problem"], d.ProblemID)
	defer func() {
		if d.AutoRemoveSolution {
			os.RemoveAll(solutionPath)
		}
	}()
	err := os.Chmod(solutionPath, os.FileMode(0755))
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	tmpPath := ""
	defer c.deleteFile(tmpPath)
	if d.Command != "" {
		cmd = exec.Command("/bin/bash", "-c", d.Command)
	} else {
		script, err := os.ReadFile(filepath.Join(problemPath, d.Script))
		if err != nil {
			return err
		}
		replacer := replacer.NewReplacer(
			d.SolutionID,
			d.ProblemID,
			d.Username,
			c.configure.StoragePath,
		)
		script = []byte(replacer.Replace(string(script)))
		tmpPath = filepath.Join(solutionPath, "judge-script.sh")
		err = os.WriteFile(tmpPath, script, os.FileMode(0755))
		if err != nil {
			return err
		}
		u, err := user.Lookup(d.Username)
		if err != nil {
			return err
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return err
		}
		err = os.Chown(tmpPath, uid, gid)
		if err != nil {
			return err
		}
		cmd = exec.Command("/bin/bash", tmpPath)
	}
	err = runner.WriteStatus(c.configure.StoragePath, d.ProblemID, d.SolutionID, -1)
	defer runner.ClearStatus(c.configure.StoragePath)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cg, err := c.spawner.SpawnCommand(cmd, d.Username, d.ResourceControl, d.SolutionID)
	if err != nil {
		if cg != nil {
			cg.Delete()
		}
		return err
	}
	defer cg.Delete()
	err = runner.WriteStatus(c.configure.StoragePath, d.ProblemID, d.SolutionID, cmd.Process.Pid)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return cmd.Wait()
}
