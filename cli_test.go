package main

import (
	"fmt"
	"github.com/capitancambio/go-subcommand"
	"github.com/daisy-consortium/pipeline-clientlib-go"
	"io/ioutil"
	"log"
	"testing"
)

var SCRIPT pipeline.Script = pipeline.Script{
	Nicename:    "test-to-test",
	Description: "Mocked script",
	Homepage:    "daisy.org",
	Href:        "daisy.org/test",
	Id:          "test",
	Options: []pipeline.Option{
		pipeline.Option{
			Required:  true,
			Sequence:  false,
			Name:      "test-opt",
			Ordered:   false,
			Mediatype: "xml",
			Desc:      "I'm a test option",
			Type:      "anyFileURI",
		},
		pipeline.Option{
			Required:  false,
			Sequence:  false,
			Name:      "another-opt",
			Ordered:   false,
			Mediatype: "xml",
			Desc:      "I'm a test option",
			Type:      "boolean",
		},
	},
	Inputs: []pipeline.Input{
		pipeline.Input{
			Desc:      "input port",
			Mediatype: "application/x-dtbook+xml",
			Name:      "source",
			Sequence:  true,
		},
		pipeline.Input{
			Desc:      "input port not seq",
			Mediatype: "application/x-dtbook+xml",
			Name:      "single",
			Sequence:  false,
		},
	},
}

var in1, in2 = "tmp/dir1/file.xml", "tmp/dir2/file.xml"

func TestParseInputs(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	inputs := fmt.Sprintf("%v,%v", in1, in2)
	urls, err := pathToUri(inputs, ",", "")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if urls[0].String() != in1 {
		t.Errorf("Url 1 is not formatted %v", urls[0].String())
	}

	if urls[1].String() != in2 {
		t.Errorf("Url 2 is not formatted %v", urls[1].String())
	}
}

func TestParseInputsBased(t *testing.T) {
	inputs := fmt.Sprintf("%v,%v", in1, in2)
	urls, err := pathToUri(inputs, ",", "/mydata/")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	//println(urls[0].String())
	if urls[0].String() != "file:///mydata/"+"tmp/dir1/file.xml" {
		t.Errorf("Url 1 is not formated %v", urls[0].String())
	}

	if urls[1].String() != "file:///mydata/"+"tmp/dir2/file.xml" {
		t.Errorf("Url 1 is not formated %v", urls[1].String())
	}
}
func TestScriptToCommand(t *testing.T) {
	parser := subcommand.NewParser("prog")
	cli := Cli{Parser: parser}
	comm, err := scriptToCommand(&cli, SCRIPT)
	if err != nil {
		t.Error("Unexpected error")
	}
	//parser.Parse([]string{"test","--i-source","value"})
	_, err = parser.Parse([]string{"test", "--i-source", "./tmp/file", "--i-single", "./tmp/file2", "--x-test-opt", "./myfile.xml", "--x-another-opt", "true"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if comm.jobRequest.Script != "test" {
		t.Error("script not set")
	}
	if comm.jobRequest.Inputs["source"][0].String() != "./tmp/file" {
		t.Errorf("Input source not set %v", comm.jobRequest.Inputs["source"][0].String())
	}
	if comm.jobRequest.Inputs["single"][0].String() != "./tmp/file2" {
		t.Errorf("Input source not set %v", comm.jobRequest.Inputs["source"][0].String())
	}
	if comm.jobRequest.Options["test-opt"][0] != "./myfile.xml" {
		t.Errorf("Option test opt not set %v", comm.jobRequest.Options["test-opt"][0])
	}

	if comm.jobRequest.Options["another-opt"][0] != "true" {
		t.Errorf("Option test opt not set %v", comm.jobRequest.Options["another-opt"][0])
	}
}
func TestCliRequiredOptions(t *testing.T) {
	cli, err := NewCli("testprog", "[]", PipelineLink{pipeline: &PipelineTest{false, 0}})
	if err != nil {
		t.Error("Unexpected error")
	}
	cli.AddScripts([]pipeline.Script{SCRIPT})
	//parser.Parse([]string{"test","--i-source","value"})
	err = cli.Run([]string{"test", "--i-source", "./tmp/file", "--i-single", "./tmp/file2", "--x-another-opt", "true"})
	if err == nil {
		t.Errorf("Missing required option wasn't thrown")
	}
	err = cli.Run([]string{"./tmp/file", "--i-single", "./tmp/file2", "--x-another-opt", "true"})
	if err == nil {
		t.Errorf("Missing required input wasn't thrown")
	}
}
func TestCliNonRequiredOptions(t *testing.T) {
	cli, err := NewCli("testprog", "[]", PipelineLink{pipeline: &PipelineTest{false, 0}})
	if err != nil {
		t.Error("Unexpected error")
	}
	cli.AddScripts([]pipeline.Script{SCRIPT})
	//parser.Parse([]string{"test","--i-source","value"})
	err = cli.Run([]string{"test", "--i-source", "./tmp/file", "--i-single", "./tmp/file2", "--x-test-opt", "./myfile.xml"})
	if err != nil {
		t.Errorf("Non required option threw an error")
	}
}

type MockExecutor struct {
	Visited bool
}

func (m *MockExecutor) Execute(job JobRequest) (ch chan Message, err error) {
        ch=make(chan Message)
	m.Visited = true
	close(ch)
	return
}

func TestCliExecuteRequest(t *testing.T) {
	mock := &MockExecutor{false}
	cli := Cli{
		Parser:   subcommand.NewParser("prog"),
		Executor: mock,
	}
	cli.AddScripts([]pipeline.Script{SCRIPT})
	//parser.Parse([]string{"test","--i-source","value"})
	err := cli.Run([]string{"test", "--i-source", "./tmp/file", "--i-single", "./tmp/file2", "--x-test-opt", "./myfile.xml"})
	if err != nil {
		t.Error("Unexpected error")
	}
	if !mock.Visited {
		t.Error("Executor not executed!")
	}
}

func TestGetBasePath(t *testing.T) {
	//return os.Getwd()
	basePath := getBasePath(true)
	if len(basePath) == 0 {
		t.Error("Base path is 0")
	}
	if basePath[len(basePath)-1] != "/"[0] {
		t.Error("Last element of the basePath should be /")
	}
	basePath = getBasePath(false)
	if len(basePath) != 0 {
		t.Errorf("Base path len is !=0: %v", basePath)
	}
}
