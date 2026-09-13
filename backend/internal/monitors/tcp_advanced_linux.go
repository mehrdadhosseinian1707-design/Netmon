//go:build linux

package monitors

import (
	"net"
	"syscall"
)

// getTCPInfo retrieves retransmission count and congestion window from the Linux kernel
// via getsockopt TCP_INFO. Returns (retransmissions, cwnd).
// retransmissions = TCPInfo.Retransmits + TCPInfo.Total_retrans
// cwnd = TCPInfo.Snd_cwnd
func getTCPInfo(conn net.Conn) (int, uint32) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, 0
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return 0, 0
	}

	var tcpInfo *syscall.TCPInfo
	var sysErr error

	ctrlErr := rawConn.Control(func(fd uintptr) {
		tcpInfo, sysErr = syscall.GetsockoptTCPInfo(int(fd), syscall.IPPROTO_TCP, syscall.TCP_INFO)
	})
	if ctrlErr != nil || sysErr != nil {
		return 0, 0
	}
	if tcpInfo == nil {
		return 0, 0
	}

	retrans := int(tcpInfo.Retransmits) + int(tcpInfo.Total_retrans)
	cwnd := tcpInfo.Snd_cwnd
	return retrans, cwnd
}
