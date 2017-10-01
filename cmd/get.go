package cmd

import (
	"log"
	"os"
	skybinrepo "skybin/repo"
)

var getCmd = Cmd{
	Name:        "get",
	Description: "Download a file from the skybin network",
	Usage:       "get <file> [destfile]",
	Run:         runGet,
}

func runGet(args []string) {
	if len(args) < 1 {
		log.Fatal("must provide filename")
	}

	repo, err := skybinrepo.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Get(args[0], os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
