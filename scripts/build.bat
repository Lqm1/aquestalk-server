@echo off

set GOOS=windows
set GOARCH=386
go build -o aquestalk-server.exe cmd/aquestalk-server/main.go
