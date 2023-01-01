package main

import (
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/spawncmd/consts"
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
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
