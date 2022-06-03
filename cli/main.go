package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {

	app := &cli.App{
		Name:  "accbridge",
		Usage: "Accumulate Bridge CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "mint",
				Usage: "generates and signs tx to mint wrapped token",
				Action: func(c *cli.Context) error {
					fmt.Print("minting...")
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
