package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

const (
	QUEUE_CALL    = "queue"
	MOVEUP_CALL   = "moveup"
	MOVEDOWN_CALL = "movedown"
	LOG_CALL      = "log"
	JOB_CALL      = "job"
)

//Sets the output of the cli to a bytes.Buffer
func overrideOutput(cli *Cli) *bytes.Buffer {
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)
	cli.Output = w
	return w
}

//utility to check if the first line of the the table corresponds to the
//values passed by parameter
func checkTableLine(r io.Reader, separator string, values []string) (ok bool, line, message string) {
	reader := bufio.NewScanner(r)
	reader.Scan() //discard the header line
	reader.Scan()
	reader.Text()
	line = reader.Text()
	read := strings.Split(line, separator)
	if len(read) != len(values) {
		return false, line,
			fmt.Sprintf("Length is different got %d expected %d", len(read), len(values))
	}
	for idx, _ := range read {
		if read[idx] != values[idx] {
			return false, line,
				fmt.Sprintf("Different values  %s expected %s", read[idx], values[idx])
		}
	}
	return true, line, ""
}

//returns a map containing the key and values according to text lines separated by :, ignores lines that do not contain paired values
func checkMapLikeOutput(r io.Reader) map[string]string {
	reader := bufio.NewScanner(r)
	values := make(map[string]string)
	for reader.Scan() {
		pair := strings.Split(reader.Text(), ":")
		if len(pair) == 2 {
			values[strings.Trim(pair[0], " ")] = strings.Trim(pair[1], " ")
		} //else ignroe
	}
	return values
}

type FailingWriter struct {
}

func (f FailingWriter) Write([]byte) (int, error) {
	return 0, errors.New("writing error")
}

//Pipeline Mock
type PipelineTest struct {
	fail           bool
	count          int
	deleted        bool
	resulted       bool
	backgrounded   bool
	authentication bool
	fsallow        bool
	call           string
	val            interface{}
	failOnCall     string
}

func (p PipelineTest) mockCall() (val interface{}, err error) {

	if p.failOnCall == p.call {
		return val, errors.New("Error")
	}
	return p.val, nil
}

func (p PipelineTest) SetUrl(string) {
}
func (p *PipelineTest) SetVal(v interface{}) {
	p.val = v
}
func newPipelineTest(fail bool) *PipelineTest {
	return &PipelineTest{
		fail:           fail,
		count:          0,
		deleted:        false,
		resulted:       false,
		backgrounded:   false,
		authentication: false,
		fsallow:        true,
		call:           "",
	}
}

func getCall(l PipelineLink) string {
	return l.pipeline.(*PipelineTest).Call()
}
func (p PipelineTest) Call() string {
	return p.call

}

func (p PipelineTest) SetCredentials(key, secret string) {
}

func (p *PipelineTest) Alive() (alive pipeline.Alive, err error) {
	if p.fail {
		return alive, errors.New("Error")
	}
	alive.Version = "test"
	alive.FsAllow = p.fsallow
	alive.Authentication = p.authentication
	return
}

func (p *PipelineTest) Scripts() (scripts pipeline.Scripts, err error) {
	if p.fail {
		return scripts, errors.New("Error")
	}
	return pipeline.Scripts{Href: "test", Scripts: []pipeline.Script{pipeline.Script{Id: "test"}}}, err
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
	p.call = JOB_CALL
	_, err = p.mockCall()
	if err != nil {
		return
	}
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
	p.call = LOG_CALL
	ret, err := p.mockCall()
	if ret != nil {
		return ret.([]byte), err
	}
	return
}
func (p *PipelineTest) Jobs() (jobs pipeline.Jobs, err error) {
	return
}

func (p *PipelineTest) Halt(key string) error {
	return nil
}

func (p *PipelineTest) Clients() (c []pipeline.Client, err error) {
	return
}
func (p *PipelineTest) NewClient(cIn pipeline.Client) (cOut pipeline.Client, err error) {
	return
}
func (p *PipelineTest) DeleteClient(id string) (ok bool, err error) {
	return
}
func (p *PipelineTest) Client(id string) (client pipeline.Client, err error) {
	return
}
func (p *PipelineTest) ModifyClient(client pipeline.Client, id string) (c pipeline.Client, err error) {
	return
}
func (p *PipelineTest) Properties() (props []pipeline.Property, err error) {
	return
}
func (p *PipelineTest) Sizes() (sizes pipeline.JobSizes, err error) {
	return
}
func (p *PipelineTest) Queue() (val []pipeline.QueueJob, err error) {
	p.call = QUEUE_CALL
	ret, err := p.mockCall()
	if ret != nil {
		return ret.([]pipeline.QueueJob), err
	}
	return

}

func (p *PipelineTest) MoveUp(id string) (queue []pipeline.QueueJob, err error) {
	p.call = MOVEUP_CALL
	ret, err := p.mockCall()
	if ret != nil {
		return ret.([]pipeline.QueueJob), err
	}
	return
}
func (p *PipelineTest) MoveDown(id string) (queue []pipeline.QueueJob, err error) {
	p.call = MOVEDOWN_CALL
	ret, err := p.mockCall()
	if ret != nil {
		return ret.([]pipeline.QueueJob), err
	}
	return
}
