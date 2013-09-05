package main

import (
	"bitbucket.org/kardianos/osext"
	"bufio"
	"fmt"
	"os"
)

const (
	CONFIG_FILE = "config.yml"
)

func main() {
	// proper error handlign missing

	cnf, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("Error loading configuaration file:\n\t%v", err))
	}

	link, err := NewLink(cnf)

	if err != nil {
		panic(fmt.Sprintf("Error connecting to the pipeline webservice:\n\t%v", err))
	}

	cli, err := NewCli("dp2", "[DP2]", *link)
	if err != nil {
		panic(fmt.Sprintf("Error creating client:\n\t%v", err))
	}
	scripts, err := link.Scripts()
	if err != nil {
		panic(fmt.Sprintf("Error loading scripts:\n\t%v", err))
	}
	cli.AddScripts(scripts)
	err = cli.Run(os.Args[1:])
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
