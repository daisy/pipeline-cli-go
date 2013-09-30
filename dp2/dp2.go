package main

import (
	"bitbucket.org/kardianos/osext"
	"bufio"
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
	cnf, err := loadConfig()
	if !cnf.Debug {
		log.SetOutput(ioutil.Discard)
	}
	if err != nil {
		panic(fmt.Sprintf("Error loading configuaration file:\n\t%v\n", err))
	}

	link, err := cli.NewLink(cnf)

	if err != nil {
		panic(fmt.Sprintf("Error connecting to the pipeline webservice:\n\t%v\n", err))
	}

	comm, err := cli.NewCli("dp2", *link)
	if err != nil {
		panic(fmt.Sprintf("Error creating client:\n\t%v\n", err))
	}
	scripts, err := link.Scripts()
	if err != nil {
		panic(fmt.Sprintf("Error loading scripts:\n\t%v\n", err))
	}
	comm.AddScripts(scripts, *link, cnf.Local)

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

func loadConfig() (cnf cli.Config, err error) {
	basePath, err := osext.ExecutableFolder()
	if err != nil {
		return
	}

	fd, err := os.Open(basePath + CONFIG_FILE)
	defer fd.Close()
	if err != nil {
		return
	}
	r := bufio.NewReader(fd)
	cnf, err = cli.NewConfig(r)
	return
}
