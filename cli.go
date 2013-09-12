package main

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"net/url"
	"os"
	"path/filepath"
	"strings"
        "log"
)

type Cli struct {
	Name         string
	OutputFlag   string
	Parser       *subcommand.Parser
	Executor     JobExecutor
	Scripts      []*JobRequestCommand
	Config       Config
	execFunction func() error
}

type JobExecutor interface {
	Execute(JobRequest) (chan string, error)
}

func NewCli(name, outputFlag string, link PipelineLink) (cli *Cli, err error) {
	cli = &Cli{
		Name:       name,
		OutputFlag: outputFlag,
		Parser:     subcommand.NewParser(name),
		Config:     link.config,
		Executor:   link,
	}
	cli.Parser.SetHelp("help", "Help description", Helper(cli, link))

	return
}

func Helper(cli *Cli, link PipelineLink) func(string, ...string) {
	return func(help string, args ...string) {
		printHelp(*cli, args...)
	}
}

func printHelp(cli Cli, args ...string) {
	scripts := cli.Scripts
	if len(args) == 0 {
		fmt.Printf("Usage %v [GLOBAL_OPTIONS] command [COMMAND_OPTIONS]\n", cli.Name)
		fmt.Printf("\nScript commands:\n\n")
		maxLen := getLongestName(scripts)
		for _, s := range scripts {
			fmt.Printf("%v%v%v\n", s.Name, strings.Repeat(" ", maxLen-len(s.Name)+4), s.Description)
		}

		fmt.Printf("\nGeneral commands:\n\n")

		fmt.Printf("List of global options:\t\t\t%v help -g\n", cli.Name)
		fmt.Printf("Detailed help for a single command:\t%v help COMMAND\n", cli.Name)
	} else {
		fmt.Printf("Usage %v [GLOBAL_OPTIONS] %v [COMMAND_OPTIONS]\n", args[0], cli.Name)
		c := cli.Parser.Commands[args[0]]
		fmt.Printf("%v\t\t%v\n", c.Name, c.Description)
		fmt.Printf("\n")
		for _, flag := range c.Flags() {
			fmt.Printf("%v\n", flag)
		}
	}
}

func getLongestName(scripts []*JobRequestCommand) int {
	max := -1
	for _, s := range scripts {
		if max < len(s.Name) {
			max = len(s.Name)
		}
	}
	return max
}

func (c *Cli) AddScripts(scripts []pipeline.Script) error {
	for _, s := range scripts {
		if err := c.addScript(s); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cli) addScript(script pipeline.Script) error {
	command, err := scriptToCommand(c, script)
	c.Scripts = append(c.Scripts, command)
	return err
}

func (c *Cli) Run(args []string) error {
	_, err := c.Parser.Parse(args)

	if err != nil {
		return err
	}
	if c.execFunction != nil {
		return c.execFunction()
	}
	return err
}

type JobRequest struct {
	Script  string
	Options map[string][]string
	Inputs  map[string][]url.URL
	Data    string
}

func newJobRequest() *JobRequest {
	return &JobRequest{
		Options: make(map[string][]string),
		Inputs:  make(map[string][]url.URL),
	}
}

type JobRequestCommand struct {
	subcommand.Command
	jobRequest *JobRequest
}

//Splits the input chain (using ,) to a slice of url's. If basePath is not empty will basify the urls to it
//
func pathToUri(paths string, separator string, basePath string) (urls []url.URL, err error) {
	var urlBase *url.URL

	if basePath != "" {
		urlBase, err = url.Parse("file:" + basePath)
	}
	if err != nil {
		return nil, err
	}
	inputs := strings.Split(paths, ",")
	for _, input := range inputs {
		var urlInput *url.URL
		if basePath != "" {
			urlInput, err = url.Parse(filepath.ToSlash(input))
			if err != nil {
				return nil, err
			}
			urlInput = urlBase.ResolveReference(urlInput)
		} else {
			//TODO is opaque really apropriate?
			urlInput = &url.URL{
				Opaque: filepath.ToSlash(input),
			}
		}
		urls = append(urls, *urlInput)
	}
	//clean
	return
}
func scriptToCommand(cli *Cli, script pipeline.Script) (jobRequestCommand *JobRequestCommand, err error) {
	jobRequest := newJobRequest()
	jobRequest.Script = script.Id
	command := cli.Parser.AddCommand(script.Id, script.Description, func(string, ...string) {
		cli.execFunction = func() error {
			_, err := cli.Executor.Execute(*jobRequest)
			return err
		}
	})
        log.Printf("cnf: %+v\n",cli.Config)
	basePath := getBasePath(cli.Config.Local)
	for _, input := range script.Inputs {
		command.AddOption("i-"+input.Name, "", input.Desc, inputFunc(input, jobRequest, basePath)).Must(true)
	}

	for _, option := range script.Options {
		//desc:=option.Desc+
		command.AddOption("x-"+option.Name, "", option.Desc, optionFunc(option, jobRequest, basePath)).Must(option.Required)
	}
	return &JobRequestCommand{*command, jobRequest}, nil
}

func inputFunc(input pipeline.Input, req *JobRequest, basePath string) func(string, string) {
	return func(name, value string) {
		var err error
		req.Inputs[name[2:]], err = pathToUri(value, ",", basePath)
		if err != nil {
			panic(err)
		}
	}
}

func optionFunc(option pipeline.Option, req *JobRequest, basePath string) func(string, string) {
	return func(name, value string) {
		name = name[2:]
		if option.Type == "anyFileURI" || option.Type == "anyDirURI" {
			urls, err := pathToUri(value, ",", basePath)
			if err != nil {
				panic(err)
			}
			for _, url := range urls {
				req.Options[name] = append(req.Options[name], url.String())
			}
		} else {
			req.Options[name] = []string{value}
		}
	}
}

func getBasePath(isLocal bool) string {
	if isLocal {
		base, err := os.Getwd()
		if err != nil {
                        panic("Error while getting current directory:"+err.Error())
		}
		return base+"/"
	} else {
		return ""
	}
}
