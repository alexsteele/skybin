package cmd

import (
	"log"
	skybinrepo "skybin/repo"
)

var storeCmd = Cmd{
	Name:        "store",
	Description: "Store a file in the skybin network",
	Usage:       "store <path>",
	Run:         runStore,
}

func runStore(args []string) {
	if len(args) < 1 {
		log.Fatal("must provide path")
	}

	repo, err := skybinrepo.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Store(args[0])
	if err != nil {
		log.Fatal(err)
	}
}
