#!/bin/sh
protoc -I=./proto3 --go_out=golang/pkg --python_out=py/ proto3/tasks.proto 
