package main

import (
	"log"
	"os"
	"skybin/cmd"
)

func init() {
	log.SetFlags(0)
}

func main() {
	if len(os.Args) < 2 {
		cmd.Usage()
		os.Exit(1)
	}

	for _, command := range cmd.Commands {
		if command.Name == os.Args[1] {
			command.Run(os.Args[2:])
			return
		}
	}

	cmd.Usage()
	os.Exit(1)
}
