package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/common/version"
	"github.com/lcpu-club/hpcjudge/spawncmd"
	"github.com/lcpu-club/hpcjudge/spawncmd/consts"
	"github.com/lcpu-club/hpcjudge/spawncmd/models"
	"github.com/urfave/cli/v3"
)

func main() {
	if os.Getenv(consts.SpawnEnvVar) != consts.SpawnEnvVarValue {
		log.Fatalln("Command", os.Args[0], "should not be called directly from outside.")
	}
	if os.Getuid() != 0 {
		log.Fatalln("Command", os.Args[0], "requires root permission.")
	}
	app := cli.NewApp()
	app.Name = "hpc-spawn"
	app.Usage = "for internal usage"
	app.Version = version.Version
	app.Authors = []*cli.Author{}
	for _, author := range version.Authors {
		app.Authors = append(app.Authors, &cli.Author{Name: author[0], Email: author[1]})
	}
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "config",
		Aliases:     []string{"c", "conf"},
		Usage:       "Configure file path",
		Value:       consts.ConfigureFilePath,
		DefaultText: consts.ConfigureFilePath,
	})
	cmd := spawncmd.NewCommand()
	app.Before = func(ctx *cli.Context) error {
		confFile := ctx.String("config")
		return cmd.Init(confFile)
	}
	app.Commands = []*cli.Command{
		{
			Name:  "run-judge-script",
			Usage: "Executes a judge script",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "data",
					Required: true,
				},
			},
			Action: func(ctx *cli.Context) error {
				dText := ctx.String("data")
				d := new(models.RunJudgeScriptData)
				err := json.Unmarshal([]byte(dText), d)
				if err != nil {
					return err
				}
				return cmd.RunJudgeScript(d)
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
