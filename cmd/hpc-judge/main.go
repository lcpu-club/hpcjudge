package main

import (
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/judge/configure"
	"github.com/lcpu-club/hpcjudge/judge/server"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.NewApp()
	app.Name = "hpc-judge"
	app.Usage = "HPCGame Judger Service"
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "run"},
		Usage:   "Start the hpc-judge service",
		Action: func(ctx *cli.Context) error {
			confFile := ctx.String("configure")
			conf, err := configure.LoadConfigure(confFile)
			if err != nil {
				return err
			}
			judger, err := server.NewJudger(conf)
			if err != nil {
				return err
			}
			return judger.Run()
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "configure",
				Aliases:     []string{"config", "c"},
				Usage:       "The path to configure file (yaml format)",
				Value:       "/etc/hpc-judge.yml",
				DefaultText: "/etc/hpc-judge.yml",
			},
		},
	})
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
