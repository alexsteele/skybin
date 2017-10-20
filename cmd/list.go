package cmd

import (
	"fmt"
	"log"
	skybinrepo "skybin/repo"
)

var listCmd = Cmd{
	Name:        "list",
	Description: "List files saved in the skybin network",
	Usage:       "list",
	Run:         runList,
}

func runList(args []string) {

	repo, err := skybinrepo.Open()
	if err != nil {
		log.Fatal(err)
	}

	files, err := repo.ListFiles()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
}
