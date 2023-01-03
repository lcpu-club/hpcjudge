package main

import (
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/common/version"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.NewApp()
	app.Name = "hpc-bridge"
	app.Usage = "HPC and Judger communication bridge"
	app.Version = version.Version
	app.Authors = []*cli.Author{}
	for _, author := range version.Authors {
		app.Authors = append(app.Authors, &cli.Author{Name: author[0], Email: author[1]})
	}
	app.Commands = append(app.Commands, &cli.Command{
		Name:    "serve",
		Aliases: []string{"s", "run"},
		Usage:   "Start the hpc-bridge service",
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
