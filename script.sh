#!/bin/bash

GOOS=windows GOARCH=amd64 go build -o cad-creation-automation.exe
GOOS=windows GOARCH=386 go build -o cad-creation-automation-32.exe
go build