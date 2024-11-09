package main

import "os"

func main() {
	app, err := NewCli()
	if err != nil {
		panic(err)
	}
	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
