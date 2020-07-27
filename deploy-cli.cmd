set GOARCH=mipsle
set GOOS=linux
REM go get -d -v -u
go build  -o ./release/goswitch-mipsle/gscli ./services/cli/
REM robocopy ./config ./release/goswitch/config cli.json /s /e
set GOARCH=arm
set GOARM=5
set GOOS=linux
go build  -o ./release/goswitch-arm/gscli ./services/cli/
