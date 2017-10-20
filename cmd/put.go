package cmd

import (
	"log"
	skybinrepo "skybin/repo"
)

var storeCmd = Cmd{
	Name:        "put",
	Description: "Store a file in the skybin network",
	Usage:       "put <path>",
	Run:         runPut,
}

func runPut(args []string) {
	if len(args) < 1 {
		log.Fatal("must provide path")
	}

	repo, err := skybinrepo.Open()
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Put(args[0], nil)
	if err != nil {
		log.Fatal(err)
	}
}
