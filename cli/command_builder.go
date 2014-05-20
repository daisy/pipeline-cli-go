package cli

import (
	"text/template"

	"github.com/capitancambio/go-subcommand"
)

//Convinience interface for building commands
type call func(...interface{}) (interface{}, error)

//commandBuilder builds commands in a reusable way
type commandBuilder struct {
	name     string //Command name
	desc     string //Command description
	linkCall call   //function to call in order to execute the command
	template string //Name of the template used to print the output
}

//Creates a new commandBuilder
func newCommandBuilder(name, desc string) *commandBuilder {
	return &commandBuilder{name: name, desc: desc}
}

//Sets the call to be wrapped within the command
func (c *commandBuilder) withCall(fn call) *commandBuilder {
	c.linkCall = fn
	return c
}

//Sets the template to be used as command output
func (c *commandBuilder) withTemplate(template string) *commandBuilder {
	c.template = template
	return c
}

//builds the commands and adds it to the cli
func (c *commandBuilder) build(cli *Cli) (cmd *subcommand.Command) {
	return cli.AddCommand(c.name, c.desc, func(name string, args ...string) error {
		//call the interface
		iArgs := make([]interface{}, 0, len(args))
		for _, arg := range args {
			iArgs = append(iArgs, arg)
		}

		data, err := c.linkCall(iArgs...)
		if err != nil {
			return err
		}
		tmpl, err := template.New("template").Parse(c.template)
		if err != nil {
			return err
		}
		err = tmpl.Execute(cli.Output, data)
		if err != nil {
			return err
		}
		return nil
	})
}

//Builds a command and configures it to expect a job id
func (c *commandBuilder) buildWithId(cli *Cli) (cmd *subcommand.Command) {
	lastId := new(bool)
	cmd = cli.AddCommand(c.name, c.desc, func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := c.linkCall(id)
		if err != nil {
			return err
		}
		tmpl, err := template.New("template").Parse(c.template)
		if err != nil {
			return err
		}
		err = tmpl.Execute(cli.Output, data)
		if err != nil {
			return err
		}
		return nil
	})

	addLastId(cmd, lastId)
	return
}
