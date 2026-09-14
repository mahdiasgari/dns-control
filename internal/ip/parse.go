package ip

import (
	"fmt"
	"net"
	"strings"
)

func Parse(value string) (net.IP, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, fmt.Errorf("IP address is empty")
	}

	result := net.ParseIP(value)

	if result == nil {
		return nil, fmt.Errorf("invalid IP address: %q", value)
	}

	return result, nil
}
