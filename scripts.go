package main

import (
	"bitbucket.org/kardianos/osext"
	"errors"
	"fmt"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//set the last id path
var LastIdPath = getLastIdPath()

func getLastIdPath() string {
	path, err := osext.ExecutableFolder()
	if err != nil {
		panic("Couldn't get the executable path")
	}
	return path + string(os.PathSeparator) + ".lastid"
}

//Represents the job request
type JobRequest struct {
	Script     string               //Script id to call
	Nicename   string               //Job's nicename
	Options    map[string][]string  //Options for the script
	Inputs     map[string][]url.URL //Input ports for the script
	Data       []byte               //Data to send with the job request
	Background bool                 //Send the request and return
}

//Creates a new JobRequest
func newJobRequest() *JobRequest {
	return &JobRequest{
		Options: make(map[string][]string),
		Inputs:  make(map[string][]url.URL),
	}
}

//Convinience method to add several scripts to a client
func (c *Cli) AddScripts(scripts []pipeline.Script, link PipelineLink, isLocal bool) error {

	for _, s := range scripts {
		if _, err := scriptToCommand(s, c, link, isLocal); err != nil {
			return err
		}
	}
	return nil
}

//Executes a job request
type jobExecution struct {
	link       PipelineLink
	req        JobRequest
	output     string
	verbose    bool
	persistent bool
}

func (j jobExecution) run() error {
	//manual check of output
	if !j.req.Background && j.output == "" {
		return errors.New("--output option is mandatory if the job is not running in the req.Background")
	}
	if j.req.Background && j.output != "" {
		fmt.Printf("Warning: --output option ignored as the job will run in the background\n")
	}
	storeId := j.req.Background || j.persistent
	//send the job
	job, messages, err := j.link.Execute(j.req)
	if err != nil {
		return err
	}
	fmt.Printf("Job %v sent to the server\n", job.Id)
	//store id if it suits
	if storeId {
		err = storeLastId(job.Id)
		if err != nil {
			return err
		}
	}
	//get realtime messages and status from the webservice
	status := job.Status
	for msg := range messages {
		if msg.Error != nil {
			err = msg.Error
			return err
		}
		//print messages
		if j.verbose {
			println(msg.String())
		}
		status = msg.Status
	}

	if status != "ERROR" {
		//get the data
		if !j.req.Background {
			data, err := j.link.Results(job.Id)
			if err != nil {
				return err
			}
			zippedDataToFolder(data, j.output)
			fmt.Println("Results stored")
			if !j.persistent {
				_, err = j.link.Delete(job.Id)
				if err != nil {
					return err
				}
				fmt.Printf("The job has been deleted from the server\n")
			}
			fmt.Printf("Job finished with status: %v\n", status)
		}

	}
	return nil
}

//Adds the command and flags to be able to call the script to the cli
func scriptToCommand(script pipeline.Script, cli *Cli, link PipelineLink, isLocal bool) (req JobRequest, err error) {
	jobRequest := newJobRequest()
	jobRequest.Script = script.Id
	basePath := getBasePath(isLocal)
	jobRequest.Background = false
	jExec := jobExecution{
		link:    link,
		req:     *jobRequest,
		output:  "",
		verbose: true,
	}
	command := cli.AddScriptCommand(script.Id, script.Description, func(string, ...string) error {
		if err := jExec.run(); err != nil {
			return err
		}
		return nil
	})

	for _, input := range script.Inputs {
		command.AddOption("i-"+input.Name, "", input.Desc, inputFunc(jobRequest, basePath)).Must(true)
	}

	for _, option := range script.Options {
		//desc:=option.Desc+
		command.AddOption("x-"+option.Name, "", option.Desc, optionFunc(jobRequest, basePath, option.Type)).Must(option.Required)
	}

	command.AddOption("nicename", "n", "Set job's nice name", func(name, nice string) error {
		jExec.req.Nicename = nice

		return nil
	})
	command.AddSwitch("quiet", "q", "Do not print the job's messages", func(string, string) error {
		jExec.verbose = false
		return nil
	})
	command.AddSwitch("persistent", "p", "Delete the job after it is executed", func(string, string) error {
		jExec.persistent = true
		return nil
	})

	command.AddSwitch("background", "b", "Sends the job and exits", func(string, string) error {
		jExec.req.Background = true
		return nil
	})
	command.AddOption("output", "o", "Directory where to store the results. This option is mandatory when the job is not executed in the background", func(name, folder string) error {
		jExec.output = folder
		return nil
	})

	if !isLocal {
		command.AddOption("data", "d", "Zip file containing the files to convert", func(name, path string) error {
			file, err := os.Open(path)
			defer func() {
				err := file.Close()
				if err != nil {
					log.Printf("Error closing file %v :%v", path, err.Error())
				}
			}()
			if err != nil {
				return err
			}
			jExec.req.Data, err = ioutil.ReadAll(file)
			log.Printf("data len %v\n", len(jExec.req.Data))
			return nil
		}).Must(true)
	}
	return *jobRequest, nil
}

//Returns a function that fills the request info with the subcommand option name
//and value
func inputFunc(req *JobRequest, basePath string) func(string, string) error {
	return func(name, value string) error {
		var err error
		req.Inputs[name[2:]], err = pathToUri(value, ",", basePath)
		return err
	}
}

//Returns a function that fills the request option with the subcommand option name
//and value
func optionFunc(req *JobRequest, basePath string, optionType string) func(string, string) error {
	return func(name, value string) error {
		name = name[2:]
		if optionType == "anyFileURI" || optionType == "anyDirURI" {
			urls, err := pathToUri(value, ",", basePath)
			if err != nil {
				return err
			}
			for _, url := range urls {
				req.Options[name] = append(req.Options[name], url.String())
			}
		} else {
			req.Options[name] = []string{value}
		}
		return nil
	}
}

//Gets the basepath. If the fwk accepts local uri's (file:///)
//getBasePath os.Getwd() otherwise it returns an empty string
func getBasePath(isLocal bool) string {
	if isLocal {
		base, err := os.Getwd()
		if err != nil {
			panic("Error while getting current directory:" + err.Error())
		}
		return base + "/"
	} else {
		return ""
	}
}

//Accepts several paths separated by separator and constructs the URLs
//relative to base path
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

func storeLastId(id string) error {
	file, err := os.Create(LastIdPath)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
	}()
	if _, err := file.Write([]byte(id)); err != nil {
		return err
	}
	return nil
}

func getLastId() (id string, err error) {
	idBuf, err := ioutil.ReadFile(LastIdPath)
	if err != nil {
		return "", err
	}
	return string(idBuf), nil
}
