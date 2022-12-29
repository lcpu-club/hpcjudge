package main

import (
	"log"
	"os"

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
			return nil
		},
		Flags: []cli.Flag{},
	})
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
