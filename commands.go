package main

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"os"
	"text/template"
)

var JobStatusTemplate = `
Job Id: {{.Data.Id }}
Status: {{.Data.Status}}
{{if .Verbose}}Messages:
{{range .Data.Messages}}
({{.Sequence}})[{{.Level}}]      {{.Content}}
{{end}}
{{end}}
`

//Convinience struct
type printableJob struct {
	Data    pipeline.Job
	Verbose bool
}

func AddJobStatusCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	printable := &printableJob{
		Data:    pipeline.Job{},
		Verbose: false,
	}
	cmd := cli.AddCommand("status", "Returns the status of a job", func(command string, args ...string) {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error")
		}
		job, err := link.Job(id)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
		}
		tmpl, err := template.New("status").Parse(JobStatusTemplate)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
		}
		printable.Data = job
		err = tmpl.Execute(os.Stdout, printable)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
		}

	})
	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) {
		printable.Verbose = true
	})
	addLastId(cmd, lastId)
}

func addLastId(cmd *subcommand.Command, lastId *bool) {
	cmd.AddSwitch("lastid", "l", "Get id from the last executed job", func(string, string) {
		*lastId = true
	})
}

func AddDeleteCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	cmd := cli.AddCommand("remove", "Removes a job from the pipeline", func(command string, args ...string) {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error")
		}
		ok, err := link.Delete(id)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
		}
		if err != nil {
			println("error", err.Error())
			return
		}
		if ok {
			fmt.Printf("Job %v removed\n", id)
		}

	})
	addLastId(cmd, lastId)
}
func checkId(lastId bool, command string, args ...string) (id string, err error) {
	if len(args) != 1 && !lastId {
		return id, fmt.Errorf("Command %v needs a job id")
	}
	//got it from file
	if lastId {
		id, err = getLastId()
		return
	} else {
		//first arg otherwise
		id = args[0]
		return
	}
}
