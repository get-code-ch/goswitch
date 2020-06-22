set GOARCH=mipsle
set GOOS=linux
go get -d -v -u
go build services/device/ -o services/device/release/goswitch
robocopy ./config/device.json ./services/device/release/goswitch/config/ /s /e