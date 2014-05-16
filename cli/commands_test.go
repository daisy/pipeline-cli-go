package cli

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

var files = []struct {
	Name, Body string
}{
	{"readme.txt", "This archive contains some text files."},
	{"fold1/gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"fold1/fold2/todo.txt", "Get animal handling licence.\nWrite more examples."},
}

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

func TestBuilderCommandConfig(t *testing.T) {
	pipe := newPipelineTest(false)
	link := PipelineLink{pipeline: pipe}
	cmdBuilder := NewCommandBuilder("name", "desc")
	fn := func(...interface{}) (interface{}, error) { return link.Queue() }
	cmdBuilder.withCall(fn)
	cmdBuilder.withTemplate(QueueTemplate)
	if cmdBuilder.linkCall == nil {
		t.Errorf("Function is not set")
	}
	if cmdBuilder.template != QueueTemplate {
		t.Errorf("Template is not set")
	}

}

func TestBuilderCommand(t *testing.T) {
	pipe := newPipelineTest(false)
	link := PipelineLink{pipeline: pipe}
	link.Version = "1"

	cmdBuilder := NewCommandBuilder("version", "ver")
	fn := func(...interface{}) (interface{}, error) {
		return Version{&link, VERSION}, nil
	}
	cmdBuilder.withCall(fn)
	cmdBuilder.withTemplate(VersionTemplate)
	cli, err := makeCli("test", &link)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)
	cli.Output = w

	cmdBuilder.build(cli)
	err = cli.Run([]string{"version"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if w.Len() == 0 {
		t.Errorf("Template didn't execute")
	}
}
func TestBuilderCommandWithId(t *testing.T) {
	pipe := newPipelineTest(false)

	link := PipelineLink{pipeline: pipe}
	link.Version = "1"

	printable := &printableJob{
		Data:    pipeline.Job{},
		Verbose: false,
	}
	cmdBuilder := NewCommandBuilder("status", "gets the status")
	fn := func(args ...interface{}) (interface{}, error) {
		job, err := link.Job(args[0].(string))
		if err != nil {
			return nil, err
		}
		printable.Data = job
		return printable, nil
	}
	cmdBuilder.withCall(fn)
	cmdBuilder.withTemplate(JobStatusTemplate)
	cli, err := makeCli("test", &link)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)
	cli.Output = w

	cmdBuilder.buildWithId(cli)
	err = cli.Run([]string{"status", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if w.Len() == 0 {
		t.Errorf("Template didn't execute")
	}
}

func TestQueueTemplate(t *testing.T) {
	//Todo make a mechanism to mock return values from the link
	queue := []pipeline.QueueJob{
		pipeline.QueueJob{
			Id:               "job1",
			ClientPriority:   "high",
			JobPriority:      "low",
			ComputedPriority: 1.555555,
			RelativeTime:     0.577777,
			TimeStamp:        1400237879517,
		},
	}
	tmpl, err := template.New("queue").Parse(QueueTemplate)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)
	err = tmpl.Execute(w, queue)
	reader := bufio.NewScanner(w)
	reader.Scan() //discard the header line
	reader.Scan()
	reader.Text()
	line := reader.Text()
	vals := strings.Split(line, "\t")
	expected := []string{
		queue[0].Id,
		fmt.Sprintf("%.2f", queue[0].ComputedPriority),
		queue[0].JobPriority,
		queue[0].ClientPriority,
		fmt.Sprintf("%.2f", queue[0].RelativeTime),
		fmt.Sprintf("%d", queue[0].TimeStamp),
	}
	for idx, _ := range vals {
		if vals[idx] != expected[idx] {
			t.Errorf("Error in displayed data %v!=%v", vals[idx], expected[idx])
		}
	}
}
