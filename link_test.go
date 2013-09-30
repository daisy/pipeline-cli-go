package main

import (
	"errors"
	"fmt"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"net/url"
	"testing"
)

var (
	JOB_REQUEST = JobRequest{
		Script:   "test",
		Nicename: "nice",
		Options: map[string][]string{
			SCRIPT.Options[0].Name: []string{"file1.xml", "file2.xml"},
			SCRIPT.Options[1].Name: []string{"true"},
		},
		Inputs: map[string][]url.URL{
			SCRIPT.Inputs[0].Name: []url.URL{
				url.URL{Opaque: "tmp/file.xml"},
				url.URL{Opaque: "tmp/file1.xml"},
			},
			SCRIPT.Inputs[1].Name: []url.URL{
				url.URL{Opaque: "tmp/file2.xml"},
			},
		},
	}
	JOB_1 = pipeline.Job{
		Status: "RUNNING",
		Messages: []pipeline.Message{
			pipeline.Message{
				Sequence: 1,
				Content:  "Message 1",
			},
			pipeline.Message{
				Sequence: 2,
				Content:  "Message 2",
			},
		},
	}
	JOB_2 = pipeline.Job{
		Status: "DONE",
		Messages: []pipeline.Message{
			pipeline.Message{
				Sequence: 3,
				Content:  "Message 3",
			},
		},
	}
)

type PipelineTest struct {
	fail         bool
	count        int
	deleted      bool
	resulted     bool
	backgrounded bool
}

func newPipelineTest(fail bool) *PipelineTest {
	return &PipelineTest{
		fail:         fail,
		count:        0,
		deleted:      false,
		resulted:     false,
		backgrounded: false,
	}
}

func (p PipelineTest) SetCredentials(key, secret string) {
}
func (p *PipelineTest) Alive() (alive pipeline.Alive, err error) {
	if p.fail {
		return alive, errors.New("Error")
	}
	alive.Version = "test"
	alive.Mode = "test"
	alive.Authentication = true
	return
}

func (p *PipelineTest) Scripts() (scripts pipeline.Scripts, err error) {
	if p.fail {
		return scripts, errors.New("Error")
	}
	return pipeline.Scripts{Href: "test", Scripts: []pipeline.Script{pipeline.Script{Id: "test"}, pipeline.Script{Id: "test"}}}, err
}

func (p *PipelineTest) Script(id string) (script pipeline.Script, err error) {
	if p.fail {
		return script, errors.New("Error")
	}
	return SCRIPT, nil

}
func (p *PipelineTest) ScriptUrl(id string) string {
	return "test"
}

func (p *PipelineTest) Job(id string, msgSeq int) (job pipeline.Job, err error) {
	if p.fail {
		return job, errors.New("Error")
	}
	if p.count == 0 {
		p.count++
		return JOB_1, nil
	} else {
		p.count++
		return JOB_2, nil
	}
}

func (p *PipelineTest) JobRequest(newJob pipeline.JobRequest, data []byte) (job pipeline.Job, err error) {
	return
}

func (p *PipelineTest) DeleteJob(id string) (ok bool, err error) {
	p.deleted = true
	return
}

func (p *PipelineTest) Results(id string) (data []byte, err error) {
	p.resulted = true
	return
}
func (p *PipelineTest) Log(id string) (data []byte, err error) {
	return
}
func (p *PipelineTest) Jobs() (jobs pipeline.Jobs, err error) {
	return
}

func (p *PipelineTest) Halt(key string) error {
	return nil
}

func TestBringUp(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(false)}
	err := bringUp(&link)
	if err != nil {
		t.Error("Unexpected error")
	}

	if link.Version != "test" {
		t.Error("Version not set")
	}
	if link.Mode != "test" {
		t.Error("Mode not set")
	}

	if !link.Authentication {
		t.Error("Authentication not set")
	}
}

func TestBringUpFail(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(true)}
	err := bringUp(&link)
	if err == nil {
		t.Error("Expected error is nil")
	}
}

func TestScripts(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(false)}
	list, err := link.Scripts()
	if err != nil {
		t.Error("Unexpected error")
	}
	if len(list) != 2 {
		t.Error("Wrong list size")
	}
	res := list[1]
	exp := SCRIPT
	if exp.Href != res.Href {
		t.Errorf("Scripts decoding failed (Href)\nexpected %v \nresult %v", exp.Href, res.Href)
	}
	if exp.Description != res.Description {
		t.Errorf("Script decoding failed (Description)\nexpected %v \nresult %v", exp.Description, res.Description)
	}
	if exp.Homepage != res.Homepage {
		t.Errorf("Scripts decoding failed (Homepage)\nexpected %v \nresult %v", exp.Homepage, res.Homepage)
	}
	if len(exp.Inputs) != len(res.Inputs) {
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Inputs), len(res.Inputs))
	}
	if len(exp.Options) != len(res.Options) {
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Options), len(res.Options))
	}

}

func TestScriptsFail(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(true)}
	_, err := link.Scripts()
	if err == nil {
		t.Error("Expected error is nil")
	}
}

func TestJobRequestToPipeline(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(false)}
	req, err := jobRequestToPipeline(JOB_REQUEST, link)
	if err != nil {
		t.Error("Unexpected error")
	}
	if req.Script.Href != SCRIPT.Id {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", SCRIPT.Id, req.Script.Href)
	}
	if "nice" != req.Nicename {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "nicename", "nice", req.Nicename)
	}

	if len(req.Inputs) != 2 {
		t.Errorf("Bad input list len %v", len(req.Inputs))
	}
	for i := 0; i < 2; i++ {
		if req.Inputs[i].Name != SCRIPT.Inputs[i].Name {
			t.Errorf("JobRequest input %v to pipeline failed \nexpected %v \nresult %v", i, SCRIPT.Inputs[i].Name, req.Inputs[i].Name)
		}

	}
	if req.Inputs[0].Items[0].Value != JOB_REQUEST.Inputs[req.Inputs[0].Name][0].String() {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Inputs[req.Inputs[0].Name][0].String(), req.Inputs[0].Items[0].Value)
	}
	if req.Inputs[0].Items[1].Value != JOB_REQUEST.Inputs[req.Inputs[0].Name][1].String() {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Inputs[req.Inputs[0].Name][1].String(), req.Inputs[0].Items[1].Value)
	}

	if req.Inputs[1].Items[0].Value != JOB_REQUEST.Inputs[req.Inputs[1].Name][0].String() {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Inputs[req.Inputs[1].Name][0].String(), req.Inputs[1].Items[0].Value)
	}

	if len(req.Options) != 2 {
		t.Errorf("Bad option list len %v", len(req.Inputs))
	}

	if req.Options[0].Name != SCRIPT.Options[0].Name {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", req.Options[0].Name, SCRIPT.Options[0].Name)
	}

	if req.Options[1].Name != SCRIPT.Options[1].Name {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", req.Options[1].Name, SCRIPT.Options[1].Name)
	}
	if req.Options[0].Items[0].Value != JOB_REQUEST.Options[req.Options[0].Name][0] {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Options[req.Options[0].Name][0], req.Options[0].Items[0].Value)
	}
	if req.Options[0].Items[1].Value != JOB_REQUEST.Options[req.Options[0].Name][1] {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Options[req.Options[0].Name][1], req.Options[0].Items[1].Value)
	}

	if len(req.Options[1].Items) != 0 {
		t.Error("Simple option lenght !=0")
	}
	if req.Options[1].Value != JOB_REQUEST.Options[req.Options[1].Name][0] {
		t.Errorf("JobRequest to pipeline failed \nexpected %v \nresult %v", JOB_REQUEST.Options[req.Options[0].Name][1], req.Options[0].Items[1].Value)
	}
}

func TestAsyncMessagesErr(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(true)}
	chMsg := make(chan Message)
	go getAsyncMessages(link, "jobId", chMsg)
	message := <-chMsg
	if message.Error == nil {
		t.Error("Expected error nil")
	}

}

func TestAsyncMessages(t *testing.T) {
	link := PipelineLink{pipeline: newPipelineTest(false)}
	chMsg := make(chan Message)
	var msgs []string
	go getAsyncMessages(link, "jobId", chMsg)
	for msg := range chMsg {
		msgs = append(msgs, msg.Message.Content)
	}
	if len(msgs) != 4 {
		t.Errorf("Wrong message list size %v", len(msgs))
	}

	for i := 1; i != 3; i++ {
		if msgs[i-1] != fmt.Sprintf("Message %v", i) {
			t.Errorf("Wrong message %v", msgs[i-1])
		}
	}
}

func TestAppendOps(t *testing.T) {
	//from empty variable
	res := appendOpts("JAVA_OPTS=")
	javaOptsEmpty := "JAVA_OPTS= " + OH_MY_GOSH
	if javaOptsEmpty != res {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "javaOptsEmpty ", javaOptsEmpty, res)
	}
	//non-empty no quotes
	res = appendOpts("JAVA_OPTS=-Dsomething")
	javaOptsNoQuotes := "JAVA_OPTS=-Dsomething " + OH_MY_GOSH
	if javaOptsNoQuotes != res {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "javaOptsNoQuotes ", javaOptsNoQuotes, res)
	}

	res = appendOpts("JAVA_OPTS=\"-Dsomething -Dandsthelse\"")
	javaOptsQuotes := "JAVA_OPTS=-Dsomething -Dandsthelse " + OH_MY_GOSH
	if javaOptsQuotes != res {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "javaOptsQuotes ", javaOptsQuotes, res)
	}
}

