package cli

import (
	"bytes"
	"testing"
)

var (
	YAML = `
#WS CONFIGURATION
host: http://localhost
port: 8181
ws_path: ws
ws_timeup: 25
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
`
	T_STRING = "Wrong %v\nexpected: %v\nresult:%v\n"
	EXP      = map[string]interface{}{
		"url":           "http://localhost:8181/ws/",
		"host":          "http://localhost",
		"port":          8181,
		"ws_path":          "ws",
		"ws_timeup":     25,
		"unix":          "unix",
		"windows":       "windows",
		"client_key":    "clientid",
		"client_secret": "supersecret",
		"time_out":       10,
		"debug":         true,
	}
)

func TestConfigCreation(t *testing.T) {
	yalmStr := bytes.NewBufferString(YAML)
	cnf, err := NewConfig(yalmStr)
	var res interface{}
	var test string

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	test = "host"
	res = cnf.Host
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "port"
	res = cnf.Port
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "ws_path"
	res = cnf.Path
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}
	test = "ws_timeup"
	res = cnf.WSTimeUp
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "unix"
	res = cnf.ExecLineNix
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "windows"
	res = cnf.ExecLineWin
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "client_key"
	res = cnf.ClientKey
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "client_secret"
	res = cnf.ClientSecret
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "time_out"
	res = cnf.TimeOut
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}

	test = "debug"
	res = cnf.Debug
	if res != EXP[test] {
		t.Errorf(T_STRING, test, EXP[test], res)
	}
}

func TestGetUrl(t *testing.T) {
        yalmStr := bytes.NewBufferString(YAML)
        cnf,err:=NewConfig(yalmStr)
        if err!=nil{
        t.Error("Unexpected error")
        }
        test:="url"
        if cnf.Url()!=EXP[test]{
        t.Errorf(T_STRING, test, EXP[test], cnf.Url())
        }
}
