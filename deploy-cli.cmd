set GOARCH=mipsle
set GOOS=linux
REM go get -d -v -u
go build  -o ./release/goswitch/gscli ./services/cli/
robocopy ./config ./release/goswitch/config cli.json /s /e