package flags

import (
	"github.com/urfave/cli/v2"
)

const envVarPrefix = "FLUXGO"

func prefixEnvVars(name string) []string {
	return []string{envVarPrefix + "_" + name}
}

var (
	HttpHostFlag = &cli.StringFlag{
		Name:    "http-host",
		Usage:   "The host of the api",
		EnvVars: prefixEnvVars("HTTP_HOST"),
		//Required: true,
	}
	HttpPortFlag = &cli.IntFlag{
		Name:    "http-port",
		Usage:   "The port of the api",
		EnvVars: prefixEnvVars("HTTP_PORT"),
		Value:   8987,
		//Required: true,
	}
	ThisTestFlag = &cli.Int64Flag{
		Name:     "This-Test",
		Usage:    "This is a test",
		Required: false,
		Value:    777,
		EnvVars:  prefixEnvVars("THIS_TEST"),
		Action:   nil,
	}
)

var requiredFlags = []cli.Flag{
	//HttpHostFlag,
	//HttpPortFlag,
}

var optionalFlags = []cli.Flag{
	ThisTestFlag,
}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}

var Flags []cli.Flag