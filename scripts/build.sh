#!/bin/bash

GOOS=windows GOARCH=386 go build -o aquestalk-server.exe cmd/aquestalk-server/main.go
