@echo off

set GOOS=windows
set GOARCH=386
go run cmd/aquestalk-server/main.go
