package utilitycmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lcpu-club/hpcjudge/bridge"
	bridgeApi "github.com/lcpu-club/hpcjudge/bridge/api"
	"github.com/lcpu-club/hpcjudge/common"
	commonConsts "github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/common/runner"
	"github.com/lcpu-club/hpcjudge/utilitycmd/configure"
	"github.com/urfave/cli/v3"
)

type Command struct {
	configure    *configure.Configure
	inJudge      bool
	judgeStatus  *runner.Status
	bridgeClient *bridge.Client
}

func NewCommand() *Command {
	cmd := new(Command)
	return cmd
}

func (c *Command) Init(conf *configure.Configure) error {
	c.configure = conf
	judgeStatus, err := runner.GetStatus(c.configure.StoragePath)
	if err != nil {
		c.inJudge = false
	} else {
		c.inJudge = true
		c.judgeStatus = judgeStatus
	}
	return nil
}

func (c *Command) getBridgeClient() *bridge.Client {
	if c.bridgeClient == nil {
		cc := common.NewCommonSignedMultiAddressClient(
			c.configure.Bridge.Address,
			[]byte(c.configure.Bridge.SecretKey),
			c.configure.Bridge.Timeout,
			true,
		)
		c.bridgeClient = bridge.NewClient(cc)
	}
	return c.bridgeClient
}

func ErrNotInJudge(command string) error {
	return fmt.Errorf("subcommand %v cannot be called when not in judge", command)
}

func ErrWrongArgumentNumber(command string, expected string) error {
	return fmt.Errorf(
		"wrong argument number for %v, expected %v\r\n   (use \"%v help %v\" for help)",
		command, expected, os.Args[0], command,
	)
}

func (c *Command) HandleProblemPath(ctx *cli.Context) error {
	if !c.inJudge {
		return ErrNotInJudge(ctx.Command.Name)
	}
	if ctx.Args().Len() >= 2 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "0 or 1")
	}
	subpath := ctx.Args().Get(0)
	p := filepath.Join(c.configure.StoragePath["problem"], c.judgeStatus.ProblemID)
	if subpath != "" {
		p = filepath.Join(p, subpath)
	}
	fmt.Println(p)
	return nil
}

func (c *Command) HandleSolutionPath(ctx *cli.Context) error {
	if !c.inJudge {
		return ErrNotInJudge(ctx.Command.Name)
	}
	if ctx.Args().Len() >= 1 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "0")
	}
	p := filepath.Join(
		c.configure.StoragePath["solution"],
		c.judgeStatus.SolutionID,
		commonConsts.SolutionFileName,
	)
	fmt.Println(p)
	return nil
}

var ErrPathNotInPartition = fmt.Errorf("path not in partition")

func (c *Command) pathToPartitionedPath(path string) (*bridgeApi.PartitionedPath, error) {
	prefixLength := -1
	current := new(bridgeApi.PartitionedPath)
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	for partition, base := range c.configure.StoragePath {
		if strings.HasPrefix(p, base) && prefixLength < len(base) {
			prefixLength = len(base)
			current.Partition = partition
			current.Path = strings.TrimLeft(strings.TrimPrefix(p, base), "/")
		}
	}
	if prefixLength == -1 {
		return nil, ErrPathNotInPartition
	}
	return current, nil
}

func (c *Command) HandleReport(ctx *cli.Context) error {
	if !c.inJudge {
		return ErrNotInJudge(ctx.Command.Name)
	}
	if ctx.Args().Len() != 1 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "1")
	}
	reportFile := ctx.Args().Get(0)
	partitionedPath, err := c.pathToPartitionedPath(reportFile)
	if err != nil {
		return err
	}
	bc := c.getBridgeClient()
	err = bc.UploadFile(
		partitionedPath.Partition,
		partitionedPath.Path,
		bridgeApi.BucketSolution,
		filepath.Join(c.judgeStatus.SolutionID, commonConsts.JudgeReportFile),
	)
	return err
}

func (c *Command) HandleUploadArtifact(ctx *cli.Context) error {
	if !c.inJudge {
		return ErrNotInJudge(ctx.Command.Name)
	}
	if ctx.Args().Len() != 2 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "2")
	}
	target := ctx.Args().Get(0)
	file := ctx.Args().Get(1)
	if target == "" || strings.ContainsAny(target, "/?#$%&*<>'\"\\`") {
		return fmt.Errorf("invalid target name \"%v\"", target)
	}
	objKey := filepath.Join(c.judgeStatus.SolutionID, commonConsts.JudgeArtifactsPath, target)
	partitionedPath, err := c.pathToPartitionedPath(file)
	if err != nil {
		return err
	}
	bc := c.getBridgeClient()
	err = bc.UploadFile(
		partitionedPath.Partition,
		partitionedPath.Path,
		bridgeApi.BucketSolution,
		objKey,
	)
	return err
}

var ErrMaskOperationsNotAllowedOnThisNode = fmt.Errorf("mask operations not allowed on this node")

func (c *Command) HandleMaskWrite(ctx *cli.Context) error {
	if !c.configure.AllowMask {
		return ErrMaskOperationsNotAllowedOnThisNode
	}
	if ctx.Args().Len() != 1 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "1")
	}
	path := ctx.Args().Get(0)
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	home, err := runner.GetHomeDirectory()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(path, home) {
		return fmt.Errorf("file %v not in user's home folder %v, permission denied", path, home)
	}
	err = os.Chown(path, 0, 0)
	if err != nil {
		return err
	}
	err = os.Chmod(path, os.FileMode(0755))
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) HandleMaskRead(ctx *cli.Context) error {
	if !c.configure.AllowMask {
		return ErrMaskOperationsNotAllowedOnThisNode
	}
	if ctx.Args().Len() != 1 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "1")
	}
	path := ctx.Args().Get(0)
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	home, err := runner.GetHomeDirectory()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(path, home) {
		return fmt.Errorf("file %v not in user's home folder %v, permission denied", path, home)
	}
	err = os.Chown(path, 0, 0)
	if err != nil {
		return err
	}
	err = os.Chmod(path, os.FileMode(0600))
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) HandleUnmask(ctx *cli.Context) error {
	if !c.configure.AllowMask {
		return ErrMaskOperationsNotAllowedOnThisNode
	}
	if ctx.Args().Len() != 1 {
		return ErrWrongArgumentNumber(ctx.Command.Name, "1")
	}
	path := ctx.Args().Get(0)
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	home, err := runner.GetHomeDirectory()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(path, home) {
		return fmt.Errorf("file %v not in user's home folder %v, permission denied", path, home)
	}
	err = os.Chown(path, os.Getuid(), os.Getgid())
	if err != nil {
		return err
	}
	err = os.Chmod(path, os.FileMode(0750))
	if err != nil {
		return err
	}
	return nil
}
