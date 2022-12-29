package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.NewApp()
	app.Name = "hpc-agent"
	app.Usage = "HPC node agent"
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "relay",
		Aliases: []string{"r"},
		Usage:   "Start the HPC relay service",
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
