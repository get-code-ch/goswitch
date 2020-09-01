module github.com/get-code-ch/goswitch

go 1.14

replace github.com/get-code-ch/mcp23008/v3 => D:/projects/mcp23008/v3

replace github.com/get-code-ch/ads1115 => D:/projects/ads1115

require (
	github.com/get-code-ch/ads1115 v0.0.0-00010101000000-000000000000
	github.com/get-code-ch/mcp23008/v3 v3.0.0
	github.com/gorilla/websocket v1.4.2
	periph.io/x/periph v3.6.4+incompatible
)
