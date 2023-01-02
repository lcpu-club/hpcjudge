package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/utilitycmd"
	"github.com/lcpu-club/hpcjudge/utilitycmd/configure"
	"github.com/lcpu-club/hpcjudge/utilitycmd/consts"
	"github.com/urfave/cli/v3"
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
		confFile, err := os.ReadFile(consts.ConfigureFilePath)
		if err != nil {
			return err
		}
		conf := new(configure.Configure)
		err = json.Unmarshal(confFile, conf)
		if err != nil {
			return err
		}
		return cmd.Init(conf)
	}
	app.Commands = []*cli.Command{
		{
			Name:   "problem-path",
			Usage:  "Get the path to problem files directory",
			Action: cmd.HandleProblemPath,
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
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("hpcgame: E:", err)
		os.Exit(-1)
	}
}
