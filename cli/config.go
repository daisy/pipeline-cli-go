package cli

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
)

//Yaml file keys
const (
	HOST         = "host"
	PORT         = "port"
	PATH         = "ws_path"
	WSTIMEUP     = "ws_timeup"
	EXECLINENIX  = "exec_line_nix"
	EXCLINEWIN   = "exec_line_win"
	CLIENTKEY    = "client_key"
	CLIENTSECRET = "client_secret"
	TIMEOUT      = "timeout_seconds"
	DEBUG        = "debug"
	STARTING     = "starting"
	ERR_STR      = "Error parsing configuration: %v"
)

var defaults = map[string]interface{}{

	HOST:         "http://localhost",
	PORT:         8181,
	PATH:         "ws",
	WSTIMEUP:     25,
	EXECLINENIX:  "",
	EXCLINEWIN:   "",
	CLIENTKEY:    "",
	CLIENTSECRET: "",
	TIMEOUT:      10,
	DEBUG:        false,
	STARTING:     false,
}

//Contains the items from the configuration
//file
type Config struct {
	Host         string //Framework host
	Port         int    //Framerwork port
	Path         string //Framework path
	WSTimeUp     int    //Time to wait till the framework comes up
	ExecLineNix  string //pipeline executable line for *nix systems
	ExecLineWin  string //pipeline executable line for windows systems
	Local        bool   //Local mode if we want to bring the pipeline up
	ClientKey    string //Client key for authorisation
	ClientSecret string //Client secret for authorisation
	TimeOut      int    //HTTP timeout
	Debug        bool   //Start in debug mode
	Starting     bool   //Starts the ws in case it is not running
}

func NewConfig() *Config {
	return &Config{

		Host:         defaults[HOST].(string),
		Port:         defaults[PORT].(int),
		Path:         defaults[PATH].(string),
		WSTimeUp:     defaults[WSTIMEUP].(int),
		ExecLineNix:  defaults[EXECLINENIX].(string),
		ExecLineWin:  defaults[EXCLINEWIN].(string),
		ClientKey:    defaults[CLIENTKEY].(string),
		ClientSecret: defaults[CLIENTSECRET].(string),
		TimeOut:      defaults[TIMEOUT].(int),
		Debug:        defaults[DEBUG].(bool),
		Starting:     defaults[STARTING].(bool),
	}
}
func (c *Config) FromYaml(r io.Reader) error {
	node, err := yaml.Parse(r)
	if err != nil {
		return err
	}
	err = nodeToConfig(node, c)
	return err
}
func (c Config) Url() string {
	return fmt.Sprintf("%v:%v/%v/", c.Host, c.Port, c.Path)
}
func nodeToConfig(node yaml.Node, conf *Config) error {
	var err error
	file := yaml.File{Root: node}
	conf.Host, err = file.Get(HOST)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err := file.GetInt(PORT)
	conf.Port = int(aux)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf.Path, err = file.Get(PATH)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err = file.GetInt(WSTIMEUP)
	conf.WSTimeUp = int(aux)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf.ExecLineNix, err = file.Get(EXECLINENIX)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf.ExecLineWin, err = file.Get(EXCLINEWIN)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf.ClientKey, err = file.Get(CLIENTKEY)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	conf.ClientSecret, err = file.Get(CLIENTSECRET)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}

	aux, err = file.GetInt(TIMEOUT)
	conf.TimeOut = int(aux)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf.Debug, err = file.GetBool(DEBUG)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	conf.Debug, err = file.GetBool(STARTING)
	if err != nil {
		return fmt.Errorf(ERR_STR, err)
	}
	return nil
}
