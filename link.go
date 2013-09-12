package main

import (
	"github.com/daisy-consortium/pipeline-clientlib-go"
        "time"
)
//waiting time for getting messages
var MSG_WAIT=200*time.Millisecond

//Convinience for testing
type PipelineApi interface {
	Alive() (alive pipeline.Alive, err error)
	Scripts() (scripts pipeline.Scripts, err error)
	Script(id string) (script pipeline.Script, err error)
        JobRequest(newJob pipeline.JobRequest) (job pipeline.Job , err error)
        ScriptUrl(id string) (string)
        Job (string,int) (pipeline.Job,error)
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
                config: conf,
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
		return err
	}
	pLink.Version = alive.Version
	pLink.Mode = alive.Mode
	pLink.Authentication = alive.Authentication
	return nil
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

func (p PipelineLink) Execute(jobReq JobRequest) (messages chan string, err error) {
        req,err:=jobRequestToPipeline(jobReq,p)
        if err!=nil{
                return
        }
        _,err=p.pipeline.JobRequest(req)
        if err!=nil{
                return
        }
	return
}


func getAsyncMessages(p PipelineLink,jobId string,messages chan string,err chan error) {
        msgNum:=0
        for{
                job,errQ:=p.pipeline.Job(jobId,msgNum)
                if errQ!=nil{
                        err <- errQ
                        return
                }
                for _,msg := range job.Messages{
                        msgNum=msg.Sequence
                        messages <- msg.Content
                }
                if job.Status=="DONE" || job.Status=="ERROR" || job.Status=="VALID"{
                       close(messages)
                       return
                }
                time.Sleep(MSG_WAIT)
        }

}

func jobRequestToPipeline(req JobRequest,p PipelineLink) (pReq pipeline.JobRequest,err error) {
        href:=p.pipeline.ScriptUrl(req.Script)
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
