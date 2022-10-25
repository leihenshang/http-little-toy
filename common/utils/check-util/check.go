package check_util

import (
	"errors"
	"net"
	"time"
)

func ConnectivityTest(ipPorts string) (err error) {
	netRes, netErr := net.DialTimeout("tcp", ipPorts, time.Second*3)
	if netErr != nil {
		return netErr
	}
	if netRes == nil {
		return errors.New("the %s is disabled.")
	}

	if netRes != nil {
		netRes.Close()
		return nil
	}

	return
}
