#!/bin/sh 

set -eu

protoc -I=./proto3 --go_out=golang/pkg --python_out=py/ proto3/tasks.proto 

# Compile the UI
(cd ui; rm -f index.html; elm make src/Main.elm)

GOLANG_UIPATH=golang/pkg/rnr/ui
mkdir -p $GOLANG_UIPATH
rm -rf $GOLANG_UIPATH/*
cp ui/index.html $GOLANG_UIPATH
