#!/bin/bash

function buildPlatform {
        #Compiling  the environment --check so we don't have to do it each time 
        if [ ! -d $(go env GOROOT)/bin/${1}_${2} ]; then  
                echo "Setting environment for ${1}/${2}"
                pushd $(go env GOROOT)/src > /dev/null; GOOS=${1} GOARCH=${2} ./make.bash --no-clean  &> /dev/null
                popd > /dev/null
        fi
        echo "Building dp2 for ${1}/${2}"
        GOOS=${1} GOARCH=${2} go build -o bin/dp2_${1}_${2} github.com/daisy-consortium/pipeline-cli-go/dp2 
        echo "Building dp2admin for ${1}/${2}"
        GOOS=${1} GOARCH=${2} go build -o bin/dp2admin_${1}_${2} github.com/daisy-consortium/pipeline-cli-go/dp2admin 
}

OS="darwin linux windows"
PLATFORMS="386 amd64"
mkdir -p target/src/github.com/daisy-consortium/pipeline-cli-go
cp -r cli dp2 dp2admin target/src/github.com/daisy-consortium/pipeline-cli-go
pushd target > /dev/null
export GOPATH=$PWD
echo "Fetching deps..."
go get github.com/capitancambio/go-subcommand
go get github.com/kylelemons/go-gypsy/yaml
go get github.com/daisy-consortium/pipeline-clientlib-go
go get bitbucket.org/kardianos/osext
echo "Testing..."
go test github.com/daisy-consortium/pipeline-cli-go/cli
echo "Building ..."
echo "Buliding native dp2"
go install github.com/daisy-consortium/pipeline-cli-go/dp2 
echo "Buliding native dp2admin"
go install github.com/daisy-consortium/pipeline-cli-go/dp2admin 
cp ../dp2/config.yml bin
for sys in $OS; do 
        for plat in $PLATFORMS; do 
                buildPlatform $sys $plat; 
        done
done
popd > /dev/null



