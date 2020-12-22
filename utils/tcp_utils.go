package utils

import (
	"log"
	"net"
)

// RetryDial keeps trying to dial in intervals.
func RetryDial(protocol, address string, logger *log.Logger) net.Conn {
	logger.Println("[INFO] pairmint: Dialing Tendermint validator...")
	for {
		if conn, err := net.Dial(protocol, address); err == nil {
			logger.Printf("[INFO] pairmint: Established connection with %v. ✓\n", address)
			return conn
		}
	}
}
