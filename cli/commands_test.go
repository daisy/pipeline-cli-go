package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

var (
	files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"fold1/gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"fold1/fold2/todo.txt", "Get animal handling licence.\nWrite more examples."},
	}

	queue = []pipeline.QueueJob{
		pipeline.QueueJob{
			Id:               "job1",
			ClientPriority:   "high",
			JobPriority:      "low",
			ComputedPriority: 1.555555,
			RelativeTime:     0.577777,
			TimeStamp:        1400237879517,
		},
	}
	queueLine = []string{
		queue[0].Id,
		fmt.Sprintf("%.2f", queue[0].ComputedPriority),
		queue[0].JobPriority,
		queue[0].ClientPriority,
		fmt.Sprintf("%.2f", queue[0].RelativeTime),
		fmt.Sprintf("%d", queue[0].TimeStamp),
	}
)

func createZipFile(t *testing.T) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	return buf.Bytes()
}

func TestDumpFiles(t *testing.T) {
	data := createZipFile(t)
	folder := filepath.Join(os.TempDir(), "pipeline_commands_test")
	err := os.MkdirAll(folder, 0755)
	visited := make(map[string]bool)
	for _, f := range files {
		visited[filepath.Clean(f.Name)] = false
	}

	defer func() {
		os.RemoveAll(folder)
	}()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	err = dumpZippedData(data, folder)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	filepath.Walk(folder, func(path string, inf os.FileInfo, err error) error {
		entry, err := filepath.Rel(folder, path)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		visited[entry] = true
		return nil
	})
	for _, f := range files {
		if !visited[filepath.Clean(f.Name)] {
			t.Errorf("%v was not visited", filepath.Clean(f.Name))
		}
	}

}

//Tests the command and checks that the output is correct
func TestQueueCommand(t *testing.T) {
	cli, link, _ := makeReturningCli(queue, t)
	r := overrideOutput(cli)
	AddQueueCommand(cli, link)
	err := cli.Run([]string{"queue"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != QUEUE_CALL {
		t.Errorf("Queue wasn't call")
	}

	if ok, line, message := checkTableLine(r, "\t", queueLine); !ok {
		t.Errorf("Queue template doesn't match (%q,%s)\n%s", queueLine, line, message)
	}
}

//Tests that the move up command links to the pipeline and checks the output format
func TestMoveUpCommand(t *testing.T) {
	cli, link, _ := makeReturningCli(queue, t)
	r := overrideOutput(cli)
	AddMoveUpCommand(cli, link)

	err := cli.Run([]string{"moveup", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != MOVEUP_CALL {
		t.Errorf("moveup wasn't called")
	}

	if ok, line, message := checkTableLine(r, "\t", queueLine); !ok {
		t.Errorf("Queue template doesn't match (%q,%s)\n%s", queueLine, line, message)
	}
}

//Tests that the move down command links to the pipeline and checks the output format
func TestMoveDownCommand(t *testing.T) {
	cli, link, _ := makeReturningCli(queue, t)
	r := overrideOutput(cli)
	AddMoveDownCommand(cli, link)

	err := cli.Run([]string{"movedown", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != MOVEDOWN_CALL {
		t.Errorf("moveup wasn't called")
	}

	if ok, line, message := checkTableLine(r, "\t", queueLine); !ok {
		t.Errorf("Queue template doesn't match (%q,%s)\n%s", queueLine, line, message)
	}
}

//Tests that the version command links to the pipeline and checks the output format
func TestVersionCommand(t *testing.T) {
	pipe := newPipelineTest(false)
	link := PipelineLink{pipeline: pipe}
	link.Version = "2.0.0-test"
	cli, err := makeCli("test", &link)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	r := overrideOutput(cli)
	AddVersionCommand(cli, link)

	err = cli.Run([]string{"version"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	values := checkMapLikeOutput(r)
	if val, ok := values["Client version"]; ok {
		//this is set by a constant, just check there is
		//a value
		if len(val) == 0 {
			t.Errorf("Client version is empty")
		}

	} else {
		t.Errorf("Client version not present")
	}

	if val, ok := values["Pipeline version"]; !ok || val != "2.0.0-test" {
		t.Errorf("Pipeline version '2.0.0-test'!='%s'", val)

	}

	if val, ok := values["Pipeline authentication"]; !ok || val != "false" {
		t.Errorf("Pipeline authentication'false'!=%s", val)

	}
}

func makeReturningCli(val interface{}, t *testing.T) (*Cli, PipelineLink, *PipelineTest) {
	pipe := newPipelineTest(false)
	pipe.val = val
	link := PipelineLink{pipeline: pipe}
	cli, err := makeCli("test", &link)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	return cli, link, pipe
}

//Checks the call to the pipeline link and the output
func TestLogCommand(t *testing.T) {
	expected := []byte("Oh my log!")
	cli, link, _ := makeReturningCli(expected, t)
	r := overrideOutput(cli)
	AddLogCommand(cli, link)
	err := cli.Run([]string{"log", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != LOG_CALL {
		t.Errorf("moveup wasn't called")
	}

	result := string(r.Bytes())
	if result != string(expected) {
		t.Errorf("Log error %s!=%s", string(expected), result)
	}
}

//Checks the call to the pipeline link and the output
func TestLogCommandWithOutputFile(t *testing.T) {
	expected := []byte("Oh my log!")
	cli, link, _ := makeReturningCli(expected, t)
	r := overrideOutput(cli)
	AddLogCommand(cli, link)
	file, err := ioutil.TempFile("", "cli_")
	defer os.Remove(file.Name())
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	err = cli.Run([]string{"log", "-o", file.Name(), "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != LOG_CALL {
		t.Errorf("moveup wasn't called")
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	result := string(contents)
	if result != string(expected) {
		t.Errorf("Log error %s!=%s", string(expected), result)
	}
	if r.Len() == 0 {
		t.Errorf("We haven't informed the user that we wrote the file somewhere")
	}
}

//Tests the log command when an error is returned by the
//link
func TestLogCommandError(t *testing.T) {
	cli, link, pipe := makeReturningCli(nil, t)
	pipe.failOnCall = LOG_CALL
	AddLogCommand(cli, link)
	err := cli.Run([]string{"log", "id"})
	if getCall(link) != LOG_CALL {
		t.Errorf("moveup wasn't called")
	}
	if err == nil {
		t.Errorf("Exepected error not returned")
	}
}

//Tests the log command when there is a writing error
func TestLogCommandWritingError(t *testing.T) {
	cli, link, _ := makeReturningCli(nil, t)
	cli.Output = FailingWriter{}
	AddLogCommand(cli, link)
	err := cli.Run([]string{"log", "id"})
	if getCall(link) != LOG_CALL {
		t.Errorf("moveup wasn't called")
	}
	if err == nil {
		t.Errorf("Exepected error not returned")
	}
}

//Checks the call to the pipeline link and the output with the status command
func TestJobStatusCommand(t *testing.T) {
	//as mocking logic is more complex for jobs
	//expected := JOB_1
	cli, link, _ := makeReturningCli(nil, t)
	r := overrideOutput(cli)
	AddJobStatusCommand(cli, link)
	err := cli.Run([]string{"status", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != JOB_CALL {
		t.Errorf("moveup wasn't called")
	}

	values := checkMapLikeOutput(r)
	if id, ok := values["Job Id"]; !ok || id != JOB_1.Id {
		t.Errorf("Id doesn't match %s!=%s", JOB_1.Id, id)
	}
	if status, ok := values["Status"]; !ok || status != JOB_1.Status {
		t.Errorf("Status doesn't match %s!=%s", JOB_1.Status, status)
	}
	if _, ok := values["Messages"]; ok {
		t.Errorf("I said to shut up! but I can see some msgs")
	}
}

//Checks that we get messages when using the verbose flag
func TestVerboseJobStatusCommand(t *testing.T) {
	//as mocking logic is more complex for jobs
	//expected := JOB_1
	cli, link, _ := makeReturningCli(nil, t)
	r := overrideOutput(cli)
	AddJobStatusCommand(cli, link)
	err := cli.Run([]string{"status", "-v", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != JOB_CALL {
		t.Errorf("moveup wasn't called")
	}
	exp := regexp.MustCompile("\\(\\d+\\)\\[\\w+\\]\\s+\\w+")
	matches := exp.FindAll(r.Bytes(), -1)
	if len(matches) != 2 {
		t.Errorf("The messages weren't printed output:\n%s", string(string(r.Bytes())))
	}

}

//Checks that the error is propagated when the link errors when calling status
func TestJobStatusCommandError(t *testing.T) {
	//as mocking logic is more complex for jobs
	//expected := JOB_1
	cli, link, p := makeReturningCli(nil, t)
	p.failOnCall = JOB_CALL
	overrideOutput(cli)
	AddJobStatusCommand(cli, link)
	err := cli.Run([]string{"status", "id"})
	if getCall(link) != JOB_CALL {
		t.Errorf("moveup wasn't called")
	}
	if err == nil {
		t.Errorf("Expected error not propagated")
	}

}
