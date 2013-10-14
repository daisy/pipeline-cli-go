package main

import (
	"fmt"
	"github.com/daisy-consortium/pipeline-cli-go/cli"
	"log"
	"os"
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

	comm, err := cli.NewCli("dp2", link)

	if err != nil {
		fmt.Printf("Error creating client:\n\t%v\n", err)
		os.Exit(-1)
	}
	comm.WithScripts = true

	cli.AddJobStatusCommand(comm, *link)
	cli.AddDeleteCommand(comm, *link)
	cli.AddResultsCommand(comm, *link)
	cli.AddJobsCommand(comm, *link)
	cli.AddLogCommand(comm, *link)
	cli.AddHaltCommand(comm, *link)
	cli.AddVersionCommand(comm, link)

	err = comm.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("Error:\n\t%v\n", err)
	}
}
