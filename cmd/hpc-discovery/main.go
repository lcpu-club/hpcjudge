package main

import (
	"log"
	"os"
	"time"

	"github.com/lcpu-club/hpcjudge/discovery/server"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.NewApp()
	app.Name = "hpc-discovery"
	app.Usage = "HPC Judge cluster discovery service"
	app.Commands = []*cli.Command{
		{
			Name:    "serve",
			Aliases: []string{"s", "run"},
			Usage:   "Start the hpc-discovery service",
			Action: func(c *cli.Context) error {
				log.Println("HPC Discovery Service")
				srv, err := server.NewServer(c.String("listen"), c.String("external-address"), c.String("data"), c.StringSlice("peers"), c.String("access-key"), c.Duration("peer-timeout"))
				if err != nil {
					return err
				}
				return srv.Start()
			},
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "peers",
					Aliases: []string{"p"},
					Usage:   "Specify peer hpc-discovery services",
				},
				&cli.StringFlag{
					Name:        "data",
					Aliases:     []string{"d"},
					Usage:       "Specify the path to hpc-discovery database file",
					Value:       "/tmp/hpc-discovery.dat",
					DefaultText: "/tmp/hpc-discovery.dat",
				},
				&cli.StringFlag{
					Name:        "listen",
					Aliases:     []string{"l", "a", "address"},
					Usage:       "Specify which address should hpc-discovery listen on",
					Value:       ":20751",
					DefaultText: ":20751",
				},
				&cli.StringFlag{
					Name:        "external-address",
					Aliases:     []string{"e"},
					Usage:       "Specify the address that other services can reach, leave empty to disable peering",
					Value:       "http://127.0.0.1:20751",
					DefaultText: "http://127.0.0.1:20751",
				},
				&cli.StringFlag{
					Name:    "access-key",
					Aliases: []string{"key", "k"},
					Usage:   "Specify the access key, leave empty to disable",
					Value:   "",
				},
				&cli.DurationFlag{
					Name:        "peer-timeout",
					Aliases:     []string{"t"},
					Usage:       "Specify the timeout duration for peering requests",
					Value:       5 * time.Second,
					DefaultText: "5s",
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
