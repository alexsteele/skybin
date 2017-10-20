package cmd

import (
	"log"
	skybinrepo "skybin/repo"
)

var infoCmd = Cmd{
	Name:        "info",
	Usage:       "info",
	Description: "Print information about the repo",
	Run:         runInfo,
}

func runInfo(args []string) {

	repo, err := skybinrepo.Open()
	if err != nil {
		log.Fatal(err)
	}

	info := repo.Info()

	log.Println()
	log.Println("Home Directory:", info.HomeDir)
	log.Println("User ID:", info.Config.UserId)
	log.Println("Node ID:", info.Config.NodeId)
	log.Println()
}
