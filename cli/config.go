package cli

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
	"io/ioutil"
	"log"
	"os"
)

//Yaml file keys
const (
	HOST         = "host"
	PORT         = "port"
	PATH         = "ws_path"
	WSTIMEUP     = "ws_timeup"
	EXECLINENIX  = "exec_line_nix"
	EXECLINEWIN  = "exec_line_win"
	CLIENTKEY    = "client_key"
	CLIENTSECRET = "client_secret"
	TIMEOUT      = "timeout_seconds"
	DEBUG        = "debug"
	STARTING     = "starting"
	ERR_STR      = "Error parsing configuration: %v"
	DEFAULT_FILE = "config.yml"
)

type Config map[string]interface{}

var config = Config{

	HOST:         "http://localhost",
	PORT:         8181,
	PATH:         "ws",
	WSTIMEUP:     25,
	EXECLINENIX:  "",
	EXECLINEWIN:  "",
	CLIENTKEY:    "",
	CLIENTSECRET: "",
	TIMEOUT:      10,
	DEBUG:        false,
	STARTING:     false,
}
var config_descriptions = map[string]string{

	HOST:         "Pipeline's webservice host",
	PORT:         "Pipeline's webserivce port",
	PATH:         "Pipeline's webservice path, as in http://daisy.org:8181/path",
	WSTIMEUP:     "Time to wait until the webserivce starts in seconds",
	EXECLINENIX:  "Pipeline webserivice executable path in unix-like systems",
	EXECLINEWIN:  "Pipeline webserivice executable path in windows systems",
	CLIENTKEY:    "Client key for authenticated requests",
	CLIENTSECRET: "Client secrect for authenticated requests",
	TIMEOUT:      "Http connection timeout in seconds",
	DEBUG:        "Print debug messages. true or false. ",
	STARTING:     "Start the webservice in the local computer if it is not running. true or false",
}

func copyConf() Config {
	ret := make(Config)
	for k, v := range config {
		ret[k] = v
	}
	return ret
}

func NewConfig() Config {
	cnf := copyConf()
	if err := loadDefault(cnf); err != nil {
		fmt.Println("Warning : no default configuration file found")
		return copyConf()
	}
	return cnf
}

func loadDefault(cnf Config) error {
	folder, err := osext.ExecutableFolder()
	if err != nil {
		return err
	}
	file, err := os.Open(folder + string(os.PathSeparator) + DEFAULT_FILE)
	if err != nil {
		return err
	}
	defer file.Close()
	err = cnf.FromYaml(file)
	if err != nil {
		return err
	}
	return nil
}

func (c Config) FromYaml(r io.Reader) error {
	node, err := yaml.Parse(r)
	if err != nil {
		return err
	}
	err = nodeToConfig(c, node)
	return err
}
func (c Config) UpdateDebug() {
	if !c[DEBUG].(bool) {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stdout)
	}
}

func (c Config) Url() string {
	return fmt.Sprintf("%v:%v/%v/", c[HOST], c[PORT], c[PATH])
}

func nodeToConfig(conf Config, node yaml.Node) error {
	var err error
	file := yaml.File{Root: node}
	conf[HOST], err = file.Get(HOST)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err := file.GetInt(PORT)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf[PORT] = int(aux)

	conf[PATH], err = file.Get(PATH)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err = file.GetInt(WSTIMEUP)
	conf[WSTIMEUP] = int(aux)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf[EXECLINENIX], err = file.Get(EXECLINENIX)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf[EXECLINEWIN], err = file.Get(EXECLINEWIN)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf[CLIENTKEY], err = file.Get(CLIENTKEY)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf[CLIENTSECRET], err = file.Get(CLIENTSECRET)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err = file.GetInt(TIMEOUT)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf[TIMEOUT] = int(aux)
	conf[DEBUG], err = file.GetBool(DEBUG)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf[STARTING], err = file.GetBool(STARTING)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf.UpdateDebug()
	return nil
}
