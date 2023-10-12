package main

import (
  "context"
  "fmt"
  "log"
  "os"
  "os/signal"

  "github.com/urfave/cli/v2"
)

func main(){
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
    BashComplete: cli.DefaultAppComplete,
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

func startTest() *cli.Command{
  var number int
  exampleFlag := &cli.IntFlag{
    Name: "example",
    Usage: "this is an example",
    Destination: &number,
    Required: true,
  }

  return &cli.Command{
    Name: "typing-test",
    Usage: "Start typing test",
    Flags: []cli.Flag{exampleFlag},
    Action: func(ctx *cli.Context) error {
      fmt.Printf("This is a test: %v", number)
      return nil
    },
  }
}

