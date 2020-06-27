set GOARCH=mipsle
set GOOS=linux
go get -d -v -u
go build  -o ./release/goswitch/gsdevice ./services/device/
robocopy ./config ./release/goswitch/config device.json /s /e