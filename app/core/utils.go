package core

import (
	"fmt"
	"net"
)

func GetMacAddress() (*string, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		fmt.Printf("Error Occurred While trying to get network interfaces")
		return nil, err
	}

	var macAddress string

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.HardwareAddr != nil {
			macAddress = iface.HardwareAddr.String()
			break
		}

	}
	return &macAddress, nil
}
