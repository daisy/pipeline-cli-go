package main

import (
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//Represents the job request
type JobRequest struct {
	Script     string               //Script id to call
	Options    map[string][]string  //Options for the script
	Inputs     map[string][]url.URL //Input ports for the script
	Data       string               //Data to send with the job request
	Verbose    bool                 //If true this request should return the job's messages
	Persitent  bool                 //Do not delete the job once it's done
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
		if _,err :=scriptToCommand(s,c,link,isLocal); err != nil {
			return err
		}
	}
	return nil
}

//Adds the command and flags to be able to call the script to the cli
func scriptToCommand(script pipeline.Script, cli *Cli, link PipelineLink, isLocal bool) (req JobRequest, err error) {
        jobRequest := newJobRequest()
	jobRequest.Script = script.Id
	basePath := getBasePath(isLocal)

	command := cli.AddScriptCommand(script.Id, script.Description, func(string, ...string) {
		messages, err := link.Execute(*jobRequest)
		if err != nil {
			//TODO: subcommands to return errors
			println("Got:", err.Error())
		}
		for msg := range messages {
			if msg.Error != nil {
				err = msg.Error
				//TODO: subcommands to return errors
				println("Got:", err.Error())
				break
			}
			println(msg.String())
		}
	})

	for _, input := range script.Inputs {
		command.AddOption("i-"+input.Name, "", input.Desc, inputFunc(jobRequest, basePath)).Must(true)
	}

	for _, option := range script.Options {
		//desc:=option.Desc+
		command.AddOption("x-"+option.Name, "", option.Desc, optionFunc(jobRequest, basePath, option.Type)).Must(option.Required)
	}
	return  *jobRequest, nil
}

//Returns a function that fills the request info with the subcommand option name
//and value
func inputFunc(req *JobRequest, basePath string) func(string, string) {
	return func(name, value string) {
		var err error
		req.Inputs[name[2:]], err = pathToUri(value, ",", basePath)
		if err != nil {
			panic(err)
		}
	}
}

//Returns a function that fills the request option with the subcommand option name
//and value
func optionFunc(req *JobRequest, basePath string, optionType string) func(string, string) {
	return func(name, value string) {
		name = name[2:]
		if optionType == "anyFileURI" || optionType == "anyDirURI" {
			urls, err := pathToUri(value, ",", basePath)
			if err != nil {
				panic(err)
			}
			for _, url := range urls {
				req.Options[name] = append(req.Options[name], url.String())
			}
		} else {
			req.Options[name] = []string{value}
		}
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
