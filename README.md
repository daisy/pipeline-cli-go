Command-Line Interface for the DAISY Pipeline 2 (Golang)
=======================================================
[![Build Status](https://travis-ci.org/daisy-consortium/pipeline-cli-go.png?branch=master)](https://travis-ci.org/daisy-consortium/pipeline-cli-go)

How to build
------------
1. Install golang from the [official site](http://golang.org/doc/install). If you'll be creating distributions of the cli please install from the [sources](http://golang.org/doc/install/source)
2. Create a go source directory:
        mkdir ~/src/golibs/
        cd ~/src/golibs/
        export GOPATH=~/src/golibs:$GOPATH
3. Install dependencies:
        go get github.com/capitancambio/go-subcommand
        go get github.com/kylelemons/go-gypsy/yaml
        go get github.com/daisy-consortium/pipeline-clientlib-go
        go get bitbucket.org/kardianos/osext
        go get github.com/daisy-consortium/pipeline-cli-go
4. The building process will create two executables, dp2 and dp2admin in the bin/ folder: 
        go install github.com/daisy-consortium/pipeline-cli-go/dp2
        go install github.com/daisy-consortium/pipeline-cli-go/dp2admin
5. Copy the default configuration file to the same directory as the binaries:
        cp src/github.com/daisy-consortium/pipeline-cli-go/dp2/config.yml bin/

How to build and distribute using maven
---------------------------------------
In order to allow the go client play nice with the rest of the pipeline ecosystem a maven build process is provided, although right now it only works on linux and mac systems ( You should be able to make it work using cygwin though).

Follow the previous instructions till step 2. installing go from the sources. 
        cd src/github.com/daisy-consortium/pipeline-cli-go/
        mvn clean install

You can find in the target/bin directory all the binaries from windows,mac and linux platforms.

Usage
-----


Configuration
-------------

Modify the settings in the file config.yml or alternatively use the global witches:

	--client_secret VALUE        Client secret  default(supersecret)
	--timeout_seconds VALUE      Connection timeout default(100)
	--authenticate VALUE         If true will send the authenticated url's to the ws default(false)
	--ws_timeup VALUE            Time in seconds to wait for the ws to start (if in local mode)  default(60)
	--local VALUE                CLI mode, true is local, false is remote (must be coherent with the ws instance)  default(true)
	--host VALUE                 Host name  default(http://localhost)
	--port VALUE                 Port number default(8181)
	--client_key VALUE           Client key default(clientkey)
	--exec_line VALUE            Path to the pipeline2 script  default(../bin/pipeline2)
	--debug VALUE                If true debug messages are printed on the terminal default(false)
	--ws_path VALUE              Path to the ws (as in http://host/ws_path)  default(ws)
