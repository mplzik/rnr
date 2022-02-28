#!/bin/sh 

set -eu

protoc -I=./proto3 --go_out=golang/pkg --python_out=py/ proto3/tasks.proto 

# Compile the UI
(cd elm_ui; rm -f index.html; elm make src/Main.elm)
cp elm_ui/index.html ui/

GOLANG_UIPATH=golang/pkg/rnr/ui
mkdir -p $GOLANG_UIPATH
rm -rf $GOLANG_UIPATH/*
cp ui/* $GOLANG_UIPATH
