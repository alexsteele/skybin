package cmd

import (
	"flag"
	"log"
	"skybin/repo"
)

var initCmd = Cmd{
	Name:        "init",
	Usage:       "init",
	Description: "Create a new repo",
	Run:         runInit,
}

func runInit(args []string) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	homeFlag := flags.String("home", "", "Repo home directory")
	flags.Parse(args)

	if len(*homeFlag) > 0 {
		repo.Init(*homeFlag)
	}

	homedir, err := repo.DefaultHomeDir()
	if err != nil {
		log.Fatal("Could not find default home dir: ", err)
	}

	repo.Init(homedir)
}
