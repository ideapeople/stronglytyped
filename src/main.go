package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"
)

func main() {
	_, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt)

	app := &cli.App{
		Name:  "stronglytyped",
		Usage: "Typing test",
		Authors: []*cli.Author{
			{
				Name: "ideapeople",
			},
		},
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Commands: []*cli.Command{
			startTest(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		rootCancel()
		log.Fatal(err)
	}

	rootCancel()
}

func startTest() *cli.Command {
	var length int

	lengthFlag := &cli.IntFlag{
		Name:        "length",
		Usage:       "Specify the length of the test in seconds",
		Destination: &length,
		Required:    false,
		Value:       30,
	}

	return &cli.Command{
		Name:  "tt",
		Usage: "Start typing test",
		Flags: []cli.Flag{lengthFlag},
		Action: func(ctx *cli.Context) error {
			fmt.Printf("This is a test: %v", length)
			return nil
		},
	}
}
