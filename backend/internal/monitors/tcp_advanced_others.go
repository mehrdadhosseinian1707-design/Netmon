//go:build !linux

package monitors

import (
	"net"
)

// getTCPInfo returns retransmission count and congestion window for non-Linux platforms.
// On Windows, these values are not available via standard syscalls and are returned as 0.
// On other platforms, 0 is also returned.
func getTCPInfo(_ net.Conn) (int, uint32) {
	return 0, 0
}
