package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"io"
	"os"
	"path/filepath"
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
			return
		}
		job, err := link.Job(id)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
			return
		}
		tmpl, err := template.New("status").Parse(JobStatusTemplate)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
			return
		}
		printable.Data = job
		err = tmpl.Execute(os.Stdout, printable)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
			return
		}

	})
	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) {
		printable.Verbose = true
	})
	addLastId(cmd, lastId)
}

func AddDeleteCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	cmd := cli.AddCommand("delete", "Removes a job from the pipeline", func(command string, args ...string) {
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

func AddResultsCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	outputPath := ""
	cmd := cli.AddCommand("results", "Stores the results from a job", func(command string, args ...string) {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error")
		}
		data, err := link.Results(id)
		if err != nil {
			//TODO subcommand functions to return errors
			println("error", err.Error())
		}
		if err != nil {
			println("error", err.Error())
			return
		}

		path, err := zippedDataToFolder(data, outputPath)
		if err != nil {
			println("error", err.Error())
			return
		}

		fmt.Printf("Results stored into %v\n", path)

	})
	cmd.AddOption("output", "o", "Directory where to store the results", func(name, folder string) {
		outputPath = folder
	}).Must(true)
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

func addLastId(cmd *subcommand.Command, lastId *bool) {
	cmd.AddSwitch("lastid", "l", "Get id from the last executed job", func(string, string) {
		*lastId = true
	})
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
