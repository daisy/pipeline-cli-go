package cli

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	MAIN_HELP_TEMPLATE = `
Usage {{.Name}} [GLOBAL_OPTIONS] command [COMMAND_OPTIONS] [PARAMS]

{{if .Scripts}}
Script commands:

        {{range .Scripts}}{{.Name}}             {{.Description}}
        {{end}}
{{end}}
General commands:

        {{range .StaticCommands}}{{.Name}}             {{.Description}}
        {{end}}

List of global options:                 {{.Name}} help -g
Detailed help for a single command:     {{.Name}} help COMMAND
`
	//TODO: Check if required options to write/ignore []
	COMMAND_HELP_TEMPLATE = `
Usage: {{.Parent.Name}} [GLOBAL_OPTIONS] {{.Name}} [OPTIONS]  {{ .Params }}

{{.Description}}

Options:
{{range .Flags }}       {{.}}
{{end}}

`

//{{range .Flags}}{{if len .Short}}-{{.Short}},{{end}}--{{.Long}}{{if isOption .}}   {{upper .Long}}{{end}}       {{.Description}}{{end}}
)

//Cli is a subcommand that differenciates between script commands and regular commands just to treat them correctly during
//the help display
type Cli struct {
	*subcommand.Parser
	Scripts        []*ScriptCommand
	StaticCommands []*subcommand.Command
}

type ScriptCommand struct {
	*subcommand.Command
	req *JobRequest
}

//Creates a new CLI with a name and pipeline link to perform queries
func NewCli(name string, link *PipelineLink) (cli *Cli, err error) {
	cli = &Cli{
		Parser: subcommand.NewParser(name),
	}
	cli.Parser.SetHelp("help", "Help description", func(help string, args ...string) error {
		return printHelp(*cli, args...)
	})
	cli.OnCommand(func() error {
		if err = link.Init(); err != nil {
			return err
		}
		if !link.IsLocal() {
			for _, cmd := range cli.Scripts {

				cmd.addDataOption()
			}
		}
		return nil
	})
	cli.addConfigOptions(link.config)
	return
}

func (c *Cli) addConfigOptions(conf Config) {
	//TODO make config a map instead of a struct?
	for option, desc := range config_descriptions {
		c.AddOption(option, "", fmt.Sprintf("%v (default %v)", desc, conf[option]), func(optName string, value string) error {
			switch conf[optName].(type) {
			case int:
				val, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("option %v must be a numeric value (found %v)", optName, value)
				}
				conf[optName] = val
			case bool:
				switch {
				case value == "true":
					conf[optName] = true
				case value == "false":
					conf[optName] = false
				default:
					return fmt.Errorf("option %v must be true or false (found %v)", optName, value)
				}

			case string:
				conf[optName] = value

			}
			conf.UpdateDebug()
			return nil
		})
	}

	c.AddOption("file", "f", "Alternative configuration file", func(string, filePath string) error {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf(err.Error())
			return fmt.Errorf("File not found %v", filePath)
		}
		return conf.FromYaml(file)
	})
}

//Adds the command to the cli and stores the it into the scripts list
func (c *Cli) AddScriptCommand(name, desc string, fn func(string, ...string) error, request *JobRequest) *subcommand.Command {
	cmd := c.Parser.AddCommand(name, desc, fn)
	c.Scripts = append(c.Scripts, &ScriptCommand{cmd, request})
	return cmd
}

//Adds a static command to the cli and keeps track of it for the displaying the help
func (c *Cli) AddCommand(name, desc string, fn func(string, ...string) error) *subcommand.Command {
	cmd := c.Parser.AddCommand(name, desc, fn)
	c.StaticCommands = append(c.StaticCommands, cmd)
	return cmd
}

//Runs the client
func (c *Cli) Run(args []string) error {
	//TODO: should I load a default file?
	_, err := c.Parser.Parse(args)
	return err
}

//prints the help
func printHelp(cli Cli, args ...string) error {
	if len(args) == 0 {
		tmpl, err := template.New("mainHelp").Parse(MAIN_HELP_TEMPLATE)
		if err != nil {
			//this is serious stuff panic!!
			println(err.Error())
			panic("Error compiling help template")
		}
		tmpl.Execute(os.Stdout, cli)

	} else {
		if len(args) > 1 {
			return fmt.Errorf("help: only one parameter is accepted. %v found (%v)", len(args), strings.Join(args, ","))
		}
		cmd, ok := cli.Parser.Commands[args[0]]
		if !ok {
			return fmt.Errorf("help: command %v not found ", args[0])
		}
		funcMap := template.FuncMap{
			"upper": strings.ToUpper,
			"isOption": func(flag subcommand.Flag) bool {
				return flag.Type == subcommand.Option
			},
		}
		tmpl, err := template.New("commandHelp").Funcs(funcMap).Parse(COMMAND_HELP_TEMPLATE)
		if err != nil {
			//this is serious stuff panic!!
			println(err.Error())
			panic("Error compiling command help template")
		}
		//cmdFlag := commmandFlag{*cmd, cli.Name}
		tmpl.Execute(os.Stdout, cmd)
	}
	return nil
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
