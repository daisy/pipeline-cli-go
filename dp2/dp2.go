package main

import (
	"fmt"
	"github.com/daisy-consortium/pipeline-cli-go/cli"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Lshortfile)
	cnf := cli.GetConfig()
	// proper error handlign missing

	//if !cnf[cli.DEBUG].(bool) {
	log.SetOutput(ioutil.Discard)
	//}

	link, err := cli.NewLink(cnf)

	if err != nil {
		panic(fmt.Sprintf("Error connecting to the pipeline webservice:\n\t%v\n", err))
	}

	comm, err := cli.NewCli("dp2", link)
	if err != nil {
		panic(fmt.Sprintf("Error creating client:\n\t%v\n", err))
	}
	scripts, err := link.Scripts()
	if err != nil {
		panic(fmt.Sprintf("Error loading scripts:\n\t%v\n", err))
	}
	comm.AddScripts(scripts, link)

	cli.AddJobStatusCommand(comm, *link)
	cli.AddDeleteCommand(comm, *link)
	cli.AddResultsCommand(comm, *link)
	cli.AddJobsCommand(comm, *link)
	cli.AddLogCommand(comm, *link)
	cli.AddHaltCommand(comm, *link)

	err = comm.Run(os.Args[1:])
	if err != nil {
		panic(fmt.Sprintf("Error:\n\t%v\n", err))
	}
}
