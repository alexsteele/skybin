package cmd

import (
	"log"
	skybinrepo "skybin/repo"
)

var syncCmd = Cmd{
	Name:        "sync",
	Usage:       "sync",
	Description: "Sync metadata with remote providers",
	Run:         runSync,
}

func runSync(args []string) {

	repo, err := skybinrepo.Open()
	if err != nil {
		log.Fatal(err)
	}

	err = repo.Sync()
	if err != nil {
		log.Fatal(err)
	}
}
