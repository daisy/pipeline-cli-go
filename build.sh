#!/bin/bash

function buildPlatform {
        #Compiling  the environment --check so we don't have to do it each time 
        if [ ! -d $(go env GOROOT)/bin/${1}_${2} ]; then  
                echo "Setting environment for ${1}/${2}"
                pushd $(go env GOROOT)/src > /dev/null; GOOS=${1} GOARCH=${2} ./make.bash --no-clean  &> /dev/null
                popd > /dev/null
        fi
        echo "Building dp2 for ${1}/${2}"
        mkdir -p bin/${1}_${2}
        GOOS=${1} GOARCH=${2} go build -o bin/${1}_${2}/dp2 github.com/daisy/pipeline-cli-go/dp2 
        echo "Building dp2admin for ${1}/${2}"
        GOOS=${1} GOARCH=${2} go build -o bin/${1}_${2}/dp2admin github.com/daisy/pipeline-cli-go/dp2admin 
        cp ../dp2/config.yml bin/${1}_${2}
}

OS="darwin linux windows"
PLATFORMS="386 amd64"
mkdir -p target/src/github.com/daisy/pipeline-cli-go
cp -r cli dp2 dp2admin target/src/github.com/daisy/pipeline-cli-go
pushd target > /dev/null
export GOPATH=$PWD
echo "Fetching deps..."
go get github.com/capitancambio/go-subcommand
go get launchpad.net/goyaml 
go get github.com/daisy/pipeline-clientlib-go
go get bitbucket.org/kardianos/osext
echo "Testing..."
go test github.com/daisy/pipeline-cli-go/cli/...
echo "Building ..."
for sys in $OS; do 
        for plat in $PLATFORMS; do 
                buildPlatform $sys $plat; 
        done
done
popd > /dev/null



