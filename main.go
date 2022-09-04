package main

import "github.com/alecthomas/kong"

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	if err := ctx.Run(); err != nil {
		panic(err)
	}
}
