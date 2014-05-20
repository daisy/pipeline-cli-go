package cli

import (
	"fmt"
	"os"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

const (
	JobStatusTemplate = `
Job Id: {{.Data.Id }}
Status: {{.Data.Status}}
{{if .Verbose}}Messages:
{{range .Data.Messages}}
({{.Sequence}})[{{.Level}}]      {{.Content}}
{{end}}
{{end}}
`

	JobListTemplate = `Job Id          (Nicename)              [STATUS]
{{range .}}{{.Id}}{{if .Nicename }}	({{.Nicename}}){{end}}	[{{.Status}}]
{{end}}`

	VersionTemplate = `
Client version:                 {{.CliVersion}}         
Pipeline version:               {{.Version}}
Pipeline authentication:        {{.Authentication}}
`

	QueueTemplate = `Job Id 			Priority	Job P.	 Client P.	Rel.Time.	 Since
{{range .}}{{.Id}}	{{.ComputedPriority | printf "%.2f"}}	{{.JobPriority}}	{{.ClientPriority}}	{{.RelativeTime | printf "%.2f"}}	{{.TimeStamp}}
{{end}}`
)

//Convinience struct for printing jobs
type printableJob struct {
	Data    pipeline.Job
	Verbose bool
}

func AddJobStatusCommand(cli *Cli, link PipelineLink) {
	printable := &printableJob{
		Data:    pipeline.Job{},
		Verbose: false,
	}
	fn := func(args ...interface{}) (interface{}, error) {
		job, err := link.Job(args[0].(string))
		if err != nil {
			return nil, err
		}
		printable.Data = job
		return printable, nil
	}
	cmd := newCommandBuilder("status", "Returns the status of the job with id JOB_ID").
		withCall(fn).withTemplate(JobStatusTemplate).
		buildWithId(cli)

	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) error {
		printable.Verbose = true
		return nil
	})
}

func AddDeleteCommand(cli *Cli, link PipelineLink) {
	fn := func(args ...interface{}) (interface{}, error) {
		id := args[0].(string)
		ok, err := link.Delete(id)
		if err == nil && ok {
			cli.Printf("Job %v removed from the server\n", id)
		}
		return ok, err
	}
	newCommandBuilder("delete", "Removes a job from the pipeline").
		withCall(fn).buildWithId(cli)
}

func AddResultsCommand(cli *Cli, link PipelineLink) {
	outputPath := ""
	cmd := newCommandBuilder("results", "Stores the results from a job").
		withCall(func(args ...interface{}) (v interface{}, err error) {
		data, err := link.Results(args[0].(string))
		if err != nil {
			return
		}
		path, err := zippedDataToFolder(data, outputPath)
		if err != nil {
			return
		}
		cli.Printf("Results stored into %v\n", path)
		return
	}).buildWithId(cli)
	cmd.AddOption("output", "o", "Directory where to store the results", func(name, folder string) error {
		outputPath = folder
		return nil
	}).Must(true)
}

func AddLogCommand(cli *Cli, link PipelineLink) {
	outputPath := ""
	fn := func(vals ...interface{}) (ret interface{}, err error) {
		data, err := link.Log(vals[0].(string))
		if err != nil {
			return
		}
		outWriter := cli.Output
		if len(outputPath) > 0 {
			file, err := os.Create(outputPath)
			defer func() {
				cli.Printf("Log written to %s", file.Name())
				file.Close()
			}()
			if err != nil {
				return ret, err
			}
			outWriter = file
		}
		_, err = outWriter.Write(data)
		if err != nil {
			return
		}
		return
	}
	cmd := newCommandBuilder("log", "Stores the results from a job").
		withCall(fn).buildWithId(cli)

	cmd.AddOption("output", "o", "Write the log lines into the file provided instead of printing it", func(name, file string) error {
		outputPath = file
		return nil
	})
}

func AddHaltCommand(cli *Cli, link PipelineLink) {
	fn := func(...interface{}) (val interface{}, err error) {
		key, err := loadKey()
		if err != nil {
			return nil, fmt.Errorf("Coudn't open key file: %s", err.Error())
		}
		err = link.Halt(key)
		if err != nil {
			return
		}
		cli.Printf("The webservice has been halted\n")
		return
	}
	newCommandBuilder("halt", "Stops the webservice").withCall(fn).build(cli)
}

func AddJobsCommand(cli *Cli, link PipelineLink) {
	newCommandBuilder("jobs", "Returns the list of jobs present in the server").
		withCall(func(...interface{}) (interface{}, error) {
		return link.Jobs()
	}).withTemplate(JobListTemplate).build(cli)
}

func AddQueueCommand(cli *Cli, link PipelineLink) {
	fn := func(...interface{}) (queue interface{}, err error) {
		return link.Queue()
	}
	newCommandBuilder("queue", "Shows the execution queue and the job's priorities. ").
		withCall(fn).withTemplate(QueueTemplate).build(cli)
}

func AddMoveUpCommand(cli *Cli, link PipelineLink) {
	fn := func(args ...interface{}) (queue interface{}, err error) {
		return link.MoveUp(args[0].(string))
	}
	newCommandBuilder("moveup", "Moves the job up the execution queue").
		withCall(fn).withTemplate(QueueTemplate).
		buildWithId(cli)

}

func AddMoveDownCommand(cli *Cli, link PipelineLink) {
	fn := func(args ...interface{}) (queue interface{}, err error) {
		return link.MoveDown(args[0].(string))
	}
	newCommandBuilder("movedown", "Moves the job down the execution queue").
		withCall(fn).withTemplate(QueueTemplate).
		buildWithId(cli)

}

type Version struct {
	PipelineLink
	CliVersion string
}

func AddVersionCommand(cli *Cli, link PipelineLink) {
	newCommandBuilder("version", "Prints the version and authentication information").
		withCall(func(...interface{}) (interface{}, error) {
		return Version{link, VERSION}, nil
	}).withTemplate(VersionTemplate).build(cli)

}
