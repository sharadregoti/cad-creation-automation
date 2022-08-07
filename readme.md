# Building for windows

64 bit
GOOS=windows GOARCH=amd64 go build -o cad-creation-automation.exe

32 bit
GOOS=windows GOARCH=386 go build -o cad-creation-automation-32.exe

Linux
go build

# Running

The binary can only be run where the config & credential & token file is located
`./cad-creation-automation`