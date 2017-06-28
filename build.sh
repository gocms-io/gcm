#!/bin/bash

# for testing. comment out for release.
#TRAVIS_BRANCH=

# grab govendor and sync
echo install govendor
go get -u github.com/kardianos/govendor

echo pulling in deps with govendor
govendor sync

function buildArch() {
    GOOS=$1 GOARCH=$2 go build -o bin/$TRAVIS_BRANCH/$3/$4
    pushd bin/$TRAVIS_BRANCH/$3
    zip -r gocms.zip *
    popd
}

# build linux64
buildArch linux amd64 linux_64 gcm
buildArch linux 386 linux_32 gcm
buildArch linux arm linux_arm gcm
buildArch darwin amd64 osx_64 gcm
buildArch windows amd64 windows_64 gcm.exe
buildArch windows 386 windows_32 gcm.exe
