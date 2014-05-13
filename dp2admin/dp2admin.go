package main

import (
	"fmt"
	"log"
	"os"

	"github.com/daisy-consortium/pipeline-cli-go/cli"
)

func main() {
	log.SetFlags(log.Lshortfile)
	cnf := cli.NewConfig()
	// proper error handlign missing

	link, err := cli.NewLink(cnf)

	if err != nil {
		fmt.Printf("Error connecting to the pipeline webservice:\n\t%v\n", err)
		os.Exit(-1)
	}

	comm, err := cli.NewCli("dp2admin", link)
	comm.WithScripts = false
	if err != nil {
		fmt.Printf("Error creating client:\n\t%v\n", err)
	}

	cli.AddHaltCommand(comm, *link)
	comm.AddClientListCommand(*link)
	comm.AddNewClientCommand(*link)
	comm.AddDeleteClientCommand(*link)
	comm.AddModifyClientCommand(*link)
	comm.AddClientCommand(*link)
	comm.AddPropertyListCommand(*link)
	comm.AddSizesCommand(*link)

	err = comm.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("Error:\n\t%v\n", err)
	}
}
