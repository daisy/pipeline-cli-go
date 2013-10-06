package cli

import (
	"bytes"
	"testing"
)

var (
	YAML = `
#WS CONFIGURATION
host: http://daisy.org
port: 9999
ws_path: ws
ws_timeup: 10
#DP2 launch config 
exec_line_nix: unix
exec_line_win: windows
local: true
# ROBOT CONF
client_key: clientid
client_secret: supersecret
#connection settings
timeout_seconds: 10
#debug
debug: true
starting: true
`
	T_STRING = "Wrong %v\nexpected: %v\nresult:%v\n"
	EXP      = map[string]interface{}{
		"url":           "http://localhost:8181/ws/",
		"host":          "http://daisy.org",
		"port":          9999,
		"ws_path":       "ws",
		"ws_timeup":     10,
		"unix":          "unix",
		"windows":       "windows",
		"client_key":    "clientid",
		"client_secret": "supersecret",
		"time_out":      10,
		"starting":      true,
		"debug":         true,
	}
)

func tCompareToExp(cnf Config, t *testing.T) {
	var res interface{}
	var test string
	test = "host"
	res = cnf[HOST]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "port"
	res = cnf[PORT]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "ws_path"
	res = cnf[PATH]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}
	test = "ws_timeup"
	res = cnf[WSTIMEUP]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "unix"
	res = cnf[EXECLINENIX]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "windows"
	res = cnf[EXECLINEWIN]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "client_key"
	res = cnf[CLIENTKEY]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "client_secret"
	res = cnf[CLIENTSECRET]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "time_out"
	res = cnf[TIMEOUT]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "debug"
	res = cnf[DEBUG]
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}
}
func TestConfigYaml(t *testing.T) {
	yalmStr := bytes.NewBufferString(YAML)
	cnf := copyConf()
	err := cnf.FromYaml(yalmStr)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	tCompareToExp(cnf, t)

}

func TestGetUrl(t *testing.T) {
	cnf := copyConf()
	test := "url"
	if cnf.Url() != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], cnf.Url())
	}
}
