package utilitycmd

import (
	"fmt"
	"path/filepath"

	commonConsts "github.com/lcpu-club/hpcjudge/common/consts"
	"github.com/lcpu-club/hpcjudge/common/runner"
	"github.com/lcpu-club/hpcjudge/utilitycmd/configure"
	"github.com/urfave/cli/v3"
)

type Command struct {
	configure   *configure.Configure
	inJudge     bool
	judgeStatus *runner.Status
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

func ErrNotInJudge(command string) error {
	return fmt.Errorf("subcommand %v cannot be called when not in judge", command)
}

func (c *Command) HandleProblemPath(ctx *cli.Context) error {
	if !c.inJudge {
		return ErrNotInJudge(ctx.Command.Name)
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
	p := filepath.Join(
		c.configure.StoragePath["solution"],
		c.judgeStatus.SolutionID,
		commonConsts.SolutionFileName,
	)
	fmt.Println(p)
	return nil
}

func (c *Command) HandleReport(ctx *cli.Context) error {
	return nil
}
