set GOARCH=mipsle
set GOOS=linux
go build  -o ./release/goswitch-mipsle/gsdevice ./device/
set GOARCH=arm
set GOARM=5
set GOOS=linux
go build  -o ./release/goswitch-arm/gsdevice ./device/
