package reqcli

import (
	"net"
	"time"
)

func NewTCPCli(addr, targetAddr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", targetAddr, timeout)
}
