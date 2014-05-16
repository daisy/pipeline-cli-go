package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/capitancambio/go-subcommand"
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

	JobListTemplate = `
Job Id          (Nicename)              [STATUS]

{{range .}}{{.Id}}{{if .Nicename }}     ({{.Nicename}}){{end}}  [{{.Status}}]
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

//Convinience struct
type printableJob struct {
	Data    pipeline.Job
	Verbose bool
}
type call func(...interface{}) (interface{}, error)
type CommandBuilder struct {
	name     string
	desc     string
	linkCall call
	template string
}

func NewCommandBuilder(name, desc string) *CommandBuilder {
	return &CommandBuilder{name: name, desc: desc}
}

func (c *CommandBuilder) withCall(fn call) *CommandBuilder {
	c.linkCall = fn
	return c
}

func (c *CommandBuilder) withTemplate(template string) *CommandBuilder {
	c.template = template
	return c
}

func (c *CommandBuilder) build(cli *Cli) (cmd *subcommand.Command) {
	return cli.AddCommand(c.name, c.desc, func(string, ...string) error {
		data, err := c.linkCall()
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
func (c *CommandBuilder) buildWithId(cli *Cli) (cmd *subcommand.Command) {
	lastId := new(bool)
	cmd = cli.AddCommand(c.name, c.desc, func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := c.linkCall(id)
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
	cmd := NewCommandBuilder("status", "Returns the status of the job with id JOB_ID").
		withCall(fn).withTemplate(JobStatusTemplate).
		buildWithId(cli)

	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) error {
		printable.Verbose = true
		return nil
	})
}

func AddDeleteCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	cmd := cli.AddCommand("delete", "Removes a job from the pipeline", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		ok, err := link.Delete(id)
		if err != nil {
			return err
		}
		if ok {
			fmt.Printf("Job %v removed\n", id)
		}

		return nil
	})
	addLastId(cmd, lastId)
}

func AddResultsCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	outputPath := ""
	cmd := cli.AddCommand("results", "Stores the results from a job", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := link.Results(id)
		if err != nil {
			return err
		}

		path, err := zippedDataToFolder(data, outputPath)
		if err != nil {
			return err
		}

		fmt.Printf("Results stored into %v\n", path)

		return nil
	})
	cmd.AddOption("output", "o", "Directory where to store the results", func(name, folder string) error {
		outputPath = folder
		return nil
	}).Must(true)
	addLastId(cmd, lastId)
}

func AddLogCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	outputPath := ""
	cmd := cli.AddCommand("log", "Stores the results from a job", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := link.Log(id)
		if err != nil {
			return err
		}
		outWriter := os.Stdout
		if len(outputPath) > 0 {
			file, err := os.Create(outputPath)
			defer func() { file.Close() }()
			if err != nil {
				return err
			}
			outWriter = file
		}
		_, err = outWriter.Write(data)
		if err != nil {
			return err
		}
		return nil
	})
	cmd.AddOption("output", "o", "Write the log lines into the file provided instead of printing it", func(name, file string) error {
		outputPath = file
		return nil
	})
	addLastId(cmd, lastId)
}
func AddHaltCommand(cli *Cli, link PipelineLink) {
	cli.AddCommand("halt", "Stops the webservice", func(command string, args ...string) error {
		key, err := loadKey()
		if err != nil {
			return err
		}
		err = link.Halt(key)
		if err != nil {
			return err
		}
		fmt.Println("The webservice has been halted")
		return nil
	})
}

func loadKey() (key string, err error) {
	//get temp dir
	path := filepath.Join(os.TempDir(), "dp2key.txt")
	file, err := os.Open(path)
	if err != nil {
		errors.New("Could not find the key file, is the webservice running in this machine?")
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	key = string(bytes)

	return
}

func AddJobsCommand(cli *Cli, link PipelineLink) {
	cli.AddCommand("jobs", "Returns the list of jobs present in the server", func(command string, args ...string) error {
		jobs, err := link.Jobs()
		if err != nil {
			return err
		}
		tmpl, err := template.New("joblist").Parse(JobListTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(os.Stdout, jobs)
		return nil
	})
}

func AddQueueCommand(cli *Cli, link PipelineLink) {
	fn := func(...interface{}) (interface{}, error) {
		return link.Queue()
	}
	NewCommandBuilder("queue", "Shows the execution queue and the job's priorities. ").
		withCall(fn).withTemplate(QueueTemplate).build(cli)
}

func AddMoveUpCommand(cli *Cli, link PipelineLink) {

}

type Version struct {
	*PipelineLink
	CliVersion string
}

func AddVersionCommand(cli *Cli, link *PipelineLink) {
	NewCommandBuilder("version", "Prints the version and authentication information").
		withCall(func(...interface{}) (interface{}, error) {
		return Version{link, VERSION}, nil
	}).withTemplate(VersionTemplate).build(cli)

}

func checkId(lastId bool, command string, args ...string) (id string, err error) {
	if len(args) != 1 && !lastId {
		return id, fmt.Errorf("Command %v needs a job id", command)
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

func addLastId(cmd *subcommand.Command, lastId *bool) {
	cmd.AddSwitch("lastid", "l", "Get id from the last executed job instead of JOB_ID", func(string, string) error {
		*lastId = true
		return nil
	})
	cmd.SetArity(-1, "[JOB_ID]")
}

//Calculates the absolute path in base of cwd and creates the directory
func createAbsoluteFolder(folder string) (absPath string, err error) {
	absPath, err = filepath.Abs(folder)
	if err != nil {
		return
	}
	return absPath, mkdir(absPath)
}

func mkdir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}

func dumpZippedData(data []byte, folder string) error {
	buff := bytes.NewReader(data)
	reader, err := zip.NewReader(buff, int64(len(data)))
	if err != nil {
		return err
	}
	// Iterate through the files in the archive,
	//and store the results
	for _, f := range reader.File {
		//Get the path of the new file
		path := filepath.Join(folder, filepath.Clean(f.Name))
		if err := mkdir(filepath.Dir(path)); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		dest, err := os.Create(path)
		if err != nil {
			return err
		}

		if _, err = io.Copy(dest, rc); err != nil {
			return err
		}

		if err := dest.Close(); err != nil {
			return err
		}

		if err := rc.Close(); err != nil {
			return err
		}

	}
	return nil
}

func zippedDataToFolder(data []byte, folder string) (absPath string, err error) {
	//Create folder
	absPath, err = createAbsoluteFolder(folder)
	err = dumpZippedData(data, absPath)
	return
}
