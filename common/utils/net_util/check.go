package net_util

import (
	"errors"
    "fmt"
    "net"
	"time"
)

func ConnectivityTest(ipPorts string) (err error) {
	netRes, netErr := net.DialTimeout("tcp", ipPorts, time.Second*3)
	if netErr != nil {
		return netErr
	}
	if netRes == nil {
        return errors.New(fmt.Sprintf("the %s is disabled.",ipPorts))
	}

	if netRes != nil {
		_ = netRes.Close()
		return nil
	}

	return
}
