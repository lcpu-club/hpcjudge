package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lcpu-club/hpcjudge/bridge/configure"
	"github.com/lcpu-club/hpcjudge/bridge/server"
	"github.com/lcpu-club/hpcjudge/common/version"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"
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
			confFile, err := os.ReadFile(ctx.String("configure"))
			if err != nil {
				return err
			}
			conf := new(configure.Configure)
			err = yaml.Unmarshal(confFile, conf)
			if err != nil {
				return err
			}
			if conf.Discovery == nil ||
				conf.MinIO == nil ||
				conf.MinIO.Buckets == nil ||
				conf.MinIO.Credentials == nil {
				return fmt.Errorf("invalid configure, some required parameters not set")
			}
			srv, err := server.NewServer(conf)
			if err != nil {
				return err
			}
			return srv.Start()
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "configure",
				Aliases:     []string{"config", "c"},
				Usage:       "The path to configure file (yaml format)",
				Value:       "/etc/hpc-bridge.yml",
				DefaultText: "/etc/hpc-bridge.yml",
			},
		},
	})
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
