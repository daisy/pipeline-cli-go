package main

import (
	"bitbucket.org/kardianos/osext"
	"bufio"
	"fmt"
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

	link, err := NewLink(cnf)

	if err != nil {
		panic(fmt.Sprintf("Error connecting to the pipeline webservice:\n\t%v\n", err))
	}

	cli, err := NewCli("dp2", *link)
	if err != nil {
		panic(fmt.Sprintf("Error creating client:\n\t%v\n", err))
	}
	scripts, err := link.Scripts()
	if err != nil {
		panic(fmt.Sprintf("Error loading scripts:\n\t%v\n", err))
	}
	cli.AddScripts(scripts, *link, cnf.Local)

	AddJobStatusCommand(cli, *link)
	AddDeleteCommand(cli, *link)
	AddResultsCommand(cli, *link)
	AddJobsCommand(cli, *link)

	err = cli.Run(os.Args[1:])
	if err != nil {
		panic(fmt.Sprintf("Error:\n\t%v\n", err))
	}
}

func loadConfig() (cnf Config, err error) {
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
	cnf, err = NewConfig(r)
	return
}
