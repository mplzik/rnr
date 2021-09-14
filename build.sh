#!/bin/sh
protoc -I=./proto3 --go_out=golang/pkg --python_out=py/ proto3/tasks.proto 

GOLANG_UIPATH=golang/pkg/rnr/ui

mkdir -p $GOLANG_UIPATH
rm -rf $GOLANG_UIPATH/*
cp ui/* $GOLANG_UIPATH
