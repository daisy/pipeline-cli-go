package main

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"os"
	"text/template"
)

const (
	MainHelpTemplate = `Usage {{.Name}} [GLOBAL_OPTIONS] command [COMMAND_OPTIONS]

Script commands:

        {{range .Scripts}}{{.Name}}             {{.Description}}
        {{end}}
General commands:

        {{range .StaticCommands}}{{.Name}}             {{.Description}}
        {{end}}

List of global options:                 {{.Name}} help -g
Detailed help for a single command:     {{.Name}} help COMMAND
`
)

//Cli is a subcommand that differenciates between script commands and regular commands just to treat them correctly during
//the help display
type Cli struct {
	*subcommand.Parser
	Scripts        []*subcommand.Command
	StaticCommands []*subcommand.Command
}

func NewCli(name string, link PipelineLink) (cli *Cli, err error) {
	cli = &Cli{
		Parser: subcommand.NewParser(name),
	}
	cli.Parser.SetHelp("help", "Help description", Helper(cli, link))

	return
}

//Adds the command to the cli and stores the it into the scripts list
func (c *Cli) AddScriptCommand(name, desc string, fn func(string, ...string) error ) *subcommand.Command {
	cmd := c.Parser.AddCommand(name, desc, fn)
	c.Scripts = append(c.Scripts, cmd)
	return cmd
}

//Adds a static command to the cli and keeps track of it for the displaying the help
func (c *Cli) AddCommand(name, desc string, fn func(string, ...string) error) *subcommand.Command {
	cmd := c.Parser.AddCommand(name, desc, fn)
	c.StaticCommands = append(c.StaticCommands, cmd)
	return cmd
}
func Helper(cli *Cli, link PipelineLink) func(string, ...string) error {
	return func(help string, args ...string) error {
		printHelp(*cli, args...)
		return nil
	}
}

func printHelp(cli Cli, args ...string) {
	if len(args) == 0 {
		tmpl, err := template.New("mainHelp").Parse(MainHelpTemplate)
		if err != nil {
			println(err.Error())
			panic("Error compiling help template")
		}
		tmpl.Execute(os.Stdout, cli)
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

//func getLongestName(scripts []*subcommand.Command) int {
//max := -1
//for _, s := range scripts {
//if max < len(s.Name) {
//max = len(s.Name)
//}
//}
//return max
//}

//func (c *Cli) addScript(script pipeline.Script) error {
//command, err := scriptToCommand(c, script)
//c.Scripts = append(c.Scripts, command)
//return err
//}

func (c *Cli) Run(args []string) error {
	_, err := c.Parser.Parse(args)
	return err
}
