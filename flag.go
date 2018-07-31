package main

import "github.com/urfave/cli"

// FlagSet ...set flag option
func FlagSet() *cli.App {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration *.toml",
		},
		cli.StringFlag{
			Name:  "port, p",
			Value: "3000",
			Usage: "Server port to be listened",
		},
		cli.StringFlag{
			Name:  "region, r",
			Value: "ap-northeast-1",
			Usage: "Setting AWS region for tomlssm",
		},
	}
	return app
}