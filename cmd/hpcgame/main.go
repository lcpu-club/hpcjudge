package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/lcpu-club/hpcjudge/utilitycmd"
	"github.com/lcpu-club/hpcjudge/utilitycmd/configure"
	"github.com/lcpu-club/hpcjudge/utilitycmd/consts"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"
)

func main() {
	consts.InSpawnMode = os.Getenv(consts.UtilitySpawnEnvVar) == consts.UtilitySpawnEnvVarValue
	if consts.EnableDevMode {
		consts.InDevMode = os.Getenv("HPCGAME_DEV_MODE") == "true"
	}
	if !consts.InSpawnMode {
		if !consts.InDevMode {
			if os.Geteuid() != 0 {
				log.Fatalln(
					"Please ensure this executable is owned by root:root and have the setuid bit set.\r\n",
					"If under development, please export HPCGAME_DEV_MODE=true.\r\n",
					"Otherwise, please: chown root:root "+os.Args[0]+" && chmod u+s "+os.Args[0],
				)
			}
		}
	}
	app := cli.NewApp()
	app.Name = "hpcgame"
	app.Usage = "HPCGame Command Line Utility"
	cmd := utilitycmd.NewCommand()
	app.Before = func(ctx *cli.Context) error {
		cStat, err := os.Stat(consts.ConfigureFilePath)
		if err != nil {
			return err
		}
		sStat, ok := cStat.Sys().(*syscall.Stat_t)
		if ok {
			if sStat.Uid != 0 || sStat.Gid != 0 {
				return fmt.Errorf("configure file should be owned by root:root, or there will be security risks")
			}
		}
		confFile, err := os.ReadFile(consts.ConfigureFilePath)
		if err != nil {
			return err
		}
		conf := new(configure.Configure)
		err = yaml.Unmarshal(confFile, conf)
		if err != nil {
			return err
		}
		return cmd.Init(conf)
	}
	app.Commands = []*cli.Command{
		{
			Name:      "problem-path",
			Usage:     "Get the path to problem files directory",
			ArgsUsage: "[PROBLEM_FILE_SUBPATH]",
			Action:    cmd.HandleProblemPath,
		},
		{
			Name:   "solution-path",
			Usage:  "Get the path to solution file",
			Action: cmd.HandleSolutionPath,
		},
		{
			Name:   "report",
			Usage:  "Report judge result in JSON format",
			Action: cmd.HandleReport,
		},
		{
			Name:      "mask-write",
			Usage:     "Protect a file from being written by user",
			ArgsUsage: "FILE_NAME",
			Action:    cmd.HandleMaskWrite,
		},
		{
			Name:      "mask-read",
			Usage:     "Protect a file from being read or written by user",
			ArgsUsage: "FILE_NAME",
			Action:    cmd.HandleMaskRead,
		},
		{
			Name:      "unmask",
			Usage:     "Disable protection on a file",
			ArgsUsage: "FILE_NAME",
			Action:    cmd.HandleUnmask,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("hpcgame:", err)
		os.Exit(-1)
	}
}
