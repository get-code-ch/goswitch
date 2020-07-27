set GOARCH=mipsle
set GOOS=linux
REM go get -d -v -u
go build  -o ./release/goswitch-mipsle/gsdevice ./services/device/
REM robocopy ./config ./release/goswitch-mipsle/config device.json /s /e
set GOARCH=arm
set GOARM=5
set GOOS=linux
go build  -o ./release/goswitch-arm/gsdevice ./services/device/
