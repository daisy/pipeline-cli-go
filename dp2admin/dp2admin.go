package main

import (
	"fmt"
	"github.com/daisy-consortium/pipeline-cli-go/cli"
	"io/ioutil"
	"log"
	"os"
)

const (
	CONFIG_FILE = "config.yml"
)

func main() {
	log.SetFlags(log.Lshortfile)
	// proper error handlign missing
	cnf := cli.NewConfig()
	if !cnf.Debug {
		log.SetOutput(ioutil.Discard)
	}

	link, err := cli.NewLink(cnf)

	if err != nil {
		panic(fmt.Sprintf("Error connecting to the pipeline webservice:\n\t%v\n", err))
	}

	comm, err := cli.NewCli("dp2admin", *link)
	if err != nil {
		panic(fmt.Sprintf("Error creating client:\n\t%v\n", err))
	}

	cli.AddHaltCommand(comm, *link)
	comm.AddClientListCommand(*link)
	comm.AddNewClientCommand(*link)
	comm.AddDeleteClientCommand(*link)
	comm.AddModifyClientCommand(*link)
	comm.AddClientCommand(*link)
	comm.AddPropertyListCommand(*link)

	err = comm.Run(os.Args[1:])
	if err != nil {
		panic(fmt.Sprintf("Error:\n\t%v\n", err))
	}
}
