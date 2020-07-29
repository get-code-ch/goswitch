set GOARCH=amd64
set GOOS=linux
REM go get -d -v -u
go build  -o ./release/goswitch-amd64/gscommand-center ./services/command-center
