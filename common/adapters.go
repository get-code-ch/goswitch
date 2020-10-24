package common

import (
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"
)

// GetMacAddress retrieve mac address of a network adapter,
// if adapter is not found return a fake address and an error message
func GetMacAddress(adapterName string) (string, error) {
	mac := ""

	// Getting list of network interfaces
	interfaces, err := net.Interfaces()
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].HardwareAddr.String() < interfaces[j].HardwareAddr.String()
	})

	// Try to find interface matching with adapterName
	if err == nil {
		for _, i := range interfaces {
			if strings.ToLower(i.Name) == strings.ToLower(adapterName) {
				mac = i.HardwareAddr.String()
				break
			}
		}
	}

	// If no device match with adapterName we get mac address of first active interface (except loopback)
	if mac == "" && err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp == net.FlagUp &&
				i.Flags&net.FlagBroadcast == net.FlagBroadcast &&
				i.Flags&net.FlagLoopback != net.FlagLoopback {
				mac = i.HardwareAddr.String()
				break
			}
		}
	}

	// If no device was found setting mac address with a random string
	// In DECnet memoriam
	if mac == "" {
		mac = fmt.Sprintf("AA:00:04:%s:%s:%s", RandomString(2), RandomString(2), RandomString(2))
	}

	return mac, err
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("0123456789ABCDEF")

	if length < 1 {
		length = 10
	}

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
