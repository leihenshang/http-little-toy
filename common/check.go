package common

import (
	"fmt"
	"net"
	"time"
)

// ConnectivityTest 检查地址的连接性
func ConnectivityTest(ipPorts string) (err error) {
	netRes, err := net.DialTimeout("tcp", ipPorts, time.Second*5)
	if err != nil {
		return err
	}
	if netRes == nil {
		return fmt.Errorf("the %s is disabled.", ipPorts)
	}

	_ = netRes.Close()
	return
}
