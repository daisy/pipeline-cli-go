Command-Line Interface for the DAISY Pipeline 2 (Golang)
=======================================================
[![Build Status](https://travis-ci.org/daisy-consortium/pipeline-cli-go.png?branch=master)](https://travis-ci.org/daisy-consortium/pipeline-cli-go)

How to build
------------

        go get github.com/capitancambio/go-subcommand
        go get github.com/daisy-consortium/pipeline-clientlib-go
        go build github.com/daisy-consortium/pipeline-cli-go	

Usage
-----

	Usage: dp2 command [options]
	
	Script commands:
	
	zedai-to-epub3			Transforms a ZedAI (DAISY 4 XML) document into an EPUB 3 publication.
	daisy202-to-epub3			Transforms a DAISY 2.02 publication into an EPUB3 publication.
	dtbook-to-zedai			Transforms DTBook XML into ZedAI XML.
	dtbook-to-epub3			Converts multiple dtbooks to epub3 format
	
	General commands:
	
	status				Shows the detailed status for a single job
	delete				Deletes a job
	result				Gets the zip file containing the job results
	halt				Stops the WS
	jobs				Shows the status for every job
	help				Shows this message or the command help 
	version				Shows version and exits
	
	To list the global options type:  	dp2 help -g
	To get help for a command type:  	dp2 help COMMAND

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
