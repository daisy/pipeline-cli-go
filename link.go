package main

import (
	"fmt"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MSG_WAIT   = 200 * time.Millisecond //waiting time for getting messages
	JAVA_OPTS  = "JAVA_OPTS"
	OH_MY_GOSH = "-Dgosh.args=--noi"
)

//Convinience for testing
type PipelineApi interface {
	Alive() (alive pipeline.Alive, err error)
	Scripts() (scripts pipeline.Scripts, err error)
	Script(id string) (script pipeline.Script, err error)
	JobRequest(newJob pipeline.JobRequest, data []byte) (job pipeline.Job, err error)
	ScriptUrl(id string) string
	Job(string, int) (pipeline.Job, error)
	DeleteJob(id string) (bool, error)
	Results(id string) ([]byte, error)
	Jobs() (pipeline.Jobs, error)
	Halt(key string) error
}

//Maintains some information about the pipeline client
type PipelineLink struct {
	pipeline       PipelineApi //Allows access to the pipeline fwk
	config         Config
	Version        string //Framework version
	Authentication bool   //Framework authentication
	Mode           string //Framework mode
}

func NewLink(conf Config) (pLink *PipelineLink, err error) {
	pLink = &PipelineLink{
		pipeline: *pipeline.NewPipeline(conf.Url()),
		config:   conf,
	}
	//assure that the pipeline is up

	err = bringUp(pLink)
	if err != nil {
		return nil, err
	}
	return
}

//checks if the pipeline is up
//otherwise it brings it up and fills the
//link object
func bringUp(pLink *PipelineLink) error {
	alive, err := pLink.pipeline.Alive()
	if err != nil {
		log.Println("Starting the fwk")
		//launch the ws

		err = start(pLink.config)
		if err != nil {
			log.Println("Error in start")
			return err
		}
		//wait til it's up and running
		timeOut := time.After(time.Duration(pLink.config.WSTimeUp) * time.Second)
		//communication
		aliveChan := make(chan pipeline.Alive)
		fmt.Println("Launching the pipeline webservice...")
		go wait(*pLink, aliveChan)
		select {
		case alive = <-aliveChan:
			fmt.Println("The webservice is UP!")
			log.Println("The ws seems to be up")
			//keep on going
		case <-timeOut:
			log.Println("bringUp timed up")
			err = fmt.Errorf("I have been waiting %v seconds for the WS to come up but it did not")
			break
		}
		if err != nil {
			return err
		}
	}
	pLink.Version = alive.Version
	pLink.Mode = alive.Mode
	pLink.Authentication = alive.Authentication
	return nil
}

func wait(link PipelineLink, cAlive chan pipeline.Alive) {
	log.Println("Calling alive")
	for {
		alive, err := link.pipeline.Alive()
		if err != nil {
			log.Printf("retrying...")
			time.Sleep(333 * time.Millisecond)
		} else {
			cAlive <- alive
			break
		}
	}

}

func start(cnf Config) error {
	path := filepath.FromSlash(cnf.ExecLineNix)
	log.Printf("command path %v\n", path)
	cmd := exec.Command(path)
	//TODO: modify current env to append -Dgosh.args=--noi to JAVA_OPTS
	cmd.Env = os.Environ()
	found := false
	for idx, env := range cmd.Env {
		if strings.HasPrefix(env, JAVA_OPTS) {
			found = true
			cmd.Env[idx] = appendOpts(env)
		}
	}
	if !found {
		cmd.Env = append(cmd.Env, appendOpts(JAVA_OPTS+"="))
	}
	log.Printf("Command %+v\n", cmd)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	//cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err := cmd.Start()
	return err
}

func appendOpts(javaOptsVar string) string {
	//just the value
	val := strings.TrimLeft(javaOptsVar, JAVA_OPTS+"=")
	val = strings.Trim(val, `"`)
	result := val + " " + OH_MY_GOSH
	return JAVA_OPTS + `=` + result
}

//ScriptList returns the list of scripts available in the framework
func (p PipelineLink) Scripts() (scripts []pipeline.Script, err error) {
	scriptsStruct, err := p.pipeline.Scripts()
	if err != nil {
		return
	}
	scripts = make([]pipeline.Script, len(scriptsStruct.Scripts))
	//fill the script list with the complete definition
	for idx, script := range scriptsStruct.Scripts {
		scripts[idx], err = p.pipeline.Script(script.Id)
		if err != nil {
			return nil, err
		}
	}
	return scripts, err
}

//Gets the job identified by the jobId
func (p PipelineLink) Job(jobId string) (job pipeline.Job, err error) {
	job, err = p.pipeline.Job(jobId, 0)
	return
}

//Deletes the given job
func (p PipelineLink) Delete(jobId string) (ok bool, err error) {
	ok, err = p.pipeline.DeleteJob(jobId)
	return
}

//Return the zipped results as a []byte
func (p PipelineLink) Results(jobId string) (data []byte, err error) {
	data, err = p.pipeline.Results(jobId)
	return
}
func (p PipelineLink) Jobs() (jobs []pipeline.Job, err error) {
	pJobs, err := p.pipeline.Jobs()
	if err != nil {
		return
	}
	jobs = pJobs.Jobs
	return
}

func (p PipelineLink) Halt(key string) error {
	return p.pipeline.Halt(key)
}

//Convience structure to handle message and errors from the communication with the pipelineApi
type Message struct {
	Message pipeline.Message
	Status  string
	Error   error
}

//Returns a simple string representation of the messages strucutre:
//(index)[LEVEL]        Message content
func (m Message) String() string {
	if m.Message.Content != "" {
		return fmt.Sprintf("(%v)[%v]\t%v", m.Message.Sequence, m.Message.Level, m.Message.Content)
	} else {
		return ""
	}
}

//Executes the job request and returns a channel fed with the job's messages,errors, and status.
//The last message will have no contents but the status of the in which the job finished
func (p PipelineLink) Execute(jobReq JobRequest) (job pipeline.Job, messages chan Message, err error) {
	req, err := jobRequestToPipeline(jobReq, p)
	if err != nil {
		return
	}
	log.Printf("data len exec %v", len(jobReq.Data))
	job, err = p.pipeline.JobRequest(req, jobReq.Data)
	if err != nil {
		return
	}
	messages = make(chan Message)
	if !jobReq.Background {
		go getAsyncMessages(p, job.Id, messages)
	} else {
		close(messages)
	}
	return
}

//Feeds the channel with the messages describing the job's execution
func getAsyncMessages(p PipelineLink, jobId string, messages chan Message) {
	msgNum := 0
	for {
		job, err := p.pipeline.Job(jobId, msgNum)
		if err != nil {
			messages <- Message{Error: err}
			close(messages)
			return
		}
		for _, msg := range job.Messages {
			msgNum = msg.Sequence
			messages <- Message{Message: msg, Status: job.Status}

		}
		if job.Status == "DONE" || job.Status == "ERROR" || job.Status == "VALID" {
			messages <- Message{Status: job.Status}
			close(messages)
			return
		}
		time.Sleep(MSG_WAIT)
	}

}

func jobRequestToPipeline(req JobRequest, p PipelineLink) (pReq pipeline.JobRequest, err error) {
	href := p.pipeline.ScriptUrl(req.Script)
	pReq = pipeline.JobRequest{
		Script: pipeline.Script{Href: href},
	}
	for name, values := range req.Inputs {
		input := pipeline.Input{Name: name}
		for _, value := range values {
			input.Items = append(input.Items, pipeline.Item{Value: value.String()})
		}
		pReq.Inputs = append(pReq.Inputs, input)
	}
	for name, values := range req.Options {
		option := pipeline.Option{Name: name}
		if len(values) > 1 {
			for _, value := range values {
				option.Items = append(option.Items, pipeline.Item{Value: value})
			}
		} else {
			option.Value = values[0]
		}
		pReq.Options = append(pReq.Options, option)

	}
	return
}
