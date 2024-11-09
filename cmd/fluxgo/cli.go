package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	flag2 "github.com/wureny/FluxGo/flags"
	"github.com/wureny/FluxGo/ioc"
)

func NewCli() (*cli.App, error) {
	flag := flag2.Flags
	return &cli.App{
		Version:              "V0.1.0",
		Description:          "FluxGo cli",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:   "TryitOut",
				Usage:  "Test the FluxGo",
				Action: Action1,
				Flags:  nil,
			},
			{
				Name:   "RunFluxGo",
				Usage:  "Run the FluxGo",
				Action: RunFluxGo,
				Flags:  nil,
			},
		},
		Flags: flag,
	}, nil
}

func Action1(ctx *cli.Context) error {
	fmt.Println(ctx.Int64(flag2.ThisTestFlag.Name))
	return nil
}

func RunFluxGo(ctx *cli.Context) error {
	e, err := ioc.Init()
	if err != nil {
		return err
	}
	if err = e.Run(":8089"); err != nil {
		return err
	}
	return nil
}
