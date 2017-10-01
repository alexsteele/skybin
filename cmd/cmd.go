package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type Cmd struct {
	Name        string
	Usage       string
	Description string
	Run         func(args []string)
}

var Commands = []Cmd{
	initCmd,
	storeCmd,
	listCmd,
	getCmd,
	serverCmd,
}

func Usage() {
	fmt.Printf("usage: %s <command> [option...]\n", os.Args[0])
	fmt.Println()
	fmt.Println("\tInteract with a skybin repo")
	fmt.Println()
	fmt.Println("commands:")

	tw := tabwriter.NewWriter(os.Stdout, 0, 5, 5, ' ', 0)

	for _, cmd := range Commands {
		fmt.Fprintf(tw, "\t%s\t%s\n", cmd.Name, cmd.Description)
	}

	tw.Flush()

	fmt.Println()
}
