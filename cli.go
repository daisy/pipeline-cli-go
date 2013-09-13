package main

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"strings"
	//"github.com/daisy-consortium/pipeline-clientlib-go"
)

//Cli is a subcommand that differenciates between script commands and regular commands just to treat them correctly during
//the help display
type Cli struct {
	*subcommand.Parser
	Scripts []subcommand.Command
}

func NewCli(name string, link PipelineLink) (cli *Cli, err error) {
	cli = &Cli{
		Parser: subcommand.NewParser(name),
	}
	cli.Parser.SetHelp("help", "Help description", Helper(cli, link))

	return
}

//Adds the command to the cli and stores the it into the scripts list
func (c *Cli) AddScriptCommand(name, desc string, fn func(string, ...string)) *subcommand.Command {
	cmd := c.Parser.AddCommand(name, desc, fn)
	c.Scripts = append(c.Scripts, *cmd)
	return cmd
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

func getLongestName(scripts []subcommand.Command) int {
	max := -1
	for _, s := range scripts {
		if max < len(s.Name) {
			max = len(s.Name)
		}
	}
	return max
}

//func (c *Cli) addScript(script pipeline.Script) error {
//command, err := scriptToCommand(c, script)
//c.Scripts = append(c.Scripts, command)
//return err
//}

func (c *Cli) Run(args []string) error {
	_, err := c.Parser.Parse(args)

	if err != nil {
		return err
	}
	return err
}

