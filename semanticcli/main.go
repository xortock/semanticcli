package main

import (
	"os"

	"github.com/urfave/cli/v2"
	"github.com/xortock/semanticcli/internal/flags"
	"github.com/xortock/semanticcli/internal/handlers"
)

var version = "0.0.0.0"

func main() {

	var cliHandler = handlers.NewCliHandler()

	var app = &cli.App{
		Name:    "semanticcli",
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "xortock",
				Email: "bgmaduro@gmail.com",
			},
		},
		Copyright:       "(C) 2024 xortock",
		HideHelp:        false,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     flags.BUCKET,
				Usage:    "s3 bucket to use [required]",
				Required: true,
			},
			&cli.StringFlag{
				Name:     flags.FILE,
				Usage:    "file name used to store version [required]",
				Required: true,
			},
			&cli.StringFlag{
				Name:     flags.MAJOR,
				Usage:    "the major version",
				Required: false,
			},
			&cli.StringFlag{
				Name:     flags.MINOR,
				Usage:    "the minor version",
				Required: false,
			},
			&cli.StringFlag{
				Name:     flags.PATCH,
				Usage:    "the patch version",
				Required: false,
			},
			&cli.StringFlag{
				Name:     flags.BUILD,
				Usage:    "the build version",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     flags.DETAILS,
				Usage:    "the current stored version",
				Required: false,
			},
		},
		Action: cliHandler.Handle,
	}

	var _ = app.Run(os.Args)
}
