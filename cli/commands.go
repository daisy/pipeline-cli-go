package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
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
)

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
	cmd := cli.AddCommand("status", "Returns the status of the job with id JOB_ID", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		job, err := link.Job(id)
		if err != nil {
			return err
		}
		tmpl, err := template.New("status").Parse(JobStatusTemplate)
		if err != nil {
			return err
		}
		printable.Data = job
		err = tmpl.Execute(os.Stdout, printable)
		if err != nil {
			return err
		}
		return nil
	})
	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) error {
		printable.Verbose = true
		return nil
	})
	addLastId(cmd, lastId)
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

func AddVersionCommand(cli *Cli, link *PipelineLink) {
	cli.AddCommand("version", "Prints the version and authentication information", func(command string, args ...string) error {
		type Version struct {
			*PipelineLink
			CliVersion string
		}

		tmpl, err := template.New("version").Parse(VersionTemplate)
		if err != nil {
			return err
		}

		ver := Version{link, VERSION}
		err = tmpl.Execute(os.Stdout, ver)
		return nil

	})
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

func addLastId(cmd *subcommand.Command, lastId *bool) {
	cmd.AddSwitch("lastid", "l", "Get id from the last executed job instead of JOB_ID", func(string, string) error {
		*lastId = true
		return nil
	})
	cmd.Params = "[JOB_ID]"
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
